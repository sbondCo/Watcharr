package main

import (
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
