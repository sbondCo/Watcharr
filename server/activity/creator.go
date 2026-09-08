package activity

import (
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/sbondCo/Watcharr/database/entity"
	"gorm.io/gorm"
)

// Activity builder.
// Must use the constructor! Required properties are placed in the constructor
// params so that we can't build Watcharr if any are not included!
type Creator struct {
	db          *gorm.DB
	userID      uint
	watchedID   uint
	typ         entity.ActivityType
	Data        string
	CustomDate  *time.Time
	countAsPlay bool
	createdBy   entity.ActivityCreatedBy
	Reason      string
}

func NewCreator(
	db *gorm.DB,
	userID uint,
	watchedID uint,
	typ entity.ActivityType,
	countAsPlay bool,
	// Underlying type is `int`. Pass `0` if we don't need to set `createdBy`.
	createdBy entity.ActivityCreatedBy,
) *Creator {
	return &Creator{
		db:          db,
		userID:      userID,
		watchedID:   watchedID,
		typ:         typ,
		countAsPlay: countAsPlay,
		createdBy:   createdBy,
	}
}

// Set Type.
func (c *Creator) SetType(d entity.ActivityType) *Creator { c.typ = d; return c }

// Set Data.
func (c *Creator) SetData(d string) *Creator { c.Data = d; return c }

// Set Data. Pass in map `d` which is marshalled into JSON string.
func (c *Creator) SetDataJSON(d map[string]any) *Creator {
	j, err := json.Marshal(d)
	if err != nil {
		slog.Error("SetDataJSON: Marshalling JSON failed!", "error", err)
		// Not fatal, data will just have to be an empty string.
	}
	c.Data = string(j)
	return c
}

// Set CustomDate. If `d` is zero value, it is set a `nil`.
func (c *Creator) SetCustomDate(d *time.Time) *Creator {
	if d != nil && d.IsZero() {
		// If `d` IsZero, set to nil.
		d = nil
	}
	c.CustomDate = d
	return c
}

// Set Reason.
func (c *Creator) SetReason(d string) *Creator { c.Reason = d; return c }

// Set CountAsPlay.
func (c *Creator) SetCountAsPlay(d bool) *Creator { c.countAsPlay = d; return c }

// Add this Creator to a MultiCreator for later saving multiple activities.
func (c *Creator) AddToMultiCreator(mc *MultiCreator) {
	mc.AddCreator(c)
}

// Done building activity.. now create it.
// NOTE: This func doesn't verify if `userId` owns the referenced watched ID.
// The caller (eg, a route handler) should do that!
func (c *Creator) Create() (entity.Activity, error) {
	if c.watchedID == 0 {
		return entity.Activity{},
			errors.New("watchedId must be set to add an activity")
	}
	activity := entity.Activity{
		UserID:      c.userID,
		WatchedID:   c.watchedID,
		Type:        c.typ,
		Data:        c.Data,
		CustomDate:  c.CustomDate,
		CountAsPlay: c.countAsPlay,
		CreatedBy:   c.createdBy,
		Reason:      c.Reason,
	}
	res := c.db.Create(&activity)
	if res.Error != nil {
		slog.Error("Create: Query failed!", "error", res.Error)
		return entity.Activity{}, errors.New("query failed")
	}
	slog.Debug("Create: Added activity", "added_activity", activity)
	return activity, nil
}
