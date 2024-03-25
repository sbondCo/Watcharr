package main

import (
	"errors"
	"log/slog"
	"time"

	"gorm.io/gorm"
)

// User details wanted for management views.
type ManagedUser struct {
	ID          uint      `json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	Username    string    `json:"username"`
	Type        UserType  `json:"type"`
	Permissions int       `json:"permissions"`
	Private     bool      `json:"private"`
}

type UpdateUserRequest struct {
	Permissions *int `json:"permissions"`
}

func getAllUsers(db *gorm.DB) ([]ManagedUser, error) {
	users := []ManagedUser{}
	if res := db.Model(&User{}).Find(&users); res.Error != nil {
		slog.Error("getAllUsers: Failed to fetch users from database", "error", res.Error)
		return []ManagedUser{}, errors.New("failed to fetch users from database")
	}
	return users, nil
}

// Update a user. For management views, for admin to update another user.
func manageUser(db *gorm.DB, userId uint, ur UpdateUserRequest) error {
	// Error now if no userId or any UpdateUserRequest property was provided.
	if userId == 0 || (ur.Permissions == nil) {
		slog.Error("manageUser: invalid arguments", "user_id", userId)
		return errors.New("invalid arguments, ensure a valid userId and at least one property has been provided for updating")
	}
	toUpdate := map[string]interface{}{}
	if ur.Permissions != nil {
		if *ur.Permissions == 0 {
			// If removing all perms, set to default of 1 (PERM_NONE).
			// Will avoid confusion and possibly bugs later on, though I doubt
			// we'd ever be checking a user to ensure they have no perms.
			toUpdate["permissions"] = PERM_NONE
		} else {
			toUpdate["permissions"] = *ur.Permissions
		}
	}
	if res := db.Model(&User{}).Where("id = ?", userId).Updates(toUpdate); res.Error != nil {
		slog.Error("manageUser: failed to update user in database", "user_id", userId, "error", res.Error)
		return errors.New("failed to update user in database")
	}
	slog.Debug("manageUser: A user has been updated", "user_id", userId)
	return nil
}
