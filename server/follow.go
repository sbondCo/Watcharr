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
