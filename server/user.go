package main

import (
	"errors"
	"log/slog"

	"gorm.io/gorm"
)

// This really only works if all settings are repassed in, otherwise they
// will be reset to their default value.
// When more settings are added in, this should be updated to work so you
// only have to pass in the setting you want to update.
func userUpdate(db *gorm.DB, userId uint, ur UserSettings) (UserSettings, error) {
	slog.Debug("user update request running", "user_id", userId, "ur", ur)
	user := new(User)
	res := db.Where("id = ?", userId).Take(&user)
	if res.Error != nil {
		slog.Error("user update failed", "user_id", userId, "error", "failed to retrieve user from database")
		return UserSettings{}, errors.New("failed to retrieve user")
	}
	user.Private = ur.Private
	db.Save(&user)
	return UserSettings{Private: user.Private}, nil
}

func userGetSettings(db *gorm.DB, userId uint) (UserSettings, error) {
	slog.Debug("user update request running", "user_id", userId)
	user := new(User)
	res := db.Where("id = ?", userId).Take(&user)
	if res.Error != nil {
		slog.Error("user get failed", "user_id", userId, "error", "failed to retrieve user from database")
		return UserSettings{}, errors.New("failed to retrieve user")
	}
	return UserSettings{Private: user.Private}, nil
}
