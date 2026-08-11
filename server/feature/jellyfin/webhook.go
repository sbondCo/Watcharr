// Jellyfin webhook service.
// Expects data in the format returned by the Jellyfin Webhook plugin:
// https://github.com/jellyfin/jellyfin-plugin-webhook

package jellyfin

import (
	"errors"
	"log/slog"
	"strconv"

	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
)

type WebhookAuthProvider interface {
	GetUserIDFromServiceClientId(name entity.UserServiceName, cid string) (uint, error)
}

type WebhookUserProvider interface {
	GetUserByID(id uint) domain.GetUserQueryBuilder
	GetUserServiceByName(userID uint, name entity.UserServiceName) (entity.UserServices, error)
}

type WebhookWatchedProvider interface {
	GetWatchedItemIDByTmdbID(userId uint, tmdbId uint, contentType entity.ContentType) (uint, error)
	AddWatched(userId uint, ar domain.WatchedAddRequest, extraProps domain.WatchedAddExtraProps) (entity.Watched, error)
	UpdateWatched(userId uint, id uint, ar domain.WatchedUpdateRequest, extra domain.WatchedUpdateRequestExtraProps) (domain.WatchedUpdateResponse, error)
}

type WebhookService struct {
	cfg             *config.ServerConfig
	jellyfinService *Service
	authProvider    WebhookAuthProvider
	userProvider    WebhookUserProvider
	watchedProvider WebhookWatchedProvider
	episodeProvider domain.AddWatchedEpisodeProvider
}

func NewWebhookService(
	cfg *config.ServerConfig,
	jellyfinService *Service,
	authProvider WebhookAuthProvider,
	userProvider WebhookUserProvider,
	watchedProvider WebhookWatchedProvider,
	episodeProvider domain.AddWatchedEpisodeProvider,
) *WebhookService {
	return &WebhookService{
		cfg:             cfg,
		jellyfinService: jellyfinService,
		authProvider:    authProvider,
		userProvider:    userProvider,
		watchedProvider: watchedProvider,
		episodeProvider: episodeProvider,
	}
}

// Entrypoint for webhook data.
func (w *WebhookService) Ingest(uuid string, data WebhookData) error {
	slog.Debug("Ingest: Starting.", "uuid", uuid, "data", data)

	if !w.validUUID(uuid) {
		slog.Error("Ingest: Invalid UUID caught.", "uuid_tried", uuid)
		return errors.New("dont be naughty")
	}

	// Process supported events.
	switch data.NotificationType {
	case WebhookTypePlaybackStart:
		return w.processPlaybackStart(&data)
	case WebhookTypePlaybackStop:
		return w.processPlaybackStop(&data)
	case WebhookTypeUserDataSaved:
		return w.processUserDataSaved(&data)
	default:
		return errors.New("unsupported notification type")
	}
}

// TODO Check user settings (Do they have auto jf sync enabled?,
// probs default that setting to true)

// TODO We should exit to avoid spamming exact same updates (user plays, stops, plays more, then finished)?
// ALTHOUGH how'd we ensure we aren't skipping valid events, like
// user playing it again later where we failed to save their finished event before?

// Process PlaybackStart event.
func (w *WebhookService) processPlaybackStart(data *WebhookData) error {
	return w.track(data)
}

// Process PlaybackStop event.
func (w *WebhookService) processPlaybackStop(data *WebhookData) error {
	return w.track(data)
}

// Process UserDataSaved event.
// We want this event because we still want to tell if the user manually marks
// an item in jellyfin as PLAYED without actually playing it through the player.
// BUT this event comes in continuously while the user is in playback, I guess
// because jellyfin is saving users playback progress, however we don't care
// about that, so to avoid wasting time, we will early exit the event if its not
// telling us an item was marked unplayed/played.
func (w *WebhookService) processUserDataSaved(data *WebhookData) error {
	slog.Debug("processUserDataSaved: Processing.",
		"save_reason", data.SaveReason)
	if data.SaveReason == "" {
		// If save reason is empty, show warning, we expect it to always have
		// one, if it doesn't it may need to be brought to admins attention.
		slog.Warn("processUserDataSaved: SaveReason is empty! Ignoring.")
		return nil
	}

	// Depending on SaveReason, do a thing.
	switch data.SaveReason {
	case WebhookUserDataSaveReasonTogglePlayed:
		return w.track(data)
	default:
		slog.Debug("processUserDataSaved: Ignoring this SaveReason.")
	}

	return nil
}

// Track jellyfin change to a watched entry.
func (w *WebhookService) track(data *WebhookData) error {
	user, userJFService, err := w.getUser(data.UserID)
	if err != nil {
		return err
	}

	tmdbID, contentType, err := w.getTopLevelTMDBID(data, user, userJFService)
	if err != nil {
		return err
	}

	err = w.applyStatusToWatched(data, &user, tmdbID, contentType)
	if err != nil {
		return err
	}

	return nil
}

// Link the webhook to a Watcharr user.
// Returns User, Users jellyfin service.
func (w *WebhookService) getUser(
	whUserID string,
) (entity.User, entity.UserServices, error) {
	userID, err := w.authProvider.GetUserIDFromServiceClientId(
		entity.UserServiceNameJellyfin, whUserID)
	if err != nil {
		slog.Error("getUser: Failed to get watcharr user_id from wh userID!",
			"error", err)
		return entity.User{}, entity.UserServices{}, err
	}
	slog.Debug("getUser: Linked a webhook to user.", "user_id", userID)

	user, err := w.userProvider.GetUserByID(userID).Done()
	if err != nil {
		slog.Error("getUser: Failed to get watcharr user!", "error", err)
		return user, entity.UserServices{}, err
	}
	slog.Debug("getUser: Got user.", "user", user)

	userJFService, err := w.userProvider.GetUserServiceByName(
		userID, entity.UserServiceNameJellyfin)
	if err != nil {
		slog.Error("getUser: Failed to get user service!", "error", err)
		return user, userJFService, err
	}
	return user, userJFService, nil
}

// Returns top level medias TMDB ID and its content type.
func (w *WebhookService) getTopLevelTMDBID(
	data *WebhookData,
	user entity.User,
	userJFService entity.UserServices,
) (uint, entity.ContentType, error) {
	var tmdbID string                      // Id of movie or tv series.
	var tmdbContentType entity.ContentType // Content type of `tmdbID` var above

	// We only need to look at Movie and Episode types, we dont get anything
	// for Series or Season types on the events we currently care about.
	switch data.ItemType {
	case WebhookItemTypeMovie:
		slog.Debug("getTopLevelTMDBID: Movie")
		tmdbID = data.ProviderTMDBID
		tmdbContentType = entity.MOVIE
	case WebhookItemTypeEpisode:
		slog.Debug("getTopLevelTMDBID: Episode")
		// The webhook data only includes ids for the episode, not the id for
		// the top-level series. We will query Jellyfin to get the data we need.
		jfItem, err := w.jellyfinService.GetSeriesByID(
			data.SeriesID, user.Username, userJFService.AuthToken)
		if err != nil {
			slog.Error("getTopLevelTMDBID: Failed to get series!", "error", err)
			return 0, "", err
		}
		tmdbID = jfItem.ProviderIds.Tmdb
		tmdbContentType = entity.SHOW
	default:
		slog.Error("getTopLevelTMDBID: Unsupported ItemType encountered",
			"item_type", data.ItemType)
		return 0, "", errors.New("unsupported ItemType")
	}

	uTMDBID, err := strconv.ParseUint(tmdbID, 10, 64)
	if err != nil {
		slog.Error("getTopLevelTMDBID: Parsing TMDBID failed!", "error", err)
		return 0, "", err
	}

	return uint(uTMDBID), tmdbContentType, nil
}

// For playback related events, we will add/update a watched entry to reflect
// jellyfin state change.
func (w *WebhookService) applyStatusToWatched(
	data *WebhookData,
	user *entity.User,
	tmdbID uint,
	contentType entity.ContentType,
) error {
	wID, err := w.watchedProvider.GetWatchedItemIDByTmdbID(
		user.ID, tmdbID, contentType)
	if err != nil {
		slog.Error("applyStatusToWatched: Linking to Watched entry failed!",
			"error", err)
		return err
	}

	// Decide new status.
	newStatus := w.decideNewWatchedStatus(data)
	if newStatus == "" {
		// Not an error, but this means we decided not to update status.
		slog.Debug("applyStatusToWatched: Not updating status. Stopping.")
		return nil
	}

	// Specifically status for top-level watched item.
	// Episode data can use `newStatus` directly, but updating Watched entries
	// should use this.
	newTopLevelStatus := newStatus
	if contentType == entity.SHOW {
		newTopLevelStatus = entity.WATCHING
	}

	// Add or update top-level watched entry
	if wID != 0 {
		// Found an existing watched entry, update it.
		slog.Debug("applyStatusToWatched: Found watched entry.", "wID", wID)
		_, err := w.watchedProvider.UpdateWatched(
			user.ID,
			wID,
			domain.WatchedUpdateRequest{
				Status: newTopLevelStatus,
			},
			domain.WatchedUpdateRequestExtraProps{
				ActivitySyncedBy: entity.ActivitySyncedByJellyfin,
			},
		)
		if err != nil {
			slog.Error("applyStatusToWatched: AddWatched failed!", "error", err)
			return err
		}
	} else {
		// Didn't find an existing watched entry, create one.
		slog.Debug("applyStatusToWatched: No watched entry, creating one.")
		w, err := w.watchedProvider.AddWatched(
			user.ID,
			domain.WatchedAddRequest{
				TMDBID:      int(tmdbID),
				ContentType: entity.ContentTypeToSupportedMedia(contentType),
				Status:      newTopLevelStatus,
			},
			domain.WatchedAddExtraProps{
				ActivityType:     entity.ADDED_WATCHED,
				ActivitySyncedBy: entity.ActivitySyncedByJellyfin,
			},
		)
		if err != nil {
			slog.Error("applyStatusToWatched: AddWatched failed!", "error", err)
			return err
		}
		// Set wID to new watched entry ID so we can use it later.
		wID = w.ID
	}

	// Set status for episode
	if contentType == entity.SHOW {
		err := w.applyStatusToSeriesEpisode(data, user, wID, newStatus)
		if err != nil {
			slog.Error("applyStatusToWatched: applyStatusToSeriesEpisode failed!",
				"error", err)
			return err
		}
	}

	return nil
}

// Apply status to series episode.
func (w *WebhookService) applyStatusToSeriesEpisode(
	data *WebhookData,
	user *entity.User,
	watchedID uint,
	newStatus entity.WatchedStatus,
) error {
	if data.SeasonNumber == nil || data.EpisodeNumber == 0 {
		return errors.New("no season and or episode number")
	}
	w.episodeProvider.AddWatchedEpisodes(
		user.ID,
		domain.WatchedEpisodeAddRequest{
			WatchedID:        watchedID,
			SeasonNumber:     *data.SeasonNumber,
			EpisodeNumber:    data.EpisodeNumber,
			Status:           newStatus,
			ActivitySyncedBy: entity.ActivitySyncedByJellyfin,
		},
	)
	return nil
}

// Decide on what the new status should be (for Movie or Episode).
// Returns newStatus, but can be empty if no change should occur, so check!
func (w *WebhookService) decideNewWatchedStatus(
	data *WebhookData,
) entity.WatchedStatus {
	var newStatus entity.WatchedStatus
	switch data.NotificationType {
	case WebhookTypePlaybackStart:
		newStatus = entity.WATCHING
	case WebhookTypePlaybackStop:
		// On PlaybackStop, we only want to update status if PlayedToCompletion
		// on the stop, otherwise the item wasn't finished on this play, so
		// we should ignore this.
		if data.PlayedToCompletion {
			newStatus = entity.FINISHED
		}
	case WebhookTypeUserDataSaved:
		if data.Played {
			newStatus = entity.FINISHED
		} else {
			newStatus = entity.PLANNED
		}
	}
	return newStatus
}

// Check UUID (aka the secret that is given to jellyfins webhook plugin)
// for the webhook is valid.
func (w *WebhookService) validUUID(uuid string) bool {
	// HACK, this should compare against a uuid in config
	if uuid == "turd" {
		return true
	}
	return false
}
