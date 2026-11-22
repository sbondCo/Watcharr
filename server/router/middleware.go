package router

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/feature/user/usermiddleware"
	"github.com/sbondCo/Watcharr/util"
)

// Location middleware (usermiddleware.WithUser should be ran before this!)
func WhereaboutsRequired(cfg *config.ServerConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Respect region query param over all.
		region := c.Query("region")
		if region != "" {
			slog.Debug("WhereaboutsRequired: Using query param value",
				"region", region)
			c.Set("userCountry", region)
			c.Next()
			return
		}

		// Then respect user setting (set by usermiddleware.WithUser).
		user := usermiddleware.UserFromContext(c)
		if user.Country != nil && *user.Country != "" {
			slog.Debug("WhereaboutsRequired: Using user setting.",
				"user_country", *user.Country)
			c.Set("userCountry", *user.Country)
			c.Next()
			return
		}

		// Then use default value from server config if set.
		if cfg.DEFAULT_COUNTRY != "" {
			slog.Debug("WhereaboutsRequired: Using server default country.",
				"default_country", cfg.DEFAULT_COUNTRY)
			c.Set("userCountry", cfg.DEFAULT_COUNTRY)
			c.Next()
			return
		}

		// Finally, failsafe on hardcoded US value.
		slog.Debug("WhereaboutsRequired: Using hard coded default (US).")
		c.Set("userCountry", "US")
		c.Next()
	}
}

// Pagination middleware
// Reusable way to get pagination values.
// If force=true then will default to page=1, otherwise
// assume pagination is disabled when query params not present.
func PaginatedRequest(force bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Determine the wanted page.
		pageStr := c.Query("page")
		page := 0
		if pageStr == "" && force {
			page = 1
			slog.Debug("PossiblyPaginated: Pagination is forced, but no page was provided. Using default.")
		} else if pageStr == "" {
			slog.Debug("PossiblyPaginated: Pagination is disabled. No parameter provided.")
			c.Set("paginationEnabled", false)
			c.Next()
			return
		} else {
			num, err := strconv.Atoi(pageStr)
			if err != nil {
				slog.Error("PossiblyPaginated: Query paramater 'page' was not parseable as an int", "err", err)
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "query param 'page' must be a number"})
				return
			}
			page = num
		}
		// Determine the (page) limit
		limitStr := c.Query("limit")
		limit := 40
		if limitStr != "" {
			num, err := strconv.Atoi(limitStr)
			if err != nil {
				slog.Error("PossiblyPaginated: Query paramater 'limit' was not parseable as an int", "err", err)
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "query param 'limit' must be a number"})
				return
			}
			limit = num
		} else {
			slog.Debug("PossiblyPaginated: Using default limit.")
		}
		slog.Debug("PossiblyPaginated: middleware hit", "page", page, "limit", limit)
		// Set request context vars and continue.
		c.Set("paginationEnabled", true)
		c.Set("paginationParams", util.PaginationParams{
			Page:  page,
			Limit: limit,
		})
		c.Next()
	}
}
