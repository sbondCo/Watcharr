package watched

import (
	"encoding/json"
	"errors"
	"log/slog"
	"strconv"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/feature/watched/addedtocontent"
	"github.com/sbondCo/Watcharr/util"
	"gorm.io/gorm"
)

type ContentProvider interface {
	GetOrCacheContent(contentType entity.ContentType, tmdbId int) (entity.Content, error)
}

type GameProvider interface {
	GetOrCache(igdbID int) (entity.Game, error)
}

type Service struct {
	db               *gorm.DB
	cp               ContentProvider
	gameProvider     GameProvider
	activityProvider domain.ActivityAddProvider
}

func NewService(
	db *gorm.DB,
	cp ContentProvider,
	gameProvider GameProvider,
	activityProvider domain.ActivityAddProvider,
) *Service {
	return &Service{
		db,
		cp,
		gameProvider,
		activityProvider,
	}
}

// Get entire watched list
func (s *Service) getWatched(userId uint) ([]entity.Watched, error) {
	watched := new([]entity.Watched)
	res := s.db.Model(&entity.Watched{}).
		Preload("Content").
		Preload("Game").
		Preload("Game.Poster").
		Preload("Activity").
		Preload("WatchedSeasons").
		Preload("WatchedEpisodes").
		Preload("Tags").
		Where("user_id = ?", userId).
		Find(&watched)
	if res.Error != nil {
		slog.Error("getWatched: Failed!", "error", res.Error)
		return []entity.Watched{}, res.Error
	}
	return *watched, nil
}

// Returns a page of users watched list.
func (s *Service) GetWatchedPage(
	userId uint,
	pp util.PaginationParams,
	wr domain.WatchedGetPageRequest,
	extraProps *domain.WatchedGetPageExtraProps,
) (util.PaginationResponse[entity.Watched, util.None], error) {
	slog.Debug("GetWatchedPage: A page was requested.",
		"user_id", userId,
		"pagination_params", pp,
		"wr", wr)
	watched := new([]entity.Watched)
	pRes := &util.PaginationResponse[entity.Watched, util.None]{}
	res := s.db.
		Debug().
		Model(&entity.Watched{}).
		Where(&entity.Watched{UserID: userId})

	if extraProps != nil {
		// Process `WatchedIds` extra prop.
		if len(extraProps.WatchedIds) > 0 {
			res = res.Where("`watcheds`.`id` IN ?", extraProps.WatchedIds)
		}

		// Search query
		if extraProps.Query != "" {
			q := "%" + extraProps.Query + "%"
			res = res.Where("Content.Title LIKE ? OR Game.Name LIKE ?", q, q)
		}
	}

	res = res.
		Joins("Content").
		Joins("Game").
		Preload("Game.Poster").
		Preload("Tags").
		Preload("WatchedSeasons").
		Preload("WatchedEpisodes").
		// Refine our results first (filters, sort);
		Scopes(
			watchedRefine(wr),
		).
		// Then count results (after filter);
		Count(&pRes.TotalResults).
		// Now calculate pagination properties with a TotalResults
		// that takes filtered out items into account.
		Scopes(
			util.Paginate(pp, pRes),
		).
		Find(&watched)
	if res.Error != nil {
		slog.Error("GetWatchedPage: Failed!", "error", res.Error)
		return util.PaginationResponse[entity.Watched, util.None]{}, res.Error
	}
	pRes.Results = *watched
	pRes.Finished(pp)
	return *pRes, nil
}

// Get a users **public** watchlist.
func (s *Service) getPublicWatched(
	userId uint,
	username string,
	pp util.PaginationParams,
	wr domain.WatchedGetPageRequest,
) (util.PaginationResponse[entity.Watched, util.None], error) {
	slog.Debug("getPublicWatched: running",
		"user_id", userId, "username", username)

	// First we need to make sure the users list is public
	user := new(entity.User)
	// I figure we require knowlege of the users id and name to make it
	// harder to just type in random ids and see someones list.. dunno
	// if this is a thing we need but its here.. for now at least.
	res := s.db.
		Where("id = ? AND username = ?", userId, username).
		Take(&user)
	if res.Error != nil {
		slog.Error("getPublicWatched: Failed to get user.",
			"user_id", userId)
		return util.PaginationResponse[entity.Watched, util.None]{}, errors.New("failed to check privacy settings")
	}
	if user.Private != nil && *user.Private {
		slog.Error("getPublicWatched: This users list is private.",
			"user_id", userId)
		return util.PaginationResponse[entity.Watched, util.None]{}, errors.New("this watched list is private")
	}

	// Now we know the user is public, return their list
	watched := new([]entity.Watched)
	pRes := &util.PaginationResponse[entity.Watched, util.None]{}
	res = s.db.
		Model(&entity.Watched{}).
		Where(&entity.Watched{UserID: userId}).
		Joins("Content").
		Joins("Game").
		Preload("Game.Poster").
		Preload("Tags").
		Preload("WatchedSeasons").
		Preload("WatchedEpisodes").
		// Refine our results first (filters, sort);
		Scopes(
			watchedRefine(wr),
		).
		// Then count results (after filter);
		Count(&pRes.TotalResults).
		// Now calculate pagination properties with a TotalResults
		// that takes filtered out items into account.
		Scopes(
			util.Paginate(pp, pRes),
		).
		Find(&watched)
	if res.Error != nil {
		slog.Error("getPublicWatched: Failed!", "error", res.Error)
		return util.PaginationResponse[entity.Watched, util.None]{}, errors.New("failed fetching the list")
	}
	pRes.Results = *watched
	pRes.Finished(pp)
	return *pRes, nil
}

// Get a watched list item by id (must be for `userId`).
func (s *Service) GetWatchedItemById(userId uint, id uint) (entity.Watched, error) {
	watched := new(entity.Watched)
	res := s.db.Model(&entity.Watched{}).Preload("Content").Where("user_id = ? AND id = ?", userId, id).Find(&watched)
	if res.Error != nil {
		slog.Error("GetWatchedItemById: Failed!", "error", res.Error)
		return entity.Watched{}, res.Error
	}
	return *watched, nil
}

// Get a watched list item by content (tmdb) id (must be for `userId`).
func (s *Service) GetWatchedItemByTmdbId(userId uint, tmdbId uint, contentType entity.ContentType) (entity.Watched, error) {
	slog.Debug("GetWatchedItemByTmdbId: Running.", "userId", userId, "tmdbId", tmdbId)
	watched := new(entity.Watched)
	res := s.db.Model(&entity.Watched{}).
		Joins("Content").
		Preload("Activity").
		Preload("WatchedSeasons").
		Preload("WatchedEpisodes").
		Preload("Tags").
		Where("user_id = ? AND Content.tmdb_id = ? AND Content.type = ?", userId, tmdbId, contentType).
		Take(&watched)
	if res.Error != nil {
		slog.Error("GetWatchedItemByTmdbId: Failed!", "error", res.Error)
		return entity.Watched{}, res.Error
	}
	slog.Debug("GetWatchedItemByTmdbId: Done.", "userId", userId, "tmdbId", tmdbId, "watched_item", watched)
	return *watched, nil
}

// Same as `getWatchedItemByTmdbId` except for getting in bulk (multiple content ids).
// `c` entries should be in format: [tmdb_id, ContentType] (Note: Couldn't figure out
// if it's possible to type this to enforce [int, ContentType] type for entries)
func (s *Service) GetWatchedItemsByTmdbIds(userId uint, c [][]any) ([]entity.Watched, error) {
	slog.Debug("GetWatchedItemsByTmdbIds: Running.", "userId", userId, "c", c)
	watched := new([]entity.Watched)
	res := s.db.Model(&entity.Watched{}).
		Joins("Content").
		Preload("Activity").
		Preload("WatchedSeasons").
		Preload("WatchedEpisodes").
		Preload("Tags").
		Where("user_id = ?", userId).
		Where("(Content.tmdb_id, Content.type) IN ?", c).
		Find(&watched)
	if res.Error != nil {
		slog.Error("GetWatchedItemsByTmdbIds: Failed!", "error", res.Error)
		return []entity.Watched{}, res.Error
	}
	slog.Debug(
		"GetWatchedItemsByTmdbIds: Done.",
		"userId", userId,
		"watcheds_found", len(*watched),
		// "wdev", *watched,
	)
	return *watched, nil
}

// Get a watched list item by game (igdb) id (must be for `userId`).
func (s *Service) GetWatchedItemByIgdbId(userId uint, igdbId uint) (entity.Watched, error) {
	slog.Debug("getWatchedItemByIgdbId: Running.", "userId", userId, "igdbId", igdbId)
	watched := new(entity.Watched)
	res := s.db.Model(&entity.Watched{}).
		Joins("Game").
		Preload("Game.Poster").
		Preload("Activity").
		Preload("Tags").
		Where("user_id = ? AND Game.igdb_id = ?", userId, igdbId).
		Take(&watched)
	if res.Error != nil {
		slog.Error("getWatchedItemByIgdbId: Failed!", "error", res.Error)
		return entity.Watched{}, res.Error
	}
	slog.Debug("getWatchedItemByIgdbId: Done.", "userId", userId, "igdbId", igdbId, "watched_item", watched)
	return *watched, nil
}

// Same as `getWatchedItemByIgdbId` except for getting in bulk (multiple content ids).
// `c` should be a slice of igdb ids.
func (s *Service) GetWatchedItemsByIgdbIds(userId uint, c []int) ([]entity.Watched, error) {
	slog.Debug("getWatchedItemsByIgdbIds: Running.", "userId", userId, "c", c)
	watched := new([]entity.Watched)
	res := s.db.Model(&entity.Watched{}).
		Joins("Game").
		Preload("Game.Poster").
		Preload("Activity").
		Preload("Tags").
		Where("user_id = ?", userId).
		Where("(Game.igdb_id) IN ?", c).
		Find(&watched)
	if res.Error != nil {
		slog.Error("getWatchedItemsByIgdbIds: Failed!", "error", res.Error)
		return []entity.Watched{}, res.Error
	}
	slog.Debug(
		"getWatchedItemsByIgdbIds: Done.",
		"userId", userId,
		"watcheds_found", len(*watched),
		// "wdev", *watched,
	)
	return *watched, nil
}

// Get watched item by an id and SupportedMedia type.
func (s *Service) GetWatchedItemBySupportedMediaId(userId uint, id uint, t util.SupportedMedia) (entity.Watched, error) {
	switch t {
	case util.SupportedMediaGame:
		return s.GetWatchedItemByIgdbId(userId, id)
	case util.SupportedMediaMovie:
		return s.GetWatchedItemByTmdbId(userId, id, entity.MOVIE)
	case util.SupportedMediaShow:
		return s.GetWatchedItemByTmdbId(userId, id, entity.SHOW)
	}
	slog.Error("GetWatchedItemBySupportedMediaId: Unsupported supportedmedia type",
		"type", t)
	return entity.Watched{}, errors.New("unsupported supportedmedia type")
}

// Get a list of watched items by a slice of ids and SupportedMedia types.
func (s *Service) GetWatchedItemsBySupportedMediaIds(userId uint, c []addedtocontent.IdToTypePair) ([]entity.Watched, error) {
	slog.Debug("GetWatchedItemsBySupportedMediaIds: Running.", "userId", userId, "c", c)
	// First we want to separate `c` into slices we can pass to the respective functions.
	tmdbIds := [][]any{}
	igdbIds := []int{}
	for _, v := range c {
		switch v.Type {
		case util.SupportedMediaMovie:
			tmdbIds = append(tmdbIds, []any{v.Id, entity.MOVIE})
		case util.SupportedMediaShow:
			tmdbIds = append(tmdbIds, []any{v.Id, entity.SHOW})
		case util.SupportedMediaGame:
			igdbIds = append(igdbIds, v.Id)
		}
	}
	// Now call each function relating to an overarching type.
	watcheds := []entity.Watched{}
	if len(tmdbIds) > 0 {
		if w, err := s.GetWatchedItemsByTmdbIds(userId, tmdbIds); err == nil {
			watcheds = append(watcheds, w...)
		} else {
			slog.Error("GetWatchedItemsBySupportedMediaIds: Failed to get items by tmdb ids.", "error", err)
			return []entity.Watched{}, err
		}
	}
	if len(igdbIds) > 0 {
		if w, err := s.GetWatchedItemsByIgdbIds(userId, igdbIds); err == nil {
			watcheds = append(watcheds, w...)
		} else {
			slog.Error("GetWatchedItemsBySupportedMediaIds: Failed to get items by igdb ids.", "error", err)
			return []entity.Watched{}, err
		}
	}
	return watcheds, nil
}

func (s *Service) AddWatched(
	userId uint,
	ar domain.WatchedAddRequest,
	at entity.ActivityType,
) (entity.Watched, error) {
	slog.Debug("Adding watched item",
		"user_id", userId,
		"add_request", ar)

	watched := entity.Watched{
		UserID: userId,
	}

	// Based on the ContentType given in the request, find the correct
	// content to attach to the Watched entry we are creating.
	switch ar.ContentType {
	case "movie", "tv":
		tmdbId := ar.TMDBID
		if tmdbId == 0 && ar.Deprecated_ContentID != 0 {
			// This if provides support for a now deprecated field.
			// When the property is removed, this if will be gone and we can
			// use `ar.TMDBID` directly for below.
			tmdbId = ar.Deprecated_ContentID
			slog.Error("AddWatched: You are using a deprecated field 'contentId', please replace it with 'tmdbId' to keep the same functionality for future releases of Watcharr.")
		}
		if tmdbId == 0 {
			return entity.Watched{}, errors.New("missing tmdb id")
		}
		// Get content cache (or cache it if we don't have it locally)
		content, err := s.cp.GetOrCacheContent(
			entity.ContentType(ar.ContentType), tmdbId)
		if err != nil {
			return entity.Watched{}, err
		}
		// Error if content has no id
		if content.ID == 0 {
			return entity.Watched{}, errors.New("failed to find content id")
		}
		// Add content to watched entry
		watched.ContentID = &content.ID
	case "game":
		if ar.IGDBID == 0 {
			return entity.Watched{}, errors.New("missing igdb id")
		}
		game, err := s.gameProvider.GetOrCache(ar.IGDBID)
		if err != nil {
			return entity.Watched{}, err
		}
		// Error if content has no id
		if game.ID == 0 {
			return entity.Watched{}, errors.New("failed to find game by id")
		}
		watched.GameID = &game.ID
	default:
		return entity.Watched{}, errors.New("invalid content type provided")
	}

	// Set default status for when content is added by
	// rating it instead of giving status first.
	if ar.Status == "" {
		if ar.ContentType == "movie" || ar.ContentType == "game" {
			ar.Status = entity.FINISHED
		} else {
			ar.Status = entity.WATCHING
		}
	}

	// Add watched data to struct
	watched.Status = ar.Status
	watched.Rating = ar.Rating
	watched.Thoughts = ar.Thoughts

	// If custom WatchedDate passed, set CreatedAt and UpdatedAt fields to it.
	if !ar.WatchedDate.IsZero() {
		slog.Debug("Adding watched item: The provided WatchedDate is valid.",
			"userId", userId,
			"request", ar)
		watched.CreatedAt = ar.WatchedDate
		watched.UpdatedAt = ar.WatchedDate
	}

	// Create the entry in db.
	res := s.db.Create(&watched)
	if res.Error != nil {
		if res.Error == gorm.ErrDuplicatedKey {
			// Try to restore the entry if unique contraint hit.
			if err := s.restoreWatched(userId, ar, &watched); err != nil {
				slog.Error("AddWatched: Failed to restore existing watched entry.")
				return entity.Watched{}, errors.New("failed when restoring existing entry")
			}
		} else {
			slog.Error("AddWatched: Error adding watched to database", "error", res.Error.Error())
			return entity.Watched{}, errors.New("failed adding watched entry to database")
		}
	}
	slog.Debug("AddWatched: Added watched list item", "item", watched)

	// Finally add activity
	activityAddReq := domain.ActivityAddRequest{
		WatchedID: watched.ID,
		Type:      at,
	}
	if activityJson, err := json.Marshal(map[string]any{
		"status": ar.Status,
		"rating": ar.Rating,
	}); err != nil {
		slog.Error("AddWatched: Failed to marshal json for add activity request, adding without data",
			"error", err)
	} else {
		activityAddReq.Data = string(activityJson)
	}
	act, _ := s.activityProvider.AddActivity(
		userId,
		activityAddReq,
	)
	watched.Activity = append(watched.Activity, act)

	return watched, nil
}

// Restore a watched entry that was soft deleted.
// Currently used for AddWatched, when it realizes the entry may exist already
// as a soft deleted record.
func (s *Service) restoreWatched(
	userId uint,
	ar domain.WatchedAddRequest,
	watchedOut *entity.Watched,
) error {
	slog.Info("restoreWatched: Attempting to restore.",
		"user_id", userId)

	// Our base where statement to find the row we want to restore by unique
	// indexes (user_id AND (content_id or game_id)).
	whereStmt := entity.Watched{
		UserID: userId,
	}

	// A check to ensure safety of the following db query:
	// We MUST have a UserID
	if whereStmt.UserID == 0 {
		return errors.New("no userid in provided watched struct")
	}
	// ...AND a (contentId || gameId).
	if watchedOut.ContentID != nil && *watchedOut.ContentID != 0 {
		whereStmt.ContentID = watchedOut.ContentID
	} else if watchedOut.GameID != nil && *watchedOut.GameID != 0 {
		whereStmt.GameID = watchedOut.GameID
	} else {
		return errors.New("no supported media ids in provided watched struct")
	}

	// Try to restore and update the possibly existing row.
	res := s.db.Debug().Model(&entity.Watched{}).
		Unscoped().
		Where(&whereStmt).
		Where("deleted_at IS NOT NULL").
		Updates(map[string]any{
			"status":     ar.Status,
			"rating":     ar.Rating,
			"thoughts":   ar.Thoughts,
			"deleted_at": nil,
		})
	if res.Error != nil {
		slog.Error("restoreWatched: Checking for record failed!",
			"error", res.Error)
		return errors.New("errored checking for soft deleted record")
	}
	if res.RowsAffected == 0 {
		slog.Error("restoreWatched: Nothing was updated. The row may already exist un-deleted.")
		return errors.New("didnt find an entry to restore")
	}
	slog.Info("restoreWatched: Restored record.",
		"user_id", userId)

	// Restore query above succeeded so now lets get all data needed and return.
	res = s.db.Debug().Model(&entity.Watched{}).
		Unscoped().
		Preload("Activity").
		Where(&whereStmt).
		Take(&watchedOut)
	if res.Error != nil {
		slog.Error("restoreWatched: Getting updated record failed!",
			"error", res.Error)
		return errors.New("errored while trying to get updated record")
	}

	return nil
}

// this method is too ugly to look at please make him look better, future irhm
func (s *Service) updateWatched(
	userId uint,
	id uint,
	ar domain.WatchedUpdateRequest,
) (domain.WatchedUpdateResponse, error) {
	slog.Debug("UpdateWatched", "request_data", ar)
	upwat := entity.Watched{}
	res := s.db.Model(&entity.Watched{}).Where("id = ? AND user_id = ?", id, userId).Take(&upwat)
	if res.Error != nil {
		slog.Error("Watched entry update failed:", "id", id, "error", res.Error.Error())
		return domain.WatchedUpdateResponse{}, errors.New("failed to update watched entry")
	}
	originalThoughts := upwat.Thoughts
	if ar.Rating != 0 {
		upwat.Rating = ar.Rating
	}
	if ar.Status != "" {
		upwat.Status = ar.Status
	}
	if ar.Thoughts != "" {
		upwat.Thoughts = ar.Thoughts
	}
	if ar.RemoveThoughts {
		upwat.Thoughts = ""
	}
	if ar.Pinned != nil {
		upwat.Pinned = *ar.Pinned
	}
	res = s.db.Save(upwat)
	if res.RowsAffected <= 0 {
		return domain.WatchedUpdateResponse{}, errors.New("no watched entry found")
	}
	addedActivity := entity.Activity{}
	if ar.Rating != 0 {
		addedActivity, _ = s.activityProvider.AddActivity(userId, domain.ActivityAddRequest{WatchedID: id, Type: entity.RATING_CHANGED, Data: strconv.Itoa(int(ar.Rating))})
	}
	if ar.Status != "" {
		addedActivity, _ = s.activityProvider.AddActivity(userId, domain.ActivityAddRequest{WatchedID: id, Type: entity.STATUS_CHANGED, Data: string(ar.Status)})
	}
	if ar.Thoughts != "" {
		addedActivity, _ = s.activityProvider.AddActivity(userId, domain.ActivityAddRequest{WatchedID: id, Type: entity.THOUGHTS_CHANGED})
	}
	if ar.RemoveThoughts {
		addedActivity, _ = s.activityProvider.AddActivity(userId, domain.ActivityAddRequest{WatchedID: id, Type: entity.THOUGHTS_REMOVED, Data: originalThoughts})
	}
	return domain.WatchedUpdateResponse{NewActivity: addedActivity}, nil
}

func (s *Service) UpdateWatchedLastViewedSeason(
	userId uint,
	id uint,
	seasonNum int,
) error {
	slog.Debug("UpdateWatchedLastViewedSeason",
		"user_id", userId,
		"id", id,
		"season_num", seasonNum)
	res := s.db.
		Model(&entity.Watched{}).
		Where("id = ? AND user_id = ?", id, userId).
		Update("last_viewed_season", seasonNum)
	if res.Error != nil {
		slog.Error("updateWatchedLastViewedSeason: Failed when updating.",
			"error", res.Error)
		return errors.New("failed to update db")
	}
	if res.RowsAffected == 0 {
		// likely the watched entry does not exist or is not owned by this `userId`.
		slog.Error("updateWatchedLastViewedSeason: Watched entry does not exist.")
		return errors.New("watched entry does not exist")
	}
	return nil
}

func (s *Service) removeWatched(
	userId uint,
	id uint,
) (domain.WatchedRemoveResponse, error) {
	slog.Debug("Removing watched item:", "id", id, "user_id", userId)
	// Our model has a deleted_at field, which will make gorm do a soft delete.
	// Since other tables (eg activities) will link their rows to a watched_id, it's best to soft
	// delete, so if user restores watched item they still have activity for example (also so
	// someone else wont get other users activity if auto increment gives them the same watched id).
	res := s.db.
		Model(&entity.Watched{}).
		Where("id = ? AND user_id = ?", id, userId).
		Delete(&entity.Watched{})
	if res.Error != nil {
		slog.Error("Removing watched entry failed",
			"id", id,
			"error", res.Error.Error())
		return domain.WatchedRemoveResponse{}, errors.New("failed to remove watched entry")
	}
	if res.RowsAffected <= 0 {
		return domain.WatchedRemoveResponse{}, errors.New("no watched entry found")
	}
	addedActivity, _ := s.activityProvider.AddActivity(userId, domain.ActivityAddRequest{WatchedID: id, Type: entity.REMOVED_WATCHED})
	return domain.WatchedRemoveResponse{NewActivity: addedActivity}, nil
}
