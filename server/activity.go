package main

import (
	"errors"
	"log/slog"
	"time"

	"gorm.io/gorm"
)

type ActivityType string

// _AUTO activities are for when logic updates something for the user (automations basically).
var (
	ADDED_WATCHED               ActivityType = "ADDED_WATCHED"
	REMOVED_WATCHED             ActivityType = "REMOVED_WATCHED"
	RATING_CHANGED              ActivityType = "RATING_CHANGED"
	STATUS_CHANGED              ActivityType = "STATUS_CHANGED"
	STATUS_CHANGED_AUTO         ActivityType = "STATUS_CHANGED_AUTO"
	THOUGHTS_CHANGED            ActivityType = "THOUGHTS_CHANGED"
	THOUGHTS_REMOVED            ActivityType = "THOUGHTS_REMOVED"
	IMPORTED_WATCHED            ActivityType = "IMPORTED_WATCHED"
	IMPORTED_WATCHED_JF         ActivityType = "IMPORTED_WATCHED_JF"
	IMPORTED_WATCHED_PLEX       ActivityType = "IMPORTED_WATCHED_PLEX"
	IMPORTED_RATING             ActivityType = "IMPORTED_RATING"        // Imported rating, but with no rating acts as original import of content to old platform (where they are importing from) activity
	IMPORTED_ADDED_WATCHED      ActivityType = "IMPORTED_ADDED_WATCHED" // Imported watched date, so we can save the original watch dates of content from users old platform (where they are importing from).
	IMPORTED_ADDED_WATCHED_JF   ActivityType = "IMPORTED_ADDED_WATCHED_JF"
	IMPORTED_ADDED_WATCHED_PLEX ActivityType = "IMPORTED_ADDED_WATCHED_PLEX"
	SEASON_ADDED                ActivityType = "SEASON_ADDED"
	SEASON_ADDED_AUTO           ActivityType = "SEASON_ADDED_AUTO"
	SEASON_ADDED_JF             ActivityType = "SEASON_ADDED_JF"
	SEASON_ADDED_PLEX           ActivityType = "SEASON_ADDED_PLEX"
	SEASON_REMOVED              ActivityType = "SEASON_REMOVED"
	SEASON_RATING_CHANGED       ActivityType = "SEASON_RATING_CHANGED"
	SEASON_STATUS_CHANGED       ActivityType = "SEASON_STATUS_CHANGED"
	SEASON_STATUS_CHANGED_AUTO  ActivityType = "SEASON_STATUS_CHANGED_AUTO"
	EPISODE_ADDED               ActivityType = "EPISODE_ADDED"
	EPISODE_ADDED_JF            ActivityType = "EPISODE_ADDED_JF"
	EPISODE_ADDED_PLEX          ActivityType = "EPISODE_ADDED_PLEX"
	EPISODE_REMOVED             ActivityType = "EPISODE_REMOVED"
	EPISODE_RATING_CHANGED      ActivityType = "EPISODE_RATING_CHANGED"
	EPISODE_STATUS_CHANGED      ActivityType = "EPISODE_STATUS_CHANGED"
)

type Activity struct {
	GormModel
	// ID of user this activity is linked to, so it can be easily
	// secured (users can only view their own activities).
	UserID uint `json:"-" gorm:"not null"`
	// ID of watched list item this activity is linked to.
	WatchedID uint `json:"watchedId" gorm:"not null"`
	// Type of activity.
	Type ActivityType `json:"type" gorm:"not null"`
	// Holds custom data (ex, if rating changed, this can
	// hold new rating - if status changed, this will hold that).
	Data string `json:"data" gorm:"not null"`
	// Custom date for the activity, that the user can define.
	CustomDate *time.Time `json:"customDate,omitempty"`
}

type ActivityAddRequest struct {
	WatchedID  uint         `json:"watchedId" binding:"required"`
	Type       ActivityType `json:"type" binding:"required"`
	Data       string       `json:"data" binding:"required"`
	CustomDate *time.Time   `json:"customDate,omitempty"`
}

type ActivityUpdateRequest struct {
	CustomDate time.Time `json:"customDate" binding:"required"`
}

func getActivity(db *gorm.DB, userId uint, watchedId uint) ([]Activity, error) {
	activity := new([]Activity)
	res := db.Model(&Activity{}).Where("user_id = ? AND watched_id = ?", userId, watchedId).Find(&activity)
	if res.Error != nil {
		slog.Error("Failed getting activity from database", "error", res.Error.Error())
		return []Activity{}, errors.New("failed getting activity")
	}
	return *activity, nil
}

func addActivity(db *gorm.DB, userId uint, ar ActivityAddRequest) (Activity, error) {
	if ar.WatchedID == 0 {
		return Activity{}, errors.New("watchedId must be set to add an activity")
	}
	activity := Activity{UserID: userId, WatchedID: ar.WatchedID, Type: ar.Type, Data: ar.Data, CustomDate: ar.CustomDate}
	res := db.Create(&activity)
	if res.Error != nil {
		slog.Error("Error adding activity to database", "error", res.Error.Error())
		return Activity{}, errors.New("failed adding new activity to database")
	}
	slog.Debug("Adding activity", "added_activity", activity)
	return activity, nil
}

func updateActivity(db *gorm.DB, userId uint, id uint, activityUpdateRequest ActivityUpdateRequest) error {
	if id == 0 {
		return errors.New("id must be set to update an activity")
	}
	if activityUpdateRequest.CustomDate.IsZero() {
		return errors.New("customDate must be set to update an activity")
	}
	res := db.Model(&Activity{}).Where("user_id = ? AND id = ?", userId, id).Update("custom_date", activityUpdateRequest.CustomDate)
	if res.Error != nil {
		slog.Error("Error updating activity in database", "error", res.Error.Error())
		return errors.New("failed updating activity in database")
	}
	if res.RowsAffected < 1 {
		slog.Error("No activities were updated. This may be because the activity doesn't exist or is not owned by the calling user.")
		return errors.New("failed updating activity in database")
	}
	slog.Debug("Updating activity", "updated_activity", id)
	return nil
}

func deleteActivity(db *gorm.DB, userId uint, id uint) error {
	if id == 0 {
		return errors.New("an id must be provided to delete an activity")
	}
	res := db.Where("user_id = ?", userId).Delete(&Activity{}, id)
	if res.Error != nil {
		slog.Error("Error deleting activity in database", "error", res.Error.Error())
		return errors.New("failed deleting activity in database")
	}
	if res.RowsAffected < 1 {
		slog.Error("No activities were deleted. This may be because the activity doesn't exist or is not owned by the calling user.")
		return errors.New("failed deleting activity from database")
	}
	return nil
}
