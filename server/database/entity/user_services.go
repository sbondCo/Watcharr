package entity

import (
	"github.com/sbondCo/Watcharr/database/dbmodel"
)

// Holds third party service auth tokens for users.
// Each service may use the fields in their own way.
// Unique index applied between service name and clientID
// to ensure no duplicates (no need to apply it against
// user_id, no accounts should share an integration).
//
// Jellyfin:
//   - AuthToken  : Users jellyfin auth token.
//   - AuthToken2 : Unused.
//
// Plex:
//   - AuthToken  : Used for requests against plex.tv
//   - AuthToken2 : Used for requests against home plex server.
type UserServices struct {
	dbmodel.GormModelNoDel
	// Service/integration name
	Name UserServiceName `gorm:"uniqueIndex:svc_name_to_cltid;not null;" json:"-"`
	// The users id on the third party service
	ClientID  string `gorm:"uniqueIndex:svc_name_to_cltid;not null;" json:"-"`
	AuthToken string `gorm:"not null;" json:"-"`
	// Second auth token, generic name so future services can use it without extra confusion.
	// Ex: We require a second auth token for use with our local server for Plex.
	AuthToken2 string `json:"-"`
	UserID     uint   `gorm:"not null;" json:"-"`
}

type UserServiceName string

const (
	UserServiceNameJellyfin UserServiceName = "jellyfin"
	UserServiceNamePlex     UserServiceName = "plex"
)
