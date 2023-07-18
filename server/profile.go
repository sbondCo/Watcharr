package main

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type Profile struct {
	Joined        time.Time `json:"joined"`
	ShowsWatched  int32     `json:"showsWatched"`
	MoviesWatched int32     `json:"moviesWatched"`
}

// Gets any data required for profile page
func getProfile(db *gorm.DB, userId uint) (Profile, error) {
	user := new(User)
	res := db.Model(&User{}).Where("id = ?", userId).Take(&user)
	if res.Error != nil {
		println("Failed to get profile:", res.Error.Error())
		return Profile{}, errors.New("failed to get profile")
	}
	watched := new([]Watched)
	res = db.Model(&Watched{}).Preload("Content").Where("user_id = ?", userId).Find(&watched)
	if res.Error != nil {
		println("Profile: Failed to get watched for processing:", res.Error.Error())
		return Profile{}, errors.New("failed to get watched for processing")
	}
	var (
		showsWatched  int32
		moviesWatched int32
	)
	for _, w := range *watched {
		if w.Status == FINISHED {
			if w.Content.Type == SHOW {
				showsWatched++
			} else if w.Content.Type == MOVIE {
				moviesWatched++
			}
		}
	}
	profile := Profile{Joined: user.CreatedAt, ShowsWatched: showsWatched, MoviesWatched: moviesWatched}
	return profile, nil
}
