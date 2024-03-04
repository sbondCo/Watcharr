package main

import (
	"encoding/json"
	"errors"
	"log/slog"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UniqueIndex applied between WatchedID and SeasonNumber to avoid duplicates incase logic fails.
type WatchedSeason struct {
	GormModel
	UserID       uint          `json:"-" gorm:"not null"`
	User         User          `json:"-"`
	WatchedID    uint          `json:"-" gorm:"uniqueIndex:ws_watched_to_season_num;not null"`
	SeasonNumber int           `json:"seasonNumber" gorm:"uniqueIndex:ws_watched_to_season_num;not null"`
	Status       WatchedStatus `json:"status"`
	Rating       int8          `json:"rating"`
}

type WatchedSeasonAddRequest struct {
	WatchedID    uint          `json:"watchedId"`
	SeasonNumber int           `json:"seasonNumber"`
	Status       WatchedStatus `json:"status"`
	Rating       int8          `json:"rating"`
}

type WatchedSeasonAddResponse struct {
	WatchedSeasons []WatchedSeason `json:"watchedSeasons"`
	AddedActivity  Activity        `json:"addedActivity"`
}

// Add/edit a watched season.
func addWatchedSeason(db *gorm.DB, userId uint, ar WatchedSeasonAddRequest) (WatchedSeasonAddResponse, error) {
	slog.Debug("Adding watched season item", "userId", userId, "watchedID", ar.WatchedID, "season", ar.SeasonNumber)
	// 1. Make sure watched item exists and it is the correct type (TV)
	var w Watched
	if resp := db.Where("id = ? AND user_id = ?", ar.WatchedID, userId).Preload("Content").Preload("WatchedSeasons").Find(&w); resp.Error != nil {
		slog.Error("Failed when adding a watched season", "error", "failed to get watched item from db")
		return WatchedSeasonAddResponse{}, errors.New("failed when retrieving watched item")
	}
	if w.ID == 0 {
		slog.Error("Failed when adding a watched season", "error", "watched item does not exist in db")
		return WatchedSeasonAddResponse{}, errors.New("can't add a watched season for a show that doesnt have a status itself")
	}
	if w.Content.Type != SHOW {
		return WatchedSeasonAddResponse{}, errors.New("can't add watched season for non show content")
	}
	found := false
	updated := false
	for i, ws := range w.WatchedSeasons {
		if ws.SeasonNumber == ar.SeasonNumber {
			slog.Debug("Existing watched season item found, updating existing")
			found = true
			if ar.Status != "" && ar.Status != w.WatchedSeasons[i].Status {
				w.WatchedSeasons[i].Status = ar.Status
				updated = true
			}
			if ar.Rating != 0 && ar.Rating != w.WatchedSeasons[i].Rating {
				w.WatchedSeasons[i].Rating = ar.Rating
				updated = true
			}
			break
		}
	}
	var addedActivity Activity
	if !found {
		slog.Debug("Existing watched season not found, adding as new entry")
		w.WatchedSeasons = append(w.WatchedSeasons, WatchedSeason{
			UserID:       userId,
			WatchedID:    ar.WatchedID,
			SeasonNumber: ar.SeasonNumber,
			Status:       ar.Status,
			Rating:       ar.Rating,
		})
	}
	if resp := db.Save(&w.WatchedSeasons); resp.Error != nil {
		slog.Debug("Failed to save watched season item in db", "error", resp.Error)
		return WatchedSeasonAddResponse{}, errors.New("failed to save")
	}
	// Add activity
	if found {
		// Only add change activity if we actually updated a value
		// (changing value to same value doesn't count).
		if updated {
			if ar.Status != "" {
				json, _ := json.Marshal(map[string]interface{}{"season": ar.SeasonNumber, "status": ar.Status})
				addedActivity, _ = addActivity(db, userId, ActivityAddRequest{WatchedID: w.ID, Type: SEASON_STATUS_CHANGED, Data: string(json)})
			}
			if ar.Rating != 0 {
				json, _ := json.Marshal(map[string]interface{}{"season": ar.SeasonNumber, "rating": ar.Rating})
				addedActivity, _ = addActivity(db, userId, ActivityAddRequest{WatchedID: w.ID, Type: SEASON_RATING_CHANGED, Data: string(json)})
			}
		}
	} else {
		json, _ := json.Marshal(map[string]interface{}{"season": ar.SeasonNumber, "status": ar.Status, "rating": ar.Rating})
		addedActivity, _ = addActivity(db, userId, ActivityAddRequest{WatchedID: w.ID, Type: SEASON_ADDED, Data: string(json)})
	}
	return WatchedSeasonAddResponse{
		WatchedSeasons: w.WatchedSeasons,
		AddedActivity:  addedActivity,
	}, nil
}

// Remove a watched season
func rmWatchedSeason(db *gorm.DB, userId uint, seasonId uint) (Activity, error) {
	slog.Debug("rmWatchedSeason called", "user_id", userId, "season_id", seasonId)
	var watchedSeason WatchedSeason
	resp := db.Clauses(clause.Returning{}).Model(&WatchedSeason{}).Unscoped().Where("id = ? AND user_id = ?", seasonId, userId).Delete(&watchedSeason)
	if resp.Error != nil {
		slog.Error("Failed when removing a watched season", "error", resp.Error)
		return Activity{}, errors.New("failed when removing watched season")
	}
	if resp.RowsAffected == 0 {
		slog.Error("Failed when removing a watched season", "error", "zero rows affected")
		return Activity{}, errors.New("wasn't removed from db.. may not exist")
	}
	slog.Debug("rmWatchedSeason, deleted row", "row", watchedSeason)
	if watchedSeason.ID != 0 {
		json, _ := json.Marshal(map[string]interface{}{
			"season": watchedSeason.SeasonNumber,
			"status": watchedSeason.Status,
			"rating": watchedSeason.Rating,
		})
		addedActivity, _ := addActivity(db, userId, ActivityAddRequest{WatchedID: watchedSeason.WatchedID, Type: SEASON_REMOVED, Data: string(json)})
		return addedActivity, nil
	}
	return Activity{}, errors.New("removed, but failed to add activity entry")
}
