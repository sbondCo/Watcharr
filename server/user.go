package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"log/slog"
	"path"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Public user details for search results
type PublicUser struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	AvatarID uint   `json:"-"`
	Avatar   Image  `json:"avatar"`
	Bio      string `json:"bio,omitempty"`
}

// Private user details, for returning users details to themselves
type PrivateUser struct {
	Username    string   `json:"username"`
	Type        UserType `json:"type"`
	Permissions int      `json:"permissions"`
	AvatarID    uint     `json:"-"`
	Avatar      Image    `json:"avatar"`
	Bio         string   `json:"bio"`
}

type UserBioUpdateRequest struct {
	NewBio string `json:"newBio" binding:"max=128"`
}

// Update user settings
func userUpdate(db *gorm.DB, userId uint, ur UserSettings) (UserSettings, error) {
	slog.Debug("user update request running", "user_id", userId, "ur", ur)
	user := new(User)
	res := db.Where("id = ?", userId).Take(&user)
	if res.Error != nil {
		slog.Error("user update failed", "user_id", userId, "error", res.Error)
		return UserSettings{}, errors.New("failed to retrieve user")
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
	db.Save(&user)
	return UserSettings{
		Private:                  user.Private,
		PrivateThoughts:          user.PrivateThoughts,
		HideSpoilers:             user.HideSpoilers,
		IncludePreviouslyWatched: user.IncludePreviouslyWatched,
		AutomateShowStatuses:     user.AutomateShowStatuses,
		Country:                  user.Country,
	}, nil
}

func userGetSettings(db *gorm.DB, userId uint) (UserSettings, error) {
	slog.Debug("user update request running", "user_id", userId)
	user := new(User)
	res := db.Where("id = ?", userId).Take(&user)
	if res.Error != nil {
		slog.Error("user get failed", "user_id", userId, "error", res.Error)
		return UserSettings{}, errors.New("failed to retrieve user")
	}
	return UserSettings{
		Private:                  user.Private,
		PrivateThoughts:          user.PrivateThoughts,
		HideSpoilers:             user.HideSpoilers,
		IncludePreviouslyWatched: user.IncludePreviouslyWatched,
		AutomateShowStatuses:     user.AutomateShowStatuses,
		Country:                  user.Country,
	}, nil
}

func userSearch(db *gorm.DB, currentUsersId uint, q string) ([]PublicUser, error) {
	slog.Debug("user search request running", "query", q)
	users := new([]PublicUser)
	res := db.Where("private = 0 AND username LIKE ? AND id != ?", "%"+q+"%", currentUsersId).Table("users").Find(&users)
	if res.Error != nil {
		slog.Error("user search failed", "error", res.Error)
		return []PublicUser{}, errors.New("failed to find users")
	}
	return *users, nil
}

func getUserInfo(db *gorm.DB, currentUsersId uint) (PrivateUser, error) {
	slog.Debug("user get info request running")
	user := new(PrivateUser)
	res := db.Where("id = ?", currentUsersId).Table("users").Preload("Avatar").Take(&user)
	if res.Error != nil {
		slog.Error("user get info failed", "error", res.Error)
		return PrivateUser{}, errors.New("failed to find current user")
	}
	return *user, nil
}

// For getting a public user's info, when viewing their list for example
func getUserPublicInfo(db *gorm.DB, userId uint, username string) (PublicUser, error) {
	slog.Debug("user get info request running")
	user := new(PublicUser)
	res := db.Where("private = 0 AND id = ? AND username = ?", userId, username).Table("users").Preload("Avatar").Take(&user)
	if res.Error != nil {
		slog.Error("public user get info failed", "error", res.Error)
		return PublicUser{}, errors.New("failed to find user")
	}
	return *user, nil
}

func userUpdateBio(db *gorm.DB, userId uint, newBio string) error {
	slog.Debug("userUpdateBio request running", "user_id", userId, "newBio", newBio)
	if res := db.Model(&User{}).Where("id = ?", userId).Update("bio", newBio); res.Error != nil {
		slog.Error("userUpdateBio failed", "user_id", userId, "error", res.Error)
		return errors.New("failed to update bio")
	}
	return nil
}

func uploadUserAvatar(c *gin.Context, db *gorm.DB, userId uint) (Image, error) {
	file, err := c.FormFile("avatar")
	if err != nil {
		slog.Error("failed to get file", "error", err)
		return Image{}, errors.New("no file found")
	}

	slog.Debug("an avatar is being uploaded", "name", file.Filename)

	f, _ := file.Open()
	if err := isValidImageType(f); err != nil {
		return Image{}, errors.New("invalid image type")
	}
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err) // TODO nu le fatal
	}
	hs := hex.EncodeToString(h.Sum(nil))

	slog.Debug("image hash calculated", "hash", hs, "first_letter", hs[0:1])

	// Upload the file to specific dst.
	outp := path.Join("img/up/", hs[0:1], hs+filepath.Ext(file.Filename))
	c.SaveUploadedFile(file, path.Join(DataPath, outp))

	_, err = f.Seek(0, 0)
	if err != nil {
		slog.Error("uploadUserAvatar seeking back to start of image failed", "error", err)
	}

	// No need to remove old image, the daily cleanup task will handle removing unused ones.
	var img Image
	err = db.Transaction(func(tx *gorm.DB) error {
		// Insert avatar into db
		img, err = insertImage(db, hs, outp, f)
		if err != nil {
			return err
		}
		if img.ID == 0 {
			return errors.New("image has no id")
		}
		// Update users avatar to newly inserted
		if err := tx.Where("id = ?", userId).Updates(&User{AvatarID: img.ID}).Error; err != nil {
			return err
		}
		// commit transaction if no errors
		return nil
	})
	if err != nil {
		slog.Error("uploadUserAvatar failed!", "error", err)
		return Image{}, errors.New("uploadUserAvatar transaction failed")
	}
	return img, nil
}
