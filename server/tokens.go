package main

import (
	"errors"
	"log/slog"
	"time"

	"gorm.io/gorm"
)

type TokenType string

var (
	TOKENTYPE_ADMIN TokenType = "ADMIN"
)

type Token struct {
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time `json:"createdAt"`
	Value     string    `gorm:"not null"`
	Type      TokenType `gorm:"not null"`
	UserID    uint      `gorm:"not null"`
}

const tokenMaxAge = 2 * time.Minute

func createOneUseToken(db *gorm.DB, t TokenType, userId uint) (string, error) {
	token, err := generateString(8)
	if err != nil {
		slog.Error("createOneUseToken: Failed to generate string!", "error", err)
		return "", errors.New("failed to generate token")
	}
	res := db.Create(&Token{Type: t, Value: token, UserID: userId})
	if res.Error != nil {
		slog.Error("createOneUseToken: Failed to insert token into db!", "error", res.Error)
		return "", errors.New("failed to generate token")
	}
	return token, nil
}

// Cleans up tokens older than 2m.
func cleanupTokens(db *gorm.DB) {
	slog.Debug("cleanupTokens: Cleaning up old tokens from db")
	twoMinsAgo := time.Now().Add(-tokenMaxAge)
	resp := db.Where("created_at < ?", twoMinsAgo).Delete(&Token{})
	if resp.Error != nil {
		slog.Error("cleanupTokens: Failed to run DELETE on old tokens!", "error", resp.Error)
	}
}
