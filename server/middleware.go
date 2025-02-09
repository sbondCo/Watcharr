package main

import (
	"log/slog"
	"net/http"
	"strconv"

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

// Pagination middleware
// Reusable way to get pagination values.
// If force=true then will default to page=1, otherwise
// assume pagination is disabled when query params not present.
func PaginatedRequest(force bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		pageStr := c.Query("p")
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
				slog.Error("PossiblyPaginated: Query paramater 'p' was not parseable as an int", "err", err)
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "query param 'p' must be a number"})
				return
			}
			page = num
		}
		limitStr := c.Query("l")
		limit := 40
		if limitStr != "" {
			num, err := strconv.Atoi(limitStr)
			if err != nil {
				slog.Error("PossiblyPaginated: Query paramater 'l' was not parseable as an int", "err", err)
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "query param 'l' must be a number"})
				return
			}
			limit = num
		} else {
			slog.Debug("PossiblyPaginated: Using default limit.")
		}
		slog.Debug("PossiblyPaginated: middleware hit", "page", page, "page_limit", limit)
		c.Set("paginationEnabled", true)
		c.Set("paginationParams", PaginationParams{
			Page:  page,
			Limit: limit,
		})
		c.Next()
	}
}
