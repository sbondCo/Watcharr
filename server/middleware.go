package main

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Location middleware
func WhereaboutsRequired(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		settings, err := userGetSettings(db, userId)

		if err != nil {
			slog.Error("Failed to get user settings", "error", err.Error())
			c.Set("userCountry", "US")
		}

		slog.Debug("WhereaboutsRequired middleware hit", "country", *settings.Country)
		c.Set("userCountry", *settings.Country)
		c.Next()
	}
}
