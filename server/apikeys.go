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
	Note string `json:"note,required"`
}

func (b *BaseRouter) addAPIKeyRoutes() {
	apiKeyGrp := b.rg.Group("/apikeys").Use(AuthRequired(nil))

	apiKeyGrp.POST("", func(ctx *gin.Context) {
		var req ApiKeyRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "'note' field is required",
			})
			return
		}

		userId := ctx.MustGet("userId").(uint)
		token, err := generateString(32)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "unable to generate an API key",
			})
			return
		}

		apiKey := ApiKey{
			Value:  token,
			UserID: userId,
			Note:   req.Note,
		}

		if res := b.db.Create(&apiKey); res.Error != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "unable to add API key",
			})
			return
		}

		ctx.JSON(http.StatusOK, apiKey)
	})

	apiKeyGrp.GET("", func(ctx *gin.Context) {
		userId := ctx.MustGet("userId").(uint)

		apiKeys := new([]ApiKey)
		if res := b.db.Model(&ApiKey{}).Where("user_id = ?", userId).Find(&apiKeys); res.Error != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "unable to add API key",
			})
			return
		}

		ctx.JSON(http.StatusOK, apiKeys)
	})

	apiKeyGrp.GET("/:id", func(ctx *gin.Context) {
		userId := ctx.MustGet("userId").(uint)

		var apiKey ApiKey
		if res := b.db.Model(&ApiKey{}).Where("user_id = ?", userId).Where("id = ?", ctx.Params.ByName("id")).First(&apiKey); res.Error != nil {
			if res.Error == gorm.ErrRecordNotFound {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}

			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "unable to add API key",
			})
			return
		}

		ctx.JSON(http.StatusOK, apiKey)
	})

	apiKeyGrp.DELETE("/:id", func(ctx *gin.Context) {
		userId := ctx.MustGet("userId").(uint)

		if res := b.db.Model(&ApiKey{}).Where("user_id = ?", userId).Where("id = ?", ctx.Params.ByName("id")).Delete(&ApiKey{}); res.Error != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "unable to add API key",
			})
			return
		}

		ctx.Status(http.StatusNoContent)
	})
}

func APIKeyMiddlewareWrapper(db *gorm.DB, mw gin.HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		slog.Debug("API Key middleware hit")
		key := ctx.GetHeader("x-api-key")
		if key == "" {
			mw(ctx)
			return
		}

		var apiKey ApiKey

		result := db.
			Model(&ApiKey{}).
			Where("value = ?", key).
			First(&apiKey)

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
}
