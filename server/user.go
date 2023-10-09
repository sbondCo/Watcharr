package main

import (
	"errors"
	"log/slog"

	"gorm.io/gorm"
)

// Public user details for search results
type PublicUser struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

// Private user details, for returning users details to themselves
type PrivateUser struct {
	Username    string   `json:"username"`
	Type        UserType `json:"type"`
	Permissions int      `json:"permissions"`
}

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

func userSearch(db *gorm.DB, currentUsersId uint, q string) ([]PublicUser, error) {
	slog.Debug("user search request running", "query", q)
	users := new([]PublicUser)
	res := db.Where("private = 0 AND username LIKE ? AND id != ?", "%"+q+"%", currentUsersId).Table("users").Find(&users)
	if res.Error != nil {
		slog.Error("user search failed", "error", "failed to query database")
		return []PublicUser{}, errors.New("failed to find users")
	}
	return *users, nil
}

func getUserInfo(db *gorm.DB, currentUsersId uint) (PrivateUser, error) {
	slog.Debug("user get info request running")
	user := new(PrivateUser)
	res := db.Where("id = ?", currentUsersId).Table("users").Take(&user)
	if res.Error != nil {
		slog.Error("user get info failed", "error", res.Error)
		return PrivateUser{}, errors.New("failed to find current user")
	}
	return *user, nil
}
