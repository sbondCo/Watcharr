package episode

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/feature/watched/season"
	"github.com/sbondCo/Watcharr/media/tmdb"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WatchedEpisodeAddRequest struct {
	WatchedID       uint                 `json:"watchedId"`
	SeasonNumber    int                  `json:"seasonNumber"`
	EpisodeNumber   int                  `json:"episodeNumber"`
	Status          entity.WatchedStatus `json:"status"`
	Rating          int8                 `json:"rating" binding:"max=10"`
	AddActivity     entity.ActivityType  `json:"-"`
	AddActivityDate time.Time            `json:"-"`
}

type WatchedEpisodeAddResponse struct {
	WatchedEpisodes []entity.WatchedEpisode `json:"watchedEpisodes"`
	AddedActivity   entity.Activity         `json:"addedActivity"`
	// Response from hook
	EpisodeStatusChangedHookResponse EpisodeStatusChangedHookResponse `json:"episodeStatusChangedHookResponse,omitempty"`
}

type EpisodeStatusChangedHookResponse struct {
	// The watched shows status if we modified it.
	NewShowStatus entity.WatchedStatus `json:"newShowStatus,omitempty"`
	// The full watched season (if created or modified).
	WatchedSeason *entity.WatchedSeason `json:"watchedSeason,omitempty"`
	// All activies we have added.
	AddedActivities []entity.Activity `json:"addedActivities,omitempty"`
	// All errors (fatal and non-fatal) that were encountered.
	Errors []string `json:"errors,omitempty"`
}

type WatchedProvider interface {
	GetWatchedItemById(userId uint, id uint) (entity.Watched, error)
}

type WatchedSeasonProvider interface {
	GetWatchedSeason(userId uint, watchedId uint, seasonNumber int) (*entity.WatchedSeason, error)
	AddWatchedSeason(userId uint, ar season.WatchedSeasonAddRequest) (season.WatchedSeasonAddResponse, error)
}

type UserProvider interface {
	UserGetSettings(userId uint) (entity.UserSettings, error)
}

type Service struct {
	db               *gorm.DB
	wp               WatchedProvider
	wsp              WatchedSeasonProvider
	tmdb             *tmdb.TMDB
	activityProvider domain.ActivityAddProvider
	userProvider     UserProvider
}

func NewService(
	db *gorm.DB,
	wp WatchedProvider,
	wsp WatchedSeasonProvider,
	tmdb *tmdb.TMDB,
	activityProvider domain.ActivityAddProvider,
	userProvider UserProvider,
) *Service {
	return &Service{
		db,
		wp,
		wsp,
		tmdb,
		activityProvider,
		userProvider,
	}
}

// Add/edit a watched episode.
func (s *Service) AddWatchedEpisodes(userId uint, ar WatchedEpisodeAddRequest) (WatchedEpisodeAddResponse, error) {
	slog.Debug("Adding watched episode item", "userId", userId, "watchedID", ar.WatchedID, "season", ar.SeasonNumber, "episode", ar.EpisodeNumber)
	// 1. Make sure watched item exists and it is the correct type (TV)
	var w entity.Watched
	if resp := s.db.Where("id = ? AND user_id = ?", ar.WatchedID, userId).Preload("Content").Preload("WatchedEpisodes").Find(&w); resp.Error != nil {
		slog.Error("Failed when adding a watched episode", "error", "failed to get watched item from db")
		return WatchedEpisodeAddResponse{}, errors.New("failed when retrieving watched item")
	}
	if w.ID == 0 {
		slog.Error("Failed when adding a watched episode", "error", "watched item does not exist in db")
		return WatchedEpisodeAddResponse{}, errors.New("can't add a watched episode for a show that doesnt have a status itself")
	}
	if w.Content.Type != entity.SHOW {
		return WatchedEpisodeAddResponse{}, errors.New("can't add watched episode for non show content")
	}
	found := false
	updated := false
	for i, we := range w.WatchedEpisodes {
		if we.SeasonNumber == ar.SeasonNumber && we.EpisodeNumber == ar.EpisodeNumber {
			slog.Debug("Existing watched episode item found, updating existing")
			found = true
			if ar.Status != "" && ar.Status != w.WatchedEpisodes[i].Status {
				w.WatchedEpisodes[i].Status = ar.Status
				updated = true
			}
			if ar.Rating != 0 && ar.Rating != w.WatchedEpisodes[i].Rating {
				w.WatchedEpisodes[i].Rating = ar.Rating
				updated = true
			}
			break
		}
	}
	var addedActivity entity.Activity
	if !found {
		slog.Debug("Existing watched episode not found, adding as new entry")
		w.WatchedEpisodes = append(w.WatchedEpisodes, entity.WatchedEpisode{
			UserID:        userId,
			WatchedID:     ar.WatchedID,
			SeasonNumber:  ar.SeasonNumber,
			EpisodeNumber: ar.EpisodeNumber,
			Status:        ar.Status,
			Rating:        ar.Rating,
		})
	}
	if resp := s.db.Save(&w.WatchedEpisodes); resp.Error != nil {
		slog.Debug("Failed to save watched episode item in db", "error", resp.Error)
		return WatchedEpisodeAddResponse{}, errors.New("failed to save")
	}
	// Add activity
	if found {
		// Only add change activity if we actually updated a value
		// (changing value to same value doesn't count).
		if updated {
			if ar.Status != "" {
				json, _ := json.Marshal(map[string]any{
					"season":  ar.SeasonNumber,
					"episode": ar.EpisodeNumber,
					"status":  ar.Status})
				addedActivity, _ = s.activityProvider.AddActivity(
					userId,
					domain.ActivityAddProps{
						WatchedID: w.ID,
						Type:      entity.EPISODE_STATUS_CHANGED,
						Data:      string(json),
					},
					false,
				)
			}
			if ar.Rating != 0 {
				json, _ := json.Marshal(map[string]any{
					"season":  ar.SeasonNumber,
					"episode": ar.EpisodeNumber,
					"rating":  ar.Rating})
				addedActivity, _ = s.activityProvider.AddActivity(
					userId,
					domain.ActivityAddProps{
						WatchedID: w.ID,
						Type:      entity.EPISODE_RATING_CHANGED,
						Data:      string(json),
					},
					false,
				)
			}
		}
	} else {
		json, _ := json.Marshal(map[string]any{
			"season":  ar.SeasonNumber,
			"episode": ar.EpisodeNumber,
			"status":  ar.Status,
			"rating":  ar.Rating})
		act := domain.ActivityAddProps{
			WatchedID: w.ID,
			Type:      entity.EPISODE_ADDED,
			Data:      string(json),
		}
		if ar.AddActivity != "" {
			act.Type = ar.AddActivity
		}
		if !ar.AddActivityDate.IsZero() {
			act.CustomDate = &ar.AddActivityDate
		}
		addedActivity, _ = s.activityProvider.AddActivity(userId, act, false)
	}
	episodeAddResp := WatchedEpisodeAddResponse{
		WatchedEpisodes: w.WatchedEpisodes,
		AddedActivity:   addedActivity,
	}
	if ar.Status != "" {
		slog.Debug("addWatchedEpisodes: Episode status was changed, calling hook.")
		episodeAddResp.EpisodeStatusChangedHookResponse =
			s.hookEpisodeStatusChanged(
				userId,
				ar.WatchedID,
				ar.SeasonNumber,
				ar.EpisodeNumber,
				ar.Status)
	}
	return episodeAddResp, nil
}

// Remove a watched episode
func (s *Service) rmWatchedEpisode(userId uint, id uint) (entity.Activity, error) {
	slog.Debug("rmWatchedSeason called", "user_id", userId, "id", id)
	var watchedEpisode entity.WatchedEpisode
	resp := s.db.
		Clauses(clause.Returning{}).
		Model(&entity.WatchedEpisode{}).
		Unscoped().
		Where("id = ? AND user_id = ?", id, userId).
		Delete(&watchedEpisode)
	if resp.Error != nil {
		slog.Error("Failed when removing a watched episode", "error", resp.Error)
		return entity.Activity{}, errors.New("failed when removing watched episode")
	}
	if resp.RowsAffected == 0 {
		slog.Error("Failed when removing a watched episode", "error", "zero rows affected")
		return entity.Activity{}, errors.New("wasn't removed from db.. may not exist")
	}
	slog.Debug("rmWatchedEpisode, deleted row", "row", watchedEpisode)
	if watchedEpisode.ID != 0 {
		json, _ := json.Marshal(map[string]interface{}{
			"season":  watchedEpisode.SeasonNumber,
			"episode": watchedEpisode.EpisodeNumber,
			"status":  watchedEpisode.Status,
			"rating":  watchedEpisode.Rating,
		})
		addedActivity, _ := s.activityProvider.AddActivity(
			userId,
			domain.ActivityAddProps{
				WatchedID: watchedEpisode.WatchedID,
				Type:      entity.EPISODE_REMOVED,
				Data:      string(json),
			},
			false,
		)
		return addedActivity, nil
	}
	return entity.Activity{}, errors.New("removed, but failed to add activity entry")
}

func (s *Service) getNumberOfWatchedEpisodesInSeason(userId uint, watchedId uint, seasonNumber int, acceptableStatus []entity.WatchedStatus) (int64, error) {
	var count int64
	if res := s.db.Model(&entity.WatchedEpisode{}).Where("user_id = ? AND watched_id = ? AND season_number = ? AND status IN ?", userId, watchedId, seasonNumber, acceptableStatus).Count(&count); res.Error != nil {
		return 0, res.Error
	}
	return count, nil
}

// Called after an episode watched status has been set.
func (s *Service) hookEpisodeStatusChanged(userId uint, watchedId uint, seasonNum int, episodeNum int, newEpisodeStatus entity.WatchedStatus) EpisodeStatusChangedHookResponse {
	userSettings, err := s.userProvider.UserGetSettings(userId)
	if err != nil {
		slog.Error("hookEpisodeStatusChanged: Failed to get user settings! Hook will continue.", "error", err)
	} else {
		if !*userSettings.AutomateShowStatuses {
			slog.Debug("hookEpisodeStatusChanged: User has AutomateShowStatuses disabled. Skipping hook.", "user_id", userId)
			return EpisodeStatusChangedHookResponse{}
		}
	}

	hookResponse := EpisodeStatusChangedHookResponse{}

	addHookActivity := func(aType entity.ActivityType, data string) {
		addedActivity, _ := s.activityProvider.AddActivity(
			userId,
			domain.ActivityAddProps{
				WatchedID: watchedId,
				Type:      aType,
				Data:      (data),
			},
			false,
		)
		hookResponse.AddedActivities = append(hookResponse.AddedActivities, addedActivity)
	}

	// 2. If the season (this episode is in) has no status or is planned, set season to watching.
	watchedSeason, err := s.wsp.GetWatchedSeason(userId, watchedId, seasonNum)
	if err != nil {
		slog.Error("hookEpisodeStatusChanged: Cannot continue, failed to get watchedSeason!", "error", err)
		return EpisodeStatusChangedHookResponse{Errors: []string{("failed to query db for watched season")}}
	}
	// If season not found, create it.
	if watchedSeason == nil {
		slog.Debug("hookEpisodeStatusChanged: Watched season does not exist. Creating now.")
		seasonStatus := newEpisodeStatus
		if newEpisodeStatus == entity.FINISHED || newEpisodeStatus == entity.DROPPED {
			seasonStatus = entity.WATCHING
		}
		resp, err := s.wsp.AddWatchedSeason(userId, season.WatchedSeasonAddRequest{
			AddActivity:     entity.SEASON_ADDED_AUTO,
			AddActivityData: map[string]interface{}{"reason": fmt.Sprintf("Episode %d was set to %s while the season had no status.", episodeNum, newEpisodeStatus)},
			WatchedID:       watchedId,
			SeasonNumber:    seasonNum,
			Status:          seasonStatus,
		})
		if err != nil {
			slog.Error("hookEpisodeStatusChanged: Failed to add watched season!", "error", err)
			hookResponse.Errors = append(hookResponse.Errors, "failed to add watched season")
		} else {
			// addWatchedSeason returns all watched seasons, get the one just added. (may be best to retrofit addWatchedSeason later to return id of season/row created)
			justAddedWatchedSeason, err := s.wsp.GetWatchedSeason(userId, watchedId, seasonNum)
			if err != nil {
				hookResponse.Errors = append(hookResponse.Errors, "failed to get newly added watched season for response")
			} else {
				watchedSeason = justAddedWatchedSeason
				hookResponse.WatchedSeason = watchedSeason
			}
			hookResponse.AddedActivities = append(hookResponse.AddedActivities, resp.AddedActivity)
		}
	} else if watchedSeason.Status == "" || watchedSeason.Status == entity.PLANNED ||
		((newEpisodeStatus == entity.FINISHED || newEpisodeStatus == entity.WATCHING) && (watchedSeason.Status == entity.HOLD || watchedSeason.Status == entity.DROPPED)) {
		reasonStr := fmt.Sprintf("Episode %d was set to %s while the season had ", episodeNum, newEpisodeStatus)
		if watchedSeason.Status == "" {
			reasonStr += "no status."
		} else {
			reasonStr += fmt.Sprintf("a status of %s.", watchedSeason.Status)
		}
		watchedSeason.Status = entity.WATCHING
		if res := s.db.Save(watchedSeason); res.Error != nil {
			slog.Error("hookEpisodeStatusChanged: Failed to update season status!", "error", res.Error)
			hookResponse.Errors = append(hookResponse.Errors, "failed to update season status")
		} else {
			hookResponse.WatchedSeason = watchedSeason
			json, _ := json.Marshal(map[string]interface{}{"season": seasonNum, "status": watchedSeason.Status, "reason": reasonStr})
			addHookActivity(entity.SEASON_STATUS_CHANGED_AUTO, string(json))
		}
	}

	// 3. If the show has no status or is planned, set it to watching.
	watchedShow, err := s.wp.GetWatchedItemById(userId, watchedId)
	if err != nil {
		slog.Error("hookEpisodeStatusChanged: Failed to get watched show, cant continue to update show status.", "error", err)
		hookResponse.Errors = append(hookResponse.Errors, "failed to get watched item for show")
		return hookResponse
	} else {
		// Show status shouldn't be empty, but watevs, handle it just incase
		if watchedShow.Status == "" || watchedShow.Status == entity.PLANNED {
			watchedShow.Status = entity.WATCHING
			if res := s.db.Save(watchedShow); res.Error != nil {
				slog.Error("hookEpisodeStatusChanged: Failed to update show status!", "error", res.Error)
			} else {
				hookResponse.NewShowStatus = watchedShow.Status
				json, _ := json.Marshal(map[string]interface{}{"status": watchedShow.Status, "reason": fmt.Sprintf("S%dE%d was set to %s.", seasonNum, episodeNum, newEpisodeStatus)})
				addHookActivity(entity.STATUS_CHANGED_AUTO, string(json))
			}
		}
	}

	// 4. If all episodes are FINISHED or DROPPED, set the season to FINISHED
	// BUG If a seasons status is removed and the last episode of the season is marked finished,
	//     this will add activity for the season being marked finished, right after it is set
	//     to Watching just above. I think this might never happen to anyone so um ye.
	tmdbIdStr := strconv.Itoa(watchedShow.Content.TmdbID)
	seasonNumStr := strconv.Itoa(seasonNum)
	seasonDetails, err := s.tmdb.SeasonDetails(tmdbIdStr, seasonNumStr)
	if err != nil {
		slog.Error("hookEpisodeStatusChanged: Failed to get season details!", "error", err)
		hookResponse.Errors = append(hookResponse.Errors, "failed to get season details for show")
		return hookResponse
	}
	allEpisodesCount := len(seasonDetails.Episodes)
	finishedEpisodesCount, err := s.getNumberOfWatchedEpisodesInSeason(userId, watchedId, seasonNum, []entity.WatchedStatus{entity.FINISHED, entity.DROPPED})
	if err != nil {
		slog.Error("hookEpisodeStatusChanged: Failed to get number of watched episodes in this season!", "error", err)
		hookResponse.Errors = append(hookResponse.Errors, "failed to get number of watched episodes in this season")
		return hookResponse
	}
	slog.Debug("hookEpisodeStatusChanged: Got episode counts.", "allEpisodesCount", allEpisodesCount, "finishedEpisodesCount", finishedEpisodesCount)
	if finishedEpisodesCount >= int64(allEpisodesCount) {
		slog.Debug("hookEpisodeStatusChanged: All episodes have been completed (finished or dropped). Marking season finished.")
		newStatus := entity.FINISHED
		if watchedSeason != nil && watchedSeason.Status == newStatus {
			slog.Debug("hookEpisodeStatusChanged: WatchedSeason status is same as newStatus so not updating.")
			return hookResponse
		}
		if res := s.db.Model(&entity.WatchedSeason{}).Where("watched_id = ? AND season_number = ? AND user_id = ?", watchedId, seasonNum, userId).Update("status", newStatus); res.Error != nil {
			slog.Error("hookEpisodeStatusChanged: Failed to update season status to finished:", "error", res.Error.Error())
			hookResponse.Errors = append(hookResponse.Errors, "failed to update season status to finished")
			return hookResponse
		} else {
			if watchedSeason != nil {
				watchedSeason.Status = newStatus
				hookResponse.WatchedSeason = watchedSeason
			} else {
				slog.Error("hookEpisodeStatusChanged: watchedSeason was nil HOW DID THIS HAPPEN? Anyways the client won't be able to update its state with the new season status until it is refreshed.")
			}
			json, _ := json.Marshal(map[string]interface{}{"season": seasonNum, "status": newStatus, "reason": fmt.Sprintf("The season was deemed completed when episode %d was set to %s.", episodeNum, newEpisodeStatus)})
			addHookActivity(entity.SEASON_STATUS_CHANGED_AUTO, string(json))
		}
	}

	return hookResponse
}
