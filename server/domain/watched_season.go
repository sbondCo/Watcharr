package domain

import (
	"time"

	"github.com/sbondCo/Watcharr/database/entity"
)

type (
	WatchedSeasonSetRequest struct {
		WatchedID    uint                 `json:"watchedId"`
		SeasonNumber int                  `json:"seasonNumber"`
		Status       entity.WatchedStatus `json:"status"`
		Rating       int8                 `json:"rating" binding:"max=10"`

		AddActivityDate   time.Time                `json:"-"`
		ActivityReason    string                   `json:"-"`
		ActivityCreatedBy entity.ActivityCreatedBy `json:"-"`
	}

	WatchedSeasonSetResponse struct {
		// The watched season.
		WatchedSeason entity.WatchedSeason `json:"watchedSeason"`
		// If the returned WatchedSeason was updated (update action).
		// True = WatchedSeason was updated (because it already existed).
		// False = it was created (because it didn't already exist).
		Update bool `json:"update"`
		// Added activities.
		AddedActivities []entity.Activity `json:"addedActivities"`
		// Response from hook
		StatusChangedHookResponse SeasonStatusChangedHookResponse `json:"statusChangedHookResponse,omitzero"`
	}

	SeasonStatusChangedHookResponse struct {
		// The watched shows status if we modified it.
		NewShowStatus entity.WatchedStatus `json:"newShowStatus,omitempty"`
		// All activies we have added.
		AddedActivities []entity.Activity `json:"addedActivities,omitempty"`
		// All errors (fatal and non-fatal) that were encountered.
		Errors []string `json:"errors,omitempty"`
	}

	// Set Watched Season provider.
	SetWatchedSeasonProvider interface {
		SetWatchedSeason(userId uint, ar WatchedSeasonSetRequest) (WatchedSeasonSetResponse, error)
	}
)
