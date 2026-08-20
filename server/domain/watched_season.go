package domain

import (
	"time"

	"github.com/sbondCo/Watcharr/database/entity"
)

type (
	WatchedSeasonAddRequest struct {
		WatchedID    uint                 `json:"watchedId"`
		SeasonNumber int                  `json:"seasonNumber"`
		Status       entity.WatchedStatus `json:"status"`
		Rating       int8                 `json:"rating" binding:"max=10"`

		AddActivityDate      time.Time                `json:"-"`
		AddActivityReason    string                   `json:"-"`
		AddActivityCreatedBy entity.ActivityCreatedBy `json:"-"`
	}

	WatchedSeasonAddResponse struct {
		WatchedSeasons []entity.WatchedSeason `json:"watchedSeasons"`
		AddedActivity  entity.Activity        `json:"addedActivity"`
	}
)
