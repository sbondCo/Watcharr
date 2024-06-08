package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UniqueIndex applied between WatchedID, SeasonNum and EpisodeNum to avoid duplicates incase logic fails.
//
// Episodes on tmdb are only queried by season number + episode number, not possible via episode id,
// since episodes can be removed and re-added. For this reason we store season and episodes nums instead
// of just the episode id.
type WatchedEpisode struct {
	GormModel
	UserID        uint          `json:"-" gorm:"not null"`
	User          User          `json:"-"`
	WatchedID     uint          `json:"-" gorm:"uniqueIndex:we_watched_to_ens;not null"`
	SeasonNumber  int           `json:"seasonNumber" gorm:"uniqueIndex:we_watched_to_ens;not null"`
	EpisodeNumber int           `json:"episodeNumber" gorm:"uniqueIndex:we_watched_to_ens;not null"`
	Status        WatchedStatus `json:"status"`
	Rating        int8          `json:"rating"`
}

type WatchedEpisodeAddRequest struct {
	WatchedID       uint          `json:"watchedId"`
	SeasonNumber    int           `json:"seasonNumber"`
	EpisodeNumber   int           `json:"episodeNumber"`
	Status          WatchedStatus `json:"status"`
	Rating          int8          `json:"rating"`
	addActivity     ActivityType  `json:"-"`
	addActivityDate time.Time     `json:"-"`
}

type WatchedEpisodeAddResponse struct {
	WatchedEpisodes []WatchedEpisode `json:"watchedEpisodes"`
	AddedActivity   Activity         `json:"addedActivity"`
	// Response from hook
	EpisodeStatusChangedHookResponse EpisodeStatusChangedHookResponse `json:"episodeStatusChangedHookResponse,omitempty"`
}

// Add/edit a watched episode.
func addWatchedEpisodes(db *gorm.DB, userId uint, ar WatchedEpisodeAddRequest) (WatchedEpisodeAddResponse, error) {
	slog.Debug("Adding watched episode item", "userId", userId, "watchedID", ar.WatchedID, "season", ar.SeasonNumber, "episode", ar.EpisodeNumber)
	// 1. Make sure watched item exists and it is the correct type (TV)
	var w Watched
	if resp := db.Where("id = ? AND user_id = ?", ar.WatchedID, userId).Preload("Content").Preload("WatchedEpisodes").Find(&w); resp.Error != nil {
		slog.Error("Failed when adding a watched episode", "error", "failed to get watched item from db")
		return WatchedEpisodeAddResponse{}, errors.New("failed when retrieving watched item")
	}
	if w.ID == 0 {
		slog.Error("Failed when adding a watched episode", "error", "watched item does not exist in db")
		return WatchedEpisodeAddResponse{}, errors.New("can't add a watched episode for a show that doesnt have a status itself")
	}
	if w.Content.Type != SHOW {
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
	var addedActivity Activity
	if !found {
		slog.Debug("Existing watched episode not found, adding as new entry")
		w.WatchedEpisodes = append(w.WatchedEpisodes, WatchedEpisode{
			UserID:        userId,
			WatchedID:     ar.WatchedID,
			SeasonNumber:  ar.SeasonNumber,
			EpisodeNumber: ar.EpisodeNumber,
			Status:        ar.Status,
			Rating:        ar.Rating,
		})
	}
	if resp := db.Save(&w.WatchedEpisodes); resp.Error != nil {
		slog.Debug("Failed to save watched episode item in db", "error", resp.Error)
		return WatchedEpisodeAddResponse{}, errors.New("failed to save")
	}
	// Add activity
	if found {
		// Only add change activity if we actually updated a value
		// (changing value to same value doesn't count).
		if updated {
			if ar.Status != "" {
				json, _ := json.Marshal(map[string]interface{}{"season": ar.SeasonNumber, "episode": ar.EpisodeNumber, "status": ar.Status})
				addedActivity, _ = addActivity(db, userId, ActivityAddRequest{WatchedID: w.ID, Type: EPISODE_STATUS_CHANGED, Data: string(json)})
			}
			if ar.Rating != 0 {
				json, _ := json.Marshal(map[string]interface{}{"season": ar.SeasonNumber, "episode": ar.EpisodeNumber, "rating": ar.Rating})
				addedActivity, _ = addActivity(db, userId, ActivityAddRequest{WatchedID: w.ID, Type: EPISODE_RATING_CHANGED, Data: string(json)})
			}
		}
	} else {
		json, _ := json.Marshal(map[string]interface{}{"season": ar.SeasonNumber, "episode": ar.EpisodeNumber, "status": ar.Status, "rating": ar.Rating})
		act := ActivityAddRequest{WatchedID: w.ID, Type: EPISODE_ADDED, Data: string(json)}
		if ar.addActivity != "" {
			act.Type = ar.addActivity
		}
		if !ar.addActivityDate.IsZero() {
			act.CustomDate = &ar.addActivityDate
		}
		addedActivity, _ = addActivity(db, userId, act)
	}
	episodeAddResp := WatchedEpisodeAddResponse{
		WatchedEpisodes: w.WatchedEpisodes,
		AddedActivity:   addedActivity,
	}
	if ar.Status != "" {
		slog.Debug("addWatchedEpisodes: Episode status was changed, calling hook.")
		episodeAddResp.EpisodeStatusChangedHookResponse = hookEpisodeStatusChanged(db, userId, ar.WatchedID, ar.SeasonNumber, ar.EpisodeNumber, ar.Status)
	}
	return episodeAddResp, nil
}

// Remove a watched episode
func rmWatchedEpisode(db *gorm.DB, userId uint, id uint) (Activity, error) {
	slog.Debug("rmWatchedSeason called", "user_id", userId, "id", id)
	var watchedEpisode WatchedEpisode
	resp := db.Clauses(clause.Returning{}).Model(&WatchedEpisode{}).Unscoped().Where("id = ? AND user_id = ?", id, userId).Delete(&watchedEpisode)
	if resp.Error != nil {
		slog.Error("Failed when removing a watched episode", "error", resp.Error)
		return Activity{}, errors.New("failed when removing watched episode")
	}
	if resp.RowsAffected == 0 {
		slog.Error("Failed when removing a watched episode", "error", "zero rows affected")
		return Activity{}, errors.New("wasn't removed from db.. may not exist")
	}
	slog.Debug("rmWatchedEpisode, deleted row", "row", watchedEpisode)
	if watchedEpisode.ID != 0 {
		json, _ := json.Marshal(map[string]interface{}{
			"season":  watchedEpisode.SeasonNumber,
			"episode": watchedEpisode.EpisodeNumber,
			"status":  watchedEpisode.Status,
			"rating":  watchedEpisode.Rating,
		})
		addedActivity, _ := addActivity(db, userId, ActivityAddRequest{WatchedID: watchedEpisode.WatchedID, Type: EPISODE_REMOVED, Data: string(json)})
		return addedActivity, nil
	}
	return Activity{}, errors.New("removed, but failed to add activity entry")
}

func getNumberOfWatchedEpisodesInSeason(db *gorm.DB, userId uint, watchedId uint, seasonNumber int, acceptableStatus []WatchedStatus) (int64, error) {
	var count int64
	if res := db.Model(&WatchedEpisode{}).Where("user_id = ? AND watched_id = ? AND season_number = ? AND status IN ?", userId, watchedId, seasonNumber, acceptableStatus).Count(&count); res.Error != nil {
		return 0, res.Error
	}
	return count, nil
}
