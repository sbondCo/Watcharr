package follow

import (
	"errors"
	"log/slog"
	"time"

	"github.com/sbondCo/Watcharr/database/entity"
	"gorm.io/gorm"
)

// For end users to see.
type FollowPublic struct {
	CreatedAt    time.Time         `json:"createdAt"`
	FollowedUser entity.PublicUser `json:"followedUser"`
}

type FollowThoughts struct {
	FollowedUser entity.PublicUser    `json:"followedUser"`
	Thoughts     string               `json:"thoughts"`
	Status       entity.WatchedStatus `json:"status"`
	Rating       float64              `json:"rating"`
}

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		db,
	}
}

func (s *Service) FollowUser(currentUserId uint, toFollowUserId uint) (FollowPublic, error) {
	f := entity.Follow{UserID: currentUserId, FollowedUserID: toFollowUserId}
	res := s.db.Model(&entity.Follow{}).Create(&f)
	if res.Error != nil {
		slog.Error("followUser: Error on inserting follow.", "error", res.Error)
		err := "failed to insert follow"
		if res.Error == gorm.ErrDuplicatedKey {
			err = "already followed"
		}
		return FollowPublic{}, errors.New(err)
	}
	// Now get the row with preloaded followed user
	var nf entity.Follow
	res = s.db.Where("user_id = ? AND followed_user_id = ?", currentUserId, toFollowUserId).Preload("FollowedUser", "private = ?", 0).Take(&nf)
	if res.Error != nil {
		slog.Error("followUser: Couldn't fetch newly followed user.", "error", res.Error)
		return FollowPublic{}, errors.New("followed, but failed to fetch followed user")
	}
	return FollowPublic{CreatedAt: nf.CreatedAt, FollowedUser: nf.FollowedUser.GetSafe()}, nil
}

func (s *Service) UnfollowUser(currentUserId uint, toFollowUserId uint) (bool, error) {
	f := entity.Follow{UserID: currentUserId, FollowedUserID: toFollowUserId}
	res := s.db.Delete(&f)
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
func (s *Service) GetFollows(userId uint) ([]FollowPublic, error) {
	var follows []entity.Follow
	res := s.db.Where("user_id = ?", userId).Preload("FollowedUser", "private = ?", 0).Find(&follows)
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
func (s *Service) GetFollowsThoughts(userId uint, mediaType string, mediaId string) ([]FollowThoughts, error) {
	var follows []entity.Follow
	res := s.db.Where("user_id = ?", userId).Preload("FollowedUser", "private = ? AND private_thoughts = ?", 0, 0).Find(&follows)
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
	var contentGameOrBookId int
	if mediaType == "game" {
		// Get our content id from type and tmdbId
		var content entity.Game
		res = s.db.Where("igdb_id = ?", mediaId).Select("id").Find(&content)
		if res.Error != nil {
			slog.Error("getFollows: Error finding content from db.", "error", res.Error)
			return []FollowThoughts{}, errors.New("failed to find content")
		}
		contentGameOrBookId = content.ID
	} else if mediaType == "book" {
		// Get our content id from type and tmdbId
		var content entity.Book
		res = s.db.Where("ol_id = ?", mediaId).Select("id").Find(&content)
		if res.Error != nil {
			slog.Error("getFollows: Error finding content from db.", "error", res.Error)
			return []FollowThoughts{}, errors.New("failed to find content")
		}
		contentGameOrBookId = content.ID
	} else if mediaType == "movie" || mediaType == "tv" {
		// Get our content id from type and tmdbId
		var content entity.Content
		res = s.db.Where("type = ? AND tmdb_id = ?", mediaType, mediaId).Select("id").Find(&content)
		if res.Error != nil {
			slog.Error("getFollows: Error finding content from db.", "error", res.Error)
			return []FollowThoughts{}, errors.New("failed to find content")
		}
		contentGameOrBookId = content.ID
	} else {
		slog.Error("getFollows: Unrecognized media type (movie, tv or game supported).", "media_type", mediaType)
		return []FollowThoughts{}, errors.New("unrecognized media type")
	}
	// Get list of followeds watcheds for this content
	var fw []entity.Watched
	if mediaType == "game" {
		res = s.db.Where("game_id = ? AND user_id IN ?", contentGameOrBookId, followIds).Find(&fw)
	} else if mediaType == "book" {
		res = s.db.Where("book_id = ? AND user_id IN ?", contentGameOrBookId, followIds).Find(&fw)
	} else {
		res = s.db.Where("content_id = ? AND user_id IN ?", contentGameOrBookId, followIds).Find(&fw)
	}
	if res.Error != nil {
		slog.Error("getFollows: Error finding followed watcheds from db.", "error", res.Error)
		return []FollowThoughts{}, errors.New("failed to find followed watcheds")
	}
	// Create followThoughts array by combining follows and fw(atcheds)
	ft := []FollowThoughts{}
	for _, v := range fw {
		var fu entity.PublicUser
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
