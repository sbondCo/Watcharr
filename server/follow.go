package main

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type Follow struct {
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	UserID         uint           `gorm:"primaryKey:usr_id_to_followed_id;not null;check:user_id != followed_user_id"`
	User           User
	FollowedUserID uint `gorm:"primaryKey:usr_id_to_followed_id;not null"`
	FollowedUser   User
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

func getFollows(db *gorm.DB, userId uint) ([]Follow, error) {
	var follows []Follow
	res := db.Where("user_id = ?", userId).Find(&follows)
	if res.Error != nil {
		return []Follow{}, errors.New("failed to find follows")
	}
	return follows, nil
}
