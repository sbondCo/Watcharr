package main

import (
	"errors"
	"log/slog"

	"gorm.io/gorm"
)

type WatchedSeason struct {
	GormModel
	UserID       uint          `json:"-" gorm:"not null"`
	User         User          `json:"-"`
	WatchedID    uint          `json:"-" gorm:"not null"`
	SeasonNumber int           `json:"seasonNumber" gorm:"not null"`
	Status       WatchedStatus `json:"status"`
	Rating       int8          `json:"rating"`
}

type WatchedSeasonAddRequest struct {
	WatchedID    uint          `json:"watchedId"`
	SeasonNumber int           `json:"seasonNumber"`
	Status       WatchedStatus `json:"status"`
	Rating       int8          `json:"rating"`
}

// Add/edit a watched season.
func addWatchedSeason(db *gorm.DB, userId uint, ar WatchedSeasonAddRequest, at ActivityType) ([]WatchedSeason, error) {
	slog.Debug("Adding watched season item", "userId", userId, "watchedID", ar.WatchedID, "season", ar.SeasonNumber)
	// 1. Make sure watched item exists and it is the correct type (TV)
	var w Watched
	if resp := db.Where("id = ? AND user_id = ?", ar.WatchedID, userId).Preload("Content").Preload("WatchedSeasons").Find(&w); resp.Error != nil {
		slog.Error("Failed when adding a watched season", "error", "failed to get watched item from db")
		return []WatchedSeason{}, errors.New("failed when retrieving watched item")
	}
	if w.ID == 0 {
		slog.Error("Failed when adding a watched season", "error", "watched item does not exist in db")
		return []WatchedSeason{}, errors.New("can't add a watched season for a show that doesnt have a status itself")
	}
	if w.Content.Type != SHOW {
		return []WatchedSeason{}, errors.New("can't add watched season for non show content")
	}
	var found bool
	for i, ws := range w.WatchedSeasons {
		slog.Debug("loop", "1", ws.SeasonNumber, "2", ar.SeasonNumber)
		if ws.SeasonNumber == ar.SeasonNumber {
			slog.Debug("Existing watched season item found, updating existing")
			found = true
			w.WatchedSeasons[i].Status = ar.Status
			w.WatchedSeasons[i].Rating = ar.Rating
			break
		}
	}
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
		return []WatchedSeason{}, errors.New("failed to save")
	}
	return w.WatchedSeasons, nil
}

// Remove a watched season
func rmWatchedSeason(db *gorm.DB, userId uint, seasonId uint) error {
	slog.Debug("rmWatchedSeason called", "user_id", userId, "season_id", seasonId)
	if resp := db.Model(&WatchedSeason{}).Unscoped().Where("id = ? AND user_id = ?", seasonId, userId).Delete(&WatchedSeason{}); resp.Error != nil {
		slog.Error("Failed when removing a watched season", "error", resp.Error)
		return errors.New("failed when removing watched season")
	}
	return nil
}
