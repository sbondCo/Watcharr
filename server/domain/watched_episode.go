package domain

import (
	"time"

	"github.com/sbondCo/Watcharr/database/entity"
)

type (
	WatchedEpisodeSetRequest struct {
		WatchedID     uint                 `json:"watchedId"`
		SeasonNumber  int                  `json:"seasonNumber"`
		EpisodeNumber int                  `json:"episodeNumber"`
		Status        entity.WatchedStatus `json:"status"`
		Rating        int8                 `json:"rating" binding:"max=10"`

		AddActivityDate   time.Time                `json:"-"`
		AddActivityReason string                   `json:"-"`
		ActivityCreatedBy entity.ActivityCreatedBy `json:"-"`
	}

	WatchedEpisodeSetResponse struct {
		// The watched episode.
		WatchedEpisode entity.WatchedEpisode `json:"watchedEpisode"`
		// If the returned WatchedEpisode was updated (update action).
		// True = WatchedEpisode was updated (because it already existed).
		// False = it was created (because it didn't already exist).
		Update bool `json:"update"`
		// Added activities.
		AddedActivities []entity.Activity `json:"addedActivities,omitempty"`
		// Response from hook
		EpisodeStatusChangedHookResponse EpisodeStatusChangedHookResponse `json:"episodeStatusChangedHookResponse,omitzero"`
	}

	EpisodeStatusChangedHookResponse struct {
		// The watched shows status if we modified it.
		NewShowStatus entity.WatchedStatus `json:"newShowStatus,omitempty"`
		// The full watched season (if created or modified).
		WatchedSeason *entity.WatchedSeason `json:"watchedSeason,omitempty"`
		// All activies we have added.
		AddedActivities []entity.Activity `json:"addedActivities,omitempty"`
		// All errors (fatal and non-fatal) that were encountered.
		Errors []string `json:"errors,omitempty"`
	}

	// Set Watched Episode provider.
	SetWatchedEpisodeProvider interface {
		SetWatchedEpisode(userId uint, ar WatchedEpisodeSetRequest) (WatchedEpisodeSetResponse, error)
	}
)
