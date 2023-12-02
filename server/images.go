package main

import (
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log/slog"
	"mime/multipart"
	"os"
	"path"
	"time"

	"github.com/buckket/go-blurhash"
	"gorm.io/gorm"
)

// For user uploaded images
type Image struct {
	ID        uint      `gorm:"primarykey" json:"-"`
	CreatedAt time.Time `json:"createdAt"`
	Hash      string    `gorm:"uniqueIndex;not null" json:"-"`
	BlurHash  string    `json:"blurHash"`
	// Path constructable from hash alone, but I can't decide
	// if I should have this or not so I figure it's easier
	// to remove it later than to add it later....... -_-
	Path string `gorm:"not null" json:"path"`
}

// Insert an image into database
func insertImage(db *gorm.DB, hash string, path string, f multipart.File) (Image, error) {
	bh, _ := getBlurHash(f)
	img := Image{
		Hash:     hash,
		Path:     path,
		BlurHash: bh,
	}
	r := db.Where(Image{Hash: hash}).FirstOrCreate(&img)
	if r.Error != nil {
		slog.Error("insertImage firstOrCreate failed!", "error", r.Error)
		return Image{}, errors.New("failed to select or create image")
	}
	return img, nil
}

func getBlurHash(img multipart.File) (string, error) {
	i, _, err := image.Decode(img)
	if err != nil {
		// Handle errors
		slog.Error("getBlurHash decoding image failed", "error", err)
		return "", errors.New("decoding image failed")
	}
	bh, _ := blurhash.Encode(6, 5, i)
	if err != nil {
		// Handle errors
		slog.Error("getBlurHash generation failed", "error", err)
		return "", errors.New("blur hash generation failed")
	}
	slog.Debug("getBlurHash", "hash", bh)
	return bh, nil
}

func cleanupImages(db *gorm.DB) {
	slog.Debug("cleanupImages running")
	var unusedImgs []Image
	// Select images that are not referenced by at least one other row.
	// Currently only used for user avatars, add new tables when used.
	db.Raw(`SELECT *
FROM images
WHERE NOT EXISTS (
	SELECT 1
	FROM users
	WHERE users.avatar_id = images.id
);`).Scan(&unusedImgs)
	slog.Debug("cleanupImages: scanned for unused images", "amount", len(unusedImgs))
	if len(unusedImgs) > 0 {
		for _, v := range unusedImgs {
			slog.Debug("cleanupImages: removing an image", "id", v.ID, "path", v.Path)
			err := db.Transaction(func(tx *gorm.DB) error {
				// Try to delete image from db
				if err := tx.Where("id = ?", v.ID).Delete(&Image{}).Error; err != nil {
					return err
				}
				// hope its ok to do this sorta thing here :skull:
				if err := os.Remove(path.Join(DataPath, v.Path)); err != nil {
					return err
				}
				// commit transaction if no errors
				return nil
			})
			if err != nil {
				slog.Error("cleanupImages: failed to remove image - db row and file kept", "img", v, "error", err)
			} else {
				slog.Debug("cleanupImages: successfully removed unused image.", "id", v.ID)
			}
		}
	}
}
