package domain

import (
	"time"

	"github.com/sbondCo/Watcharr/database/entity"
)

type (
	WatchedEpisodeAddRequest struct {
		WatchedID       uint                 `json:"watchedId"`
		SeasonNumber    int                  `json:"seasonNumber"`
		EpisodeNumber   int                  `json:"episodeNumber"`
		Status          entity.WatchedStatus `json:"status"`
		Rating          int8                 `json:"rating" binding:"max=10"`
		AddActivity     entity.ActivityType  `json:"-"`
		AddActivityDate time.Time            `json:"-"`
		// Set the SyncedBy value on the created activity (only use if coming from sync job).
		ActivitySyncedBy entity.ActivitySyncedBy `json:"-"`
	}

	WatchedEpisodeAddResponse struct {
		WatchedEpisodes []entity.WatchedEpisode `json:"watchedEpisodes"`
		AddedActivity   entity.Activity         `json:"addedActivity"`
		// Response from hook
		EpisodeStatusChangedHookResponse EpisodeStatusChangedHookResponse `json:"episodeStatusChangedHookResponse,omitempty"`
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

	// Add Watched Episode provider.
	AddWatchedEpisodeProvider interface {
		AddWatchedEpisodes(userId uint, ar WatchedEpisodeAddRequest) (WatchedEpisodeAddResponse, error)
	}
)
