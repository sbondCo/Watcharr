package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"

	"gorm.io/gorm"
)

type EpisodeStatusChangedHookResponse struct {
	// The watched shows status if we modified it.
	NewShowStatus WatchedStatus `json:"newShowStatus,omitempty"`
	// The full watched season (if created or modified).
	WatchedSeason *WatchedSeason `json:"watchedSeason,omitempty"`
	// All activies we have added.
	AddedActivities []Activity `json:"addedActivities,omitempty"`
	// All errors (fatal and non-fatal) that were encountered.
	Errors []string `json:"errors,omitempty"`
}

// Called after an episode watched status has been set.
func hookEpisodeStatusChanged(db *gorm.DB, userId uint, watchedId uint, seasonNum int, episodeNum int, newEpisodeStatus WatchedStatus) EpisodeStatusChangedHookResponse {
	userSettings, err := userGetSettings(db, userId)
	if err != nil {
		slog.Error("hookEpisodeStatusChanged: Failed to get user settings! Hook will continue.", "error", err)
	} else {
		if !*userSettings.AutomateShowStatuses {
			slog.Debug("hookEpisodeStatusChanged: User has AutomateShowStatuses disabled. Skipping hook.", "user_id", userId)
			return EpisodeStatusChangedHookResponse{}
		}
	}

	// 1. Only continue if the episode was not marked dropped.
	if newEpisodeStatus == DROPPED {
		slog.Error("hookEpisodeStatusChanged: newEpisodeStatus is DROPPED, not continuing.")
		return EpisodeStatusChangedHookResponse{}
	}

	hookResponse := EpisodeStatusChangedHookResponse{}

	addHookActivity := func(aType ActivityType, data string) {
		addedActivity, _ := addActivity(db, userId, ActivityAddRequest{WatchedID: watchedId, Type: aType, Data: (data)})
		hookResponse.AddedActivities = append(hookResponse.AddedActivities, addedActivity)
	}

	// 2. If the season (this episode is in) has no status or is planned, set season to watching.
	watchedSeason, err := getWatchedSeason(db, userId, watchedId, seasonNum)
	if err != nil {
		slog.Error("hookEpisodeStatusChanged: Cannot continue, failed to get watchedSeason!", "error", err)
		return EpisodeStatusChangedHookResponse{Errors: []string{("failed to query db for watched season")}}
	}
	// If season not found, create it.
	if watchedSeason == nil {
		slog.Debug("hookEpisodeStatusChanged: Watched season does not exist. Creating now.")
		seasonStatus := newEpisodeStatus
		if newEpisodeStatus == FINISHED {
			seasonStatus = WATCHING
		}
		resp, err := addWatchedSeason(db, userId, WatchedSeasonAddRequest{
			addActivity:     SEASON_ADDED_AUTO,
			addActivityData: map[string]interface{}{"reason": fmt.Sprintf("Episode %d was set to %s while the season had no status.", episodeNum, newEpisodeStatus)},
			WatchedID:       watchedId,
			SeasonNumber:    seasonNum,
			Status:          seasonStatus,
		})
		if err != nil {
			slog.Error("hookEpisodeStatusChanged: Failed to add watched season!", "error", err)
			hookResponse.Errors = append(hookResponse.Errors, "failed to add watched season")
		} else {
			// addWatchedSeason returns all watched seasons, get the one just added. (may be best to retrofit addWatchedSeason later to return id of season/row created)
			justAddedWatchedSeason, err := getWatchedSeason(db, userId, watchedId, seasonNum)
			if err != nil {
				hookResponse.Errors = append(hookResponse.Errors, "failed to get newly added watched season for response")
			} else {
				watchedSeason = justAddedWatchedSeason
				hookResponse.WatchedSeason = watchedSeason
			}
			hookResponse.AddedActivities = append(hookResponse.AddedActivities, resp.AddedActivity)
		}
	} else if watchedSeason.Status == "" || watchedSeason.Status == PLANNED {
		watchedSeason.Status = WATCHING
		if res := db.Save(watchedSeason); res.Error != nil {
			slog.Error("hookEpisodeStatusChanged: Failed to update season status!", "error", res.Error)
			hookResponse.Errors = append(hookResponse.Errors, "failed to update season status")
		} else {
			hookResponse.WatchedSeason = watchedSeason
			json, _ := json.Marshal(map[string]interface{}{"season": seasonNum, "status": watchedSeason.Status, "reason": fmt.Sprintf("Episode %d was set to %s while the season had no status.", episodeNum, newEpisodeStatus)})
			addHookActivity(SEASON_STATUS_CHANGED_AUTO, string(json))
		}
	}

	// 3. If the show has no status or is planned, set it to watching.
	watchedShow, err := getWatchedItemById(db, userId, watchedId)
	if err != nil {
		slog.Error("hookEpisodeStatusChanged: Failed to get watched show, cant continue to update show status.", "error", err)
		hookResponse.Errors = append(hookResponse.Errors, "failed to get watched item for show")
		return hookResponse
	} else {
		// Show status shouldn't be empty, but watevs, handle it just incase
		if watchedShow.Status == "" || watchedShow.Status == PLANNED {
			watchedShow.Status = WATCHING
			if res := db.Save(watchedShow); res.Error != nil {
				slog.Error("hookEpisodeStatusChanged: Failed to update show status!", "error", res.Error)
			} else {
				hookResponse.NewShowStatus = watchedShow.Status
				json, _ := json.Marshal(map[string]interface{}{"status": watchedShow.Status, "reason": fmt.Sprintf("S%dE%d was set to %s.", seasonNum, episodeNum, newEpisodeStatus)})
				addHookActivity(STATUS_CHANGED_AUTO, string(json))
			}
		}
	}

	// 4. If all episodes are FINISHED or DROPPED, set the season to FINISHED
	// BUG If a seasons status is removed and the last episode of the season is marked finished,
	//     this will add activity for the season being marked finished, right after it is set
	//     to Watching just above. I think this might never happen to anyone so um ye.
	tmdbIdStr := strconv.Itoa(watchedShow.Content.TmdbID)
	seasonNumStr := strconv.Itoa(seasonNum)
	seasonDetails, err := seasonDetails(tmdbIdStr, seasonNumStr)
	if err != nil {
		slog.Error("hookEpisodeStatusChanged: Failed to get season details!", "error", err)
		hookResponse.Errors = append(hookResponse.Errors, "failed to get season details for show")
		return hookResponse
	}
	allEpisodesCount := len(seasonDetails.Episodes)
	finishedEpisodesCount, err := getNumberOfWatchedEpisodesInSeason(db, userId, watchedId, seasonNum, []WatchedStatus{FINISHED, DROPPED})
	if err != nil {
		slog.Error("hookEpisodeStatusChanged: Failed to get number of watched episodes in this season!", "error", err)
		hookResponse.Errors = append(hookResponse.Errors, "failed to get number of watched episodes in this season")
		return hookResponse
	}
	slog.Debug("hookEpisodeStatusChanged: Got episode counts.", "allEpisodesCount", allEpisodesCount, "finishedEpisodesCount", finishedEpisodesCount)
	if finishedEpisodesCount >= int64(allEpisodesCount) {
		slog.Debug("hookEpisodeStatusChanged: All episodes have been completed (finished or dropped). Marking season finished.")
		newStatus := FINISHED
		if watchedSeason != nil && watchedSeason.Status == newStatus {
			slog.Debug("hookEpisodeStatusChanged: WatchedSeason status is same as newStatus so not updating.")
			return hookResponse
		}
		if res := db.Model(&WatchedSeason{}).Where("watched_id = ? AND season_number = ? AND user_id = ?", watchedId, seasonNum, userId).Update("status", newStatus); res.Error != nil {
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
			addHookActivity(SEASON_STATUS_CHANGED_AUTO, string(json))
		}
	}

	return hookResponse
}
