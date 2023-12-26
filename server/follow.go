package main

import (
	"errors"
	"log/slog"
	"time"

	"gorm.io/gorm"
)

// Database struct, only internal.
type Follow struct {
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"-"`
	UserID         uint      `gorm:"primaryKey:usr_id_to_followed_id;not null;check:user_id != followed_user_id" json:"-"`
	User           User      `json:"-"`
	FollowedUserID uint      `gorm:"primaryKey:usr_id_to_followed_id;not null" json:"-"`
	FollowedUser   User      `json:"-"`
}

// For end users to see.
type FollowPublic struct {
	CreatedAt    time.Time  `json:"createdAt"`
	FollowedUser PublicUser `json:"followedUser"`
}

type FollowThoughts struct {
	FollowedUser PublicUser    `json:"followedUser"`
	Thoughts     string        `json:"thoughts"`
	Status       WatchedStatus `json:"status"`
	Rating       int8          `json:"rating"`
}

func followUser(db *gorm.DB, currentUserId uint, toFollowUserId uint) (FollowPublic, error) {
	f := Follow{UserID: currentUserId, FollowedUserID: toFollowUserId}
	res := db.Model(&Follow{}).Create(&f)
	if res.Error != nil {
		slog.Error("followUser: Error on inserting follow.", "error", res.Error)
		err := "failed to insert follow"
		if res.Error == gorm.ErrDuplicatedKey {
			err = "already followed"
		}
		return FollowPublic{}, errors.New(err)
	}
	// Now get the row with preloaded followed user
	var nf Follow
	res = db.Where("user_id = ? AND followed_user_id = ?", currentUserId, toFollowUserId).Preload("FollowedUser", "private = ?", 0).Take(&nf)
	if res.Error != nil {
		slog.Error("followUser: Couldn't fetch newly followed user.", "error", res.Error)
		return FollowPublic{}, errors.New("followed, but failed to fetch followed user")
	}
	return FollowPublic{CreatedAt: nf.CreatedAt, FollowedUser: nf.FollowedUser.GetSafe()}, nil
}

func unfollowUser(db *gorm.DB, currentUserId uint, toFollowUserId uint) (bool, error) {
	f := Follow{UserID: currentUserId, FollowedUserID: toFollowUserId}
	res := db.Delete(&f)
	if res.Error != nil {
		slog.Error("unfollowUser: Error deleting follow.", "error", res.Error)
		err := "failed to remove follow"
		if res.Error == gorm.ErrRecordNotFound {
			err = "not following"
		}
		return false, errors.New(err)
	}
	return true, nil
}

// Get current users follows
func getFollows(db *gorm.DB, userId uint) ([]FollowPublic, error) {
	var follows []Follow
	res := db.Where("user_id = ?", userId).Preload("FollowedUser", "private = ?", 0).Find(&follows)
	if res.Error != nil {
		slog.Error("getFollows: Error finding follows.", "error", res.Error)
		return []FollowPublic{}, errors.New("failed to find follows")
	}
	fpub := []FollowPublic{}
	for _, v := range follows {
		// Skip followed users without an ID..
		// this will be because they have made
		// their account private after we followed them.
		if v.FollowedUser.ID == 0 {
			continue
		}
		fpub = append(fpub, FollowPublic{CreatedAt: v.CreatedAt, FollowedUser: v.FollowedUser.GetSafe()})
	}
	return fpub, nil
}

// Get followed profile thoughts, rating, etc on specific content.
func getFollowsThoughts(db *gorm.DB, userId uint, mediaType ContentType, tmdbId string) ([]FollowThoughts, error) {
	var follows []Follow
	res := db.Where("user_id = ?", userId).Preload("FollowedUser", "private = ? AND private_thoughts = ?", 0, 0).Find(&follows)
	if res.Error != nil {
		slog.Error("getFollows: Error finding follows.", "error", res.Error)
		return []FollowThoughts{}, errors.New("failed to find follows")
	}
	slog.Info("getFollowsThoughts")
	var followIds []uint
	for _, v := range follows {
		// Skip empty followedUsers.. they are private.
		if v.FollowedUser.ID == 0 {
			continue
		}
		followIds = append(followIds, v.FollowedUser.ID)
	}
	// Get our content id from type and tmdbId
	var content Content
	res = db.Where("type = ? AND tmdb_id = ?", mediaType, tmdbId).Select("id").Find(&content)
	if res.Error != nil {
		slog.Error("getFollows: Error finding content from db.", "error", res.Error)
		return []FollowThoughts{}, errors.New("failed to find content")
	}
	// Get list of followeds watcheds for this content
	var fw []Watched
	res = db.Where("content_id = ? AND user_id IN ?", content.ID, followIds).Find(&fw)
	if res.Error != nil {
		slog.Error("getFollows: Error finding followed watcheds from db.", "error", res.Error)
		return []FollowThoughts{}, errors.New("failed to find followed watcheds")
	}
	slog.Debug("someting", "lol", fw)
	// Create followThoughts array by combining follows and fw(atcheds)
	ft := []FollowThoughts{}
	for _, v := range fw {
		var fu PublicUser
		for _, f := range follows {
			if f.FollowedUser.ID == v.UserID {
				fu = f.FollowedUser.GetSafe()
				break
			}
		}
		// If we didn't find a related followedUser.. skip this watched entry
		if fu.ID == 0 {
			continue
		}
		ft = append(ft, FollowThoughts{FollowedUser: fu, Thoughts: v.Thoughts, Status: v.Status, Rating: v.Rating})
	}
	return ft, nil
}
