package domain

import (
	"time"

	"github.com/sbondCo/Watcharr/database/entity"
)

type (
	// Internal struct accepted by AddActivity function.
	ActivityAddProps struct {
		WatchedID  uint                `json:"watchedId" binding:"required"`
		Type       entity.ActivityType `json:"type" binding:"required"`
		Data       string              `json:"data" binding:"required"`
		CustomDate *time.Time          `json:"customDate,omitempty"`
	}

	ActivityUpdateRequest struct {
		CustomDate time.Time `json:"customDate" binding:"required"`
	}

	ActivityAddProvider interface {
		AddActivity(
			userId uint,
			ar ActivityAddProps,
			countAsPlay bool,
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
