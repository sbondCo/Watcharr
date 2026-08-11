package domain

import (
	"time"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/tri"
)

type (
	// Internal struct accepted by AddActivity function.
	ActivityAddProps struct {
		WatchedID  uint                `json:"watchedId" binding:"required"`
		Type       entity.ActivityType `json:"type" binding:"required"`
		Data       string              `json:"data" binding:"required"`
		CustomDate *time.Time          `json:"customDate,omitempty"`
	}

	ActivityAddExtraProps struct {
		// If this activity counts as a play.
		CountAsPlay tri.State
		// Activity synced by?
		SyncedBy entity.ActivitySyncedBy
	}

	ActivityUpdateRequest struct {
		CustomDate time.Time `json:"customDate" binding:"required"`
	}

	ActivityAddProvider interface {
		AddActivity(
			userId uint,
			ar ActivityAddProps,
			extra ActivityAddExtraProps,
		) (entity.Activity, error)
	}
)

// Looks through Activity for Watched entry and calculates the amount
// that count as plays.
func getPlaysFromActivity(a []entity.Activity) int {
	plays := 0
	for i := range a {
		if a[i].CountAsPlay {
			plays++
		}
	}
	return plays
}
