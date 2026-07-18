package user

import (
	"errors"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/image"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		db,
	}
}

// Update user settings
func (s *Service) UserUpdate(userId uint, ur entity.UserSettings) (entity.UserSettings, error) {
	slog.Debug("user update request running", "user_id", userId, "ur", ur)
	user := new(entity.User)
	res := s.db.Where("id = ?", userId).Take(&user)
	if res.Error != nil {
		slog.Error("user update failed", "user_id", userId, "error", res.Error)
		return entity.UserSettings{}, errors.New("failed to retrieve user")
	}
	if ur.HideSpoilers != nil {
		user.HideSpoilers = ur.HideSpoilers
	}
	if ur.Private != nil {
		user.Private = ur.Private
	}
	if ur.PrivateThoughts != nil {
		user.PrivateThoughts = ur.PrivateThoughts
	}
	if ur.IncludePreviouslyWatched != nil {
		user.IncludePreviouslyWatched = ur.IncludePreviouslyWatched
	}
	if ur.AutomateShowStatuses != nil {
		user.AutomateShowStatuses = ur.AutomateShowStatuses
	}
	if ur.Country != nil {
		user.Country = ur.Country
	}
	if ur.RatingSystem != nil {
		user.RatingSystem = ur.RatingSystem
	}
	if ur.RatingStep != nil {
		user.RatingStep = ur.RatingStep
	}
	s.db.Save(&user)
	return entity.UserSettings{
		Private:                  user.Private,
		PrivateThoughts:          user.PrivateThoughts,
		HideSpoilers:             user.HideSpoilers,
		IncludePreviouslyWatched: user.IncludePreviouslyWatched,
		AutomateShowStatuses:     user.AutomateShowStatuses,
		Country:                  user.Country,
	}, nil
}

func (s *Service) UserGetSettings(userId uint) (entity.UserSettings, error) {
	slog.Debug("UserGetSettings: Request running.", "user_id", userId)
	user := new(entity.User)
	res := s.db.Where("id = ?", userId).Take(&user)
	if res.Error != nil {
		slog.Error("user get failed", "user_id", userId, "error", res.Error)
		return entity.UserSettings{}, errors.New("failed to retrieve user")
	}
	return entity.UserSettings{
		Private:                  user.Private,
		PrivateThoughts:          user.PrivateThoughts,
		HideSpoilers:             user.HideSpoilers,
		IncludePreviouslyWatched: user.IncludePreviouslyWatched,
		AutomateShowStatuses:     user.AutomateShowStatuses,
		Country:                  user.Country,
		RatingSystem:             user.RatingSystem,
		RatingStep:               user.RatingStep,
	}, nil
}

func (s *Service) UserSearch(currentUsersId uint, q string) ([]entity.PublicUser, error) {
	slog.Debug("user search request running", "query", q)
	users := new([]entity.PublicUser)
	res := s.db.Where("private = 0 AND username LIKE ? AND id != ?", "%"+q+"%", currentUsersId).Table("users").Find(&users)
	if res.Error != nil {
		slog.Error("user search failed", "error", res.Error)
		return []entity.PublicUser{}, errors.New("failed to find users")
	}
	return *users, nil
}

func (s *Service) GetUserInfo(currentUsersId uint) (entity.PrivateUser, error) {
	slog.Debug("user get info request running")
	user := new(entity.PrivateUser)
	res := s.db.Where("id = ?", currentUsersId).Table("users").Preload("Avatar").Take(&user)
	if res.Error != nil {
		slog.Error("user get info failed", "error", res.Error)
		return entity.PrivateUser{}, errors.New("failed to find current user")
	}
	return *user, nil
}

// For getting a public user's info, when viewing their list for example
func (s *Service) GetUserPublicInfo(userId uint, username string) (entity.PublicUser, error) {
	slog.Debug("user get info request running")
	user := new(entity.PublicUser)
	res := s.db.Where("private = 0 AND id = ? AND username = ?", userId, username).Table("users").Preload("Avatar").Take(&user)
	if res.Error != nil {
		slog.Error("public user get info failed", "error", res.Error)
		return entity.PublicUser{}, errors.New("failed to find user")
	}
	return *user, nil
}

func (s *Service) UserUpdateBio(userId uint, newBio string) error {
	slog.Debug("userUpdateBio request running", "user_id", userId, "newBio", newBio)
	if res := s.db.Model(&entity.User{}).Where("id = ?", userId).Update("bio", newBio); res.Error != nil {
		slog.Error("userUpdateBio failed", "user_id", userId, "error", res.Error)
		return errors.New("failed to update bio")
	}
	return nil
}

func (s *Service) UploadUserAvatar(
	c *gin.Context,
	userId uint,
) (entity.Image, error) {
	file, err := c.FormFile("avatar")
	if err != nil {
		slog.Error("failed to get file", "error", err)
		return entity.Image{}, errors.New("no file found")
	}

	slog.Debug("UploadUserAvatar: An avatar is being uploaded",
		"name", file.Filename)

	f, _ := file.Open()
	defer f.Close()

	img, err := image.
		NewSaver(s.db, "up", image.ValidateOptions{}).
		DownloadAndInsert(f)
	if err != nil {
		slog.Error("UploadUserAvatar: DownloadAndInsert failed!",
			"error", err)
		return entity.Image{}, errors.New("processing image failed")
	}

	// No need to remove old image, the daily cleanup task will handle removing
	// unused ones.

	// Update users avatar to newly inserted
	res := s.db.
		Where("id = ?", userId).
		Updates(&entity.User{AvatarID: img.ID})
	if res.Error != nil {
		slog.Error("UploadUserAvatar: Updating the users avatar in db failed!",
			"error", err)
		return entity.Image{}, errors.New("updating user failed")
	}

	return img, nil
}
