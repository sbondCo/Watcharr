package domain

import (
	"time"

	"github.com/sbondCo/Watcharr/database/entity"
)

type (
	ActivityUpdateRequest struct {
		CustomDate time.Time `json:"customDate" binding:"required"`
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
