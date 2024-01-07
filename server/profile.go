package main

import (
	"errors"
	"log/slog"
	"time"

	"gorm.io/gorm"
)

type Profile struct {
	Joined               time.Time `json:"joined"`
	ShowsWatched         int32     `json:"showsWatched"`
	MoviesWatched        int32     `json:"moviesWatched"`
	MoviesWatchedRuntime uint32    `json:"moviesWatchedRuntime"`
	ShowsWatchedRuntime  uint32    `json:"showsWatchedRuntime"`
}

// Gets any data required for profile page
func getProfile(db *gorm.DB, userId uint) (Profile, error) {
	user := new(User)
	res := db.Model(&User{}).Where("id = ?", userId).Take(&user)
	if res.Error != nil {
		slog.Error("Failed to get profile:", "error", res.Error.Error())
		return Profile{}, errors.New("failed to get profile")
	}
	watched := new([]Watched)
	res = db.Model(&Watched{}).Preload("Content").Where("user_id = ?", userId).Find(&watched)
	if res.Error != nil {
		slog.Error("Profile: Failed to get watched for processing:", "error", res.Error.Error())
		return Profile{}, errors.New("failed to get watched for processing")
	}
	var (
		showsWatched         int32
		moviesWatched        int32
		moviesWatchedRuntime uint32
		showsWatchedRuntime  uint32
	)
	for _, w := range *watched {
		if w.Status == FINISHED {
			if w.Content.Type == SHOW {
				showsWatched++
				// This aint a science, just a very inaccurate guesstimate.
				if w.Content.NumberOfEpisodes != 0 {
					var showRuntime uint32 = 30
					if w.Content.Runtime != 0 {
						showRuntime = w.Content.Runtime
					}
					showsWatchedRuntime += showRuntime * w.Content.NumberOfEpisodes
					slog.Debug("calcualted", "show", w.Content.Title, "runti", showRuntime*w.Content.NumberOfEpisodes)
				}
			} else if w.Content.Type == MOVIE {
				moviesWatched++
				moviesWatchedRuntime += w.Content.Runtime
			}
		}
	}
	profile := Profile{
		Joined:               user.CreatedAt,
		ShowsWatched:         showsWatched,
		MoviesWatched:        moviesWatched,
		MoviesWatchedRuntime: moviesWatchedRuntime,
		ShowsWatchedRuntime:  showsWatchedRuntime,
	}
	return profile, nil
}
