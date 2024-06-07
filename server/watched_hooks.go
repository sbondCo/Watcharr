package main

import (
	"log/slog"
	"strconv"

	"gorm.io/gorm"
)

type EpisodeStatusChangedHookResponse struct {
}

// Called after an episode watched status has been set.
func hookEpisodeStatusChanged(db *gorm.DB, userId uint, watchedId uint, seasonNum int, episodeNum int, newEpisodeStatus WatchedStatus) {
	// 1. Only continue if the episode was not marked dropped.
	if newEpisodeStatus == DROPPED {
		slog.Error("hookEpisodeStatusChanged: newEpisodeStatus is DROPPED, not continuing.")
		return
	}

	// 2. If the season (this episode is in) has no status or is planned, set season to watching.
	watchedSeason, err := getWatchedSeason(db, userId, watchedId, seasonNum)
	if err != nil {
		slog.Error("hookEpisodeStatusChanged: Cannot continue, failed to get watchedSeason!", "error", err)
		return
	}
	// If season not found, create it.
	if watchedSeason == nil {
		slog.Debug("hookEpisodeStatusChanged: Watched season does not exist. Creating now.")
		seasonStatus := newEpisodeStatus
		if newEpisodeStatus == FINISHED {
			seasonStatus = WATCHING
		}
		_, err := addWatchedSeason(db, userId, WatchedSeasonAddRequest{
			addActivity:  SEASON_ADDED_AUTO,
			WatchedID:    watchedId,
			SeasonNumber: seasonNum,
			Status:       seasonStatus,
		})
		if err != nil {
			slog.Error("hookEpisodeStatusChanged: Failed to add watched season!", "error", err)
		}
	} else if watchedSeason.Status == "" || watchedSeason.Status == PLANNED {
		watchedSeason.Status = WATCHING
		if res := db.Save(watchedSeason); res.Error != nil {
			slog.Error("hookEpisodeStatusChanged: Failed to update season status!", "error", res.Error)
		}
	}

	// 3. If the show has no status or is planned, set it to watching.
	watchedShow, err := getWatchedItemById(db, userId, watchedId)
	if err != nil {
		slog.Error("hookEpisodeStatusChanged: Failed to get watched show, cant continue to update show status.", "error", err)
		return
	} else {
		// Show status shouldn't be empty, but watevs, handle it just incase
		if watchedShow.Status == "" || watchedShow.Status == PLANNED {
			watchedShow.Status = WATCHING
			if res := db.Save(watchedShow); res.Error != nil {
				slog.Error("hookEpisodeStatusChanged: Failed to update show status!", "error", res.Error)
			}
		}
	}

	// 4. If all episodes are FINISHED or DROPPED, set the season to FINISHED
	tmdbIdStr := strconv.Itoa(watchedShow.Content.TmdbID)
	seasonNumStr := strconv.Itoa(seasonNum)
	seasonDetails, err := seasonDetails(tmdbIdStr, seasonNumStr)
	if err != nil {
		slog.Error("hookEpisodeStatusChanged: Failed to get season details!", "error", err)
		return
	}
	allEpisodesCount := len(seasonDetails.Episodes)
	finishedEpisodesCount, err := getNumberOfWatchedEpisodesInSeason(db, userId, watchedId, seasonNum, []WatchedStatus{FINISHED, DROPPED})
	if err != nil {
		slog.Error("hookEpisodeStatusChanged: Failed to get number of watched episodes in this season!", "error", err)
		return
	}
	slog.Debug("hookEpisodeStatusChanged: Got episode counts.", "allEpisodesCount", allEpisodesCount, "finishedEpisodesCount", finishedEpisodesCount)
	if finishedEpisodesCount >= int64(allEpisodesCount) {
		slog.Debug("hookEpisodeStatusChanged: All episodes have been completed (finished or dropped). Marking season finished.")
		if res := db.Model(&WatchedSeason{}).Where("watched_id = ? AND season_number = ? AND user_id = ?", watchedId, seasonNum, userId).Update("status", FINISHED); res.Error != nil {
			slog.Error("hookEpisodeStatusChanged: Failed to update season status to finished:", "error", res.Error.Error())
			return
		}
	}
}
