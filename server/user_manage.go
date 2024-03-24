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

func getAllUsers(db *gorm.DB) ([]ManagedUser, error) {
	users := []ManagedUser{}
	if res := db.Model(&User{}).Find(&users); res.Error != nil {
		slog.Error("getAllUsers: Failed to fetch users from database", "error", res.Error)
		return []ManagedUser{}, errors.New("failed to fetch users from database")
	}
	return users, nil
}
