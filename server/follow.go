package main

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Database struct, only internal.
type Follow struct {
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"-"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
	UserID         uint           `gorm:"primaryKey:usr_id_to_followed_id;not null;check:user_id != followed_user_id" json:"-"`
	User           User           `json:"-"`
	FollowedUserID uint           `gorm:"primaryKey:usr_id_to_followed_id;not null" json:"-"`
	FollowedUser   User           `json:"-"`
}

// For end users to see.
type FollowPublic struct {
	CreatedAt    time.Time  `json:"createdAt"`
	FollowedUser PublicUser `json:"followedUser"`
}

func followUser(db *gorm.DB, currentUserId uint, toFollowUserId uint) (Follow, error) {
	f := Follow{UserID: currentUserId, FollowedUserID: toFollowUserId}
	res := db.Model(&Follow{}).Create(&f)
	if res.Error != nil {
		err := "failed to insert follow"
		if res.Error == gorm.ErrDuplicatedKey {
			err = "already followed"
		}
		return Follow{}, errors.New(err)
	}
	return f, nil
}

// TODO UserId must be public user (unless current user)
func getFollows(db *gorm.DB, userId uint) ([]FollowPublic, error) {
	var follows []Follow
	res := db.Where("user_id = ?", userId).Preload("FollowedUser", "private = ?", 0).Find(&follows)
	if res.Error != nil {
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
