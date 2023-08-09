package main

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type ActivityType string

var (
	ADDED_WATCHED    ActivityType = "ADDED_WATCHED"
	REMOVED_WATCHED  ActivityType = "REMOVED_WATCHED"
	RATING_CHANGED   ActivityType = "RATING_CHANGED"
	STATUS_CHANGED   ActivityType = "STATUS_CHANGED"
	THOUGHTS_CHANGED ActivityType = "THOUGHTS_CHANGED"
	THOUGHTS_REMOVED ActivityType = "THOUGHTS_REMOVED"
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
}

type ActivityAddRequest struct {
	WatchedID uint         `json:"watchedId" binding:"required"`
	Type      ActivityType `json:"type" binding:"required"`
	Data      string       `json:"data" binding:"required"`
}

func getActivity(db *gorm.DB, userId uint, watchedId uint) ([]Activity, error) {
	activity := new([]Activity)
	res := db.Model(&Activity{}).Where("user_id = ? AND watched_id = ?", userId, watchedId).Find(&activity)
	if res.Error != nil {
		println("Failed getting activity from database:", res.Error.Error())
		return []Activity{}, errors.New("failed getting activity")
	}
	return *activity, nil
}

func addActivity(db *gorm.DB, userId uint, ar ActivityAddRequest) (Activity, error) {
	if ar.WatchedID == 0 {
		return Activity{}, errors.New("watchedId must be set to add an activity")
	}
	activity := Activity{UserID: userId, WatchedID: ar.WatchedID, Type: ar.Type, Data: ar.Data}
	res := db.Create(&activity)
	if res.Error != nil {
		println("Error adding activity to database:", res.Error.Error())
		return Activity{}, errors.New("failed adding new activity to database")
	}
	fmt.Printf("%+v\n", activity)
	return activity, nil
}
