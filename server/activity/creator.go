package activity

import (
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

func (c *Creator) SetType(d entity.ActivityType) *Creator { c.typ = d; return c }
func (c *Creator) SetData(d string) *Creator              { c.Data = d; return c }
func (c *Creator) SetCustomDate(d *time.Time) *Creator    { c.CustomDate = d; return c }
func (c *Creator) SetReason(d string) *Creator            { c.Reason = d; return c }
func (c *Creator) SetCountAsPlay(d bool) *Creator         { c.countAsPlay = d; return c }

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
		slog.Error("Error adding activity to database",
			"error", res.Error.Error())
		return entity.Activity{},
			errors.New("failed adding new activity to database")
	}
	slog.Debug("Adding activity", "added_activity", activity)
	return activity, nil
}
