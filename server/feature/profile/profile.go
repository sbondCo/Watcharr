package profile

import (
	"errors"
	"log/slog"
	"time"

	"github.com/sbondCo/Watcharr/database/entity"
	"gorm.io/gorm"
)

type Profile struct {
	Joined               time.Time `json:"joined"`
	ShowsWatched         int32     `json:"showsWatched"`
	MoviesWatched        int32     `json:"moviesWatched"`
	MoviesWatchedRuntime uint32    `json:"moviesWatchedRuntime"`
	ShowsWatchedRuntime  uint32    `json:"showsWatchedRuntime"`
}

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		db,
	}
}

// Checks if item has been previously watched by scanning for any activity
// that counts as a play.
func (s *Service) hasBeenPreviouslyWatched(a *[]entity.Activity) bool {
	for _, v := range *a {
		if v.CountAsPlay {
			return true
		}
	}
	return false
}

// Gets any data required for profile page
func (s *Service) getProfile(userId uint) (Profile, error) {
	// Get user.
	user := new(entity.User)
	res := s.db.Model(&entity.User{}).Where("id = ?", userId).Take(&user)
	if res.Error != nil {
		slog.Error("Failed to get profile:",
			"error", res.Error)
		return Profile{}, errors.New("failed to get profile")
	}

	// Process stats.
	watched := new([]entity.Watched)
	res = s.db.Model(&entity.Watched{}).
		Preload("Content").
		Preload("Activity").
		Where("user_id = ?", userId).
		Find(&watched)
	if res.Error != nil {
		slog.Error("Profile: Failed to get watched for processing:",
			"error", res.Error)
		return Profile{}, errors.New("failed to get watched for processing")
	}
	var (
		showsWatched         int32
		moviesWatched        int32
		moviesWatchedRuntime uint32
		showsWatchedRuntime  uint32
	)
	for _, w := range *watched {
		isFinished := false
		// Note: Deliberately always checking `hasBeenPreviouslyWatched` for any
		// items without status set to FINISHED without checking users
		// `IncludePreviouslyWatched` setting, because that setting is useful
		// for filters, BUT not for these stats. I think it is always expected
		// that all previously watched stuff is included in finished stats.
		if w.Status == entity.FINISHED || s.hasBeenPreviouslyWatched(&w.Activity) {
			isFinished = true
		}
		if isFinished {
			if w.Content == nil {
				continue
			}
			c := *w.Content
			switch c.Type {
			case entity.SHOW:
				showsWatched++
				// This aint a science, just a very inaccurate guesstimate.
				if c.NumberOfEpisodes != 0 {
					var showRuntime uint32 = 30
					if c.Runtime != 0 {
						showRuntime = c.Runtime
					}
					showsWatchedRuntime += showRuntime * c.NumberOfEpisodes
					slog.Debug("profile stat calculated",
						"show", c.Title,
						"runti", showRuntime*c.NumberOfEpisodes)
				}
			case entity.MOVIE:
				moviesWatched++
				moviesWatchedRuntime += c.Runtime
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
