package main

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

// Location middleware
func WhereaboutsRequired() gin.HandlerFunc {
	// TODO: This should also take into account a user
	// preference (of location.. which also needs to be
	// added) and server preference (env var). The US is
	// good enough for now as a default.
	return func(c *gin.Context) {
		slog.Debug("WhereaboutsRequired middleware hit", "country", "US")
		c.Set("userCountry", "US")
		c.Next()
	}
}
