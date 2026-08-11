package domain

import (
	"time"

	"github.com/sbondCo/Watcharr/database/entity"
)

type (
	WatchedSeasonAddRequest struct {
		WatchedID       uint                 `json:"watchedId"`
		SeasonNumber    int                  `json:"seasonNumber"`
		Status          entity.WatchedStatus `json:"status"`
		Rating          int8                 `json:"rating" binding:"max=10"`
		AddActivity     entity.ActivityType  `json:"-"`
		AddActivityDate time.Time            `json:"-"`
		// Data to add to activity if the season is created.
		// Combined with data we already add.
		AddActivityData map[string]interface{} `json:"-"`
	}

	WatchedSeasonAddResponse struct {
		WatchedSeasons []entity.WatchedSeason `json:"watchedSeasons"`
		AddedActivity  entity.Activity        `json:"addedActivity"`
	}
)
