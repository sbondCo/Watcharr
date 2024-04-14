package main

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

// Location middleware
func WhereaboutsRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		region := c.Query("region")
		slog.Debug("WhereaboutsRequired: middleware hit", "region", region)
		if region == "" {
			// If no region is passed, default to server region.
			if Config.DEFAULT_COUNTRY != "" {
				slog.Debug("WhereaboutsRequired: Using server default country.", "default_country", Config.DEFAULT_COUNTRY)
				c.Set("userCountry", Config.DEFAULT_COUNTRY)
				c.Next()
				return
			}
			// If no server region set, default to US.
			slog.Debug("WhereaboutsRequired: Using hard coded default (US).")
			c.Set("userCountry", "US")
			c.Next()
			return
		}
		c.Set("userCountry", region)
		c.Next()
	}
}
