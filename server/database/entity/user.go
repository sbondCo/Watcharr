package entity

import (
	"github.com/sbondCo/Watcharr/database/dbmodel"
)

type UserType uint8

var (
	WATCHARR_USER UserType = 0
	JELLYFIN_USER UserType = 1
	PLEX_USER     UserType = 2
	// Registered via trusted header auth
	PROXY_USER UserType = 3
)

// User Perms
// iota auto increments for us so when adding new
// perms, add to bottom as to not change other perm
// values.
const (
	PERM_NONE int = 1 << iota
	PERM_ADMIN
	PERM_REQUEST_CONTENT
	PERM_REQUEST_CONTENT_AUTO_APPROVE
)

// uniqueIndex applied between Username and UserType, so same usernames can exist, but only with different types.
// This is incase different users with same name from different services try to signup.
type User struct {
	dbmodel.GormModel
	Username string `gorm:"uniqueIndex:usr_name_to_type;not null" json:"username" binding:"required"`
	Password string `gorm:"not null" json:"password" binding:"required"`
	AvatarID uint   `json:"-"`
	Avatar   Image  `json:"avatar"`
	Bio      string `json:"bio"`
	// The type of user/which auth service they originate from.
	// Empty if from Watcharr, or the name of the service (eg. jellyfin)
	Type UserType `gorm:"uniqueIndex:usr_name_to_type;not null;default:0" json:"type"`
	// Users third party integrations
	UserServices []UserServices `json:"-"`
	Watched      []Watched
	// All Tags
	Tags []Tag `json:"-"`
	// Users permissions
	Permissions int `gorm:"default:1" json:"-"`
	// All user settings cols, in another struct for reusability
	UserSettings
}

func (u *User) GetSafe() PublicUser {
	return PublicUser{
		ID:       u.ID,
		Username: u.Username,
		Avatar:   u.Avatar,
		Bio:      u.Bio,
	}
}

// This struct uses pointer to the values, so in update user settings,
// we can tell which setting is being updated (if not nil..).
type UserSettings struct {
	// Is profile private
	Private *bool `gorm:"default:false" json:"private"`
	// Are watched list content thoughts public (profile must also be public is false)
	PrivateThoughts *bool `gorm:"default:false" json:"privateThoughts"`
	// If ui 'spoilers' should be shown
	HideSpoilers *bool `gorm:"default:false" json:"hideSpoilers"`
	// If user wants previously watched items to show in 'Finished' filter,
	// even if the watched item state has since been changed.
	// Also if user wants to show in watched stats.
	IncludePreviouslyWatched *bool `gorm:"default:false" json:"includePreviouslyWatched"`
	// User's country to get correct content streaming providers.
	// TODO Enforce iso_3166_1 validity (same as tmdb)
	Country *string `gorm:"default:'US'" json:"country"`
	// Does the user want show, season and episode automations enabled.
	AutomateShowStatuses *bool `gorm:"default:true" json:"automateShowStatuses"`
	// Rating system user wants to use (frontend only).
	// RatingSystem enum in frontend maxes out at 3, so just max=3 on this and we should be gut.
	RatingSystem *int `json:"ratingSystem" binding:"omitempty,max=3"`
	// Rating step for supported rating systems (frontend only, enum goes up to 2).
	RatingStep *int `json:"ratingStep" binding:"omitempty,max=2"`
}

// Public user details for search results
type PublicUser struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	AvatarID uint   `json:"-"`
	Avatar   Image  `json:"avatar,omitzero"`
	Bio      string `json:"bio,omitempty"`
}

// Private user details, for returning users details to themselves
type PrivateUser struct {
	Username    string   `json:"username"`
	Type        UserType `json:"type"`
	Permissions int      `json:"permissions"`
	AvatarID    uint     `json:"-"`
	Avatar      Image    `json:"avatar"`
	Bio         string   `json:"bio"`
}
