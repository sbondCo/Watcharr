package main

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ApiKey struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Value     string    `gorm:"not null" json:"value"`
	UserID    uint      `gorm:"not null" json:"-"`
	Note      string    `gorm:"not null" json:"note"`
}

type ApiKeyRequest struct {
	Note string `json:"note" binding:"required"`
}

func createAPIKey(db *gorm.DB, userId uint, note string) (*ApiKey, error) {
	token, err := generateString(32)
	if err != nil {
		return nil, err
	}

	apiKey := ApiKey{
		Value:  token,
		UserID: userId,
		Note:   note,
	}

	res := db.Create(&apiKey)
	return &apiKey, res.Error
}

func getAllAPIKeys(db *gorm.DB, userId uint) ([]ApiKey, error) {
	apiKeys := new([]ApiKey)
	res := db.Model(&ApiKey{}).Where("user_id = ?", userId).Find(&apiKeys)
	return *apiKeys, res.Error
}

func getAPIKey(db *gorm.DB, userId uint, apiKeyId string) (*ApiKey, error) {
	var apiKey ApiKey
	res := db.Model(&ApiKey{}).Where("user_id = ?", userId).Where("id = ?", apiKeyId).First(&apiKey)
	return &apiKey, res.Error
}

func deleteAPIKey(db *gorm.DB, userId uint, apiKeyId string) error {
	res := db.Model(&ApiKey{}).Where("user_id = ?", userId).Where("id = ?", apiKeyId).Delete(&ApiKey{})
	return res.Error
}

func APIKeyMiddleware(db *gorm.DB, key string, ctx *gin.Context) {
	slog.Debug("API Key middleware hit")

	var apiKey ApiKey

	result := db.First(&apiKey, "value=?", key)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.Set("userId", apiKey.UserID)
	ctx.Next()
}
