package entity

import (
	"time"

	"github.com/sbondCo/Watcharr/database/dbmodel"
)

type ActivityType string

// _AUTO activities are for when logic updates something for the user (automations basically).
var (
	ADDED_WATCHED               ActivityType = "ADDED_WATCHED"
	REMOVED_WATCHED             ActivityType = "REMOVED_WATCHED"
	RATING_CHANGED              ActivityType = "RATING_CHANGED"
	STATUS_CHANGED              ActivityType = "STATUS_CHANGED"
	STATUS_CHANGED_AUTO         ActivityType = "STATUS_CHANGED_AUTO"
	THOUGHTS_CHANGED            ActivityType = "THOUGHTS_CHANGED"
	THOUGHTS_REMOVED            ActivityType = "THOUGHTS_REMOVED"
	IMPORTED_WATCHED            ActivityType = "IMPORTED_WATCHED"
	IMPORTED_WATCHED_JF         ActivityType = "IMPORTED_WATCHED_JF"
	IMPORTED_WATCHED_PLEX       ActivityType = "IMPORTED_WATCHED_PLEX"
	IMPORTED_RATING             ActivityType = "IMPORTED_RATING"        // Imported rating, but with no rating acts as original import of content to old platform (where they are importing from) activity
	IMPORTED_ADDED_WATCHED      ActivityType = "IMPORTED_ADDED_WATCHED" // Imported watched date, so we can save the original watch dates of content from users old platform (where they are importing from).
	IMPORTED_ADDED_WATCHED_JF   ActivityType = "IMPORTED_ADDED_WATCHED_JF"
	IMPORTED_ADDED_WATCHED_PLEX ActivityType = "IMPORTED_ADDED_WATCHED_PLEX"
	SEASON_ADDED                ActivityType = "SEASON_ADDED"
	SEASON_ADDED_AUTO           ActivityType = "SEASON_ADDED_AUTO"
	SEASON_ADDED_JF             ActivityType = "SEASON_ADDED_JF"
	SEASON_ADDED_PLEX           ActivityType = "SEASON_ADDED_PLEX"
	SEASON_REMOVED              ActivityType = "SEASON_REMOVED"
	SEASON_RATING_CHANGED       ActivityType = "SEASON_RATING_CHANGED"
	SEASON_STATUS_CHANGED       ActivityType = "SEASON_STATUS_CHANGED"
	SEASON_STATUS_CHANGED_AUTO  ActivityType = "SEASON_STATUS_CHANGED_AUTO"
	EPISODE_ADDED               ActivityType = "EPISODE_ADDED"
	EPISODE_ADDED_JF            ActivityType = "EPISODE_ADDED_JF"
	EPISODE_ADDED_PLEX          ActivityType = "EPISODE_ADDED_PLEX"
	EPISODE_REMOVED             ActivityType = "EPISODE_REMOVED"
	EPISODE_RATING_CHANGED      ActivityType = "EPISODE_RATING_CHANGED"
	EPISODE_STATUS_CHANGED      ActivityType = "EPISODE_STATUS_CHANGED"
)

type Activity struct {
	dbmodel.GormModel
	// ID of user this activity is linked to, so it can be easily
	// secured (users can only view their own activities).
	UserID uint `json:"-" gorm:"not null"`
	// ID of watched list item this activity is linked to.
	WatchedID uint `json:"watchedId" gorm:"not null;index"`
	// Type of activity.
	Type ActivityType `json:"type" gorm:"not null"`
	// Holds custom data (ex, if rating changed, this can
	// hold new rating - if status changed, this will hold that).
	Data string `json:"data" gorm:"not null"`
	// Custom date for the activity, that the user can define.
	CustomDate *time.Time `json:"customDate,omitempty"`
	// Count this Activity as a Play?
	// Currently this was the best way I could see forward for implementing
	// counting plays of media that doesn't involve inefficient querying of
	// the actitivties table (or a whole new table, which would create extra
	// complexities itself, ie, plays/activity showing different records).
	// We write to this field when creating the activity to count is as a play
	// or not (ie when we create STATUS CHANGE activities with status of
	// FINISHED, imports, etc).
	// We won't support the user (or the system) modifying this value after
	// creation; if the user wants to delete a 'Play', they should delete the
	// activity.
	// Indexed (check migrations) to make search faster, since we frequently
	// do it over the whole table for watched sorting at the moment.
	CountAsPlay bool `json:"countAsPlay"`
	// If this activity wasn't created because of a manual interaction by the
	// user, the creator should store who they are in this property.
	// Eg: Jellyfin webhook syncer will store that it Created this activity.
	CreatedBy ActivityCreatedBy `json:"createdBy,omitempty"`
	// The reason behind this activity existing (if there's one worth setting).
	// Eg: Watcharr automations can place their full human-readable reason for
	// triggering in this property for display on the frontend.
	Reason string `json:"reason,omitempty"`
}

type ActivityCreatedBy int

// Obviously do not re-assign these, they link to db cells! New properties
// should be appended to the bottom! Note to self: I couldn't decide on a way
// of organizing these so I opted for a simple "enum".
const (
	// Watcharr automation.
	ActivityCreatedByWatcharr ActivityCreatedBy = iota + 1
	// Our generic importer service.
	ActivityCreatedByGenericImport
	// Jellyfin import.
	ActivityCreatedByJellyfinImport
	// Jellyfin webhook auto sync.
	ActivityCreatedByJellyfinWebhook
	// Plex import.
	ActivityCreatedByPlexImport
)
