package main

import (
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log/slog"
	"mime/multipart"
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
