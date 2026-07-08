package image

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"

	"github.com/buckket/go-blurhash"
	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/util"
	"gorm.io/gorm"
)

const (
	defMaxSize        int64 = 10 * util.Mebibyte
	defMaxWidthHeight int64 = 7680
	defMaxPixels      int64 = 10_000_000
)

// TODO now that this file is in the image package it no longer needs to have "image(s)"
// in the name of all the functions..

// Insert an image into database
func Insert(
	db *gorm.DB,
	hash string,
	path string,
	b []byte,
) (entity.Image, error) {
	br := bytes.NewReader(b)
	bh, _ := GetBlurHash(br)
	img := entity.Image{
		Hash:     hash,
		Path:     path,
		BlurHash: bh,
	}
	r := db.Where(entity.Image{Hash: hash}).FirstOrCreate(&img)
	if r.Error != nil {
		slog.Error("insertImage firstOrCreate failed!", "error", r.Error)
		return entity.Image{}, errors.New("failed to select or create image")
	}
	return img, nil
}

func GetBlurHash(img io.Reader) (string, error) {
	i, _, err := image.Decode(img)
	if err != nil {
		// Handle errors
		slog.Error("getBlurHash decoding image failed", "error", err)
		return "", errors.New("decoding image failed")
	}
	bh, err := blurhash.Encode(6, 5, i)
	if err != nil {
		// Handle errors
		slog.Error("getBlurHash generation failed", "error", err)
		return "", errors.New("blur hash generation failed")
	}
	slog.Debug("getBlurHash", "hash", bh)
	return bh, nil
}

func CleanupImages(db *gorm.DB) {
	slog.Info("cleanupImages running")
	var unusedImgs []entity.Image
	// Select images that are not referenced by at least one other row.
	// Currently only used for user avatars, add new tables when used.
	db.Raw(`SELECT *
FROM images
WHERE NOT EXISTS (
	SELECT 1
	FROM users
	WHERE users.avatar_id = images.id
);`).Scan(&unusedImgs)
	slog.Info("cleanupImages: scanned for unused images", "amount", len(unusedImgs))
	if len(unusedImgs) > 0 {
		for _, v := range unusedImgs {
			slog.Debug("cleanupImages: removing an image", "id", v.ID, "path", v.Path)
			err := db.Transaction(func(tx *gorm.DB) error {
				// Try to delete image from db
				if err := tx.Where("id = ?", v.ID).Delete(&entity.Image{}).Error; err != nil {
					return err
				}
				// hope its ok to do this sorta thing here :skull:
				if err := os.Remove(path.Join(config.DataPath, v.Path)); err != nil {
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

// Validate the image safely.
// Re-encodes the image in our own desired format, which helps verify the file
// is a valid image and any undesirable data (eg think xss; appended html,
// exit, etc) is not kept in the final image file we store.
// We currently prefer jpg for the format we re-encode to, which also helps us
// save storage space (adds compression, which the image we are validating
// could lack or have a higher quality setting, etc).
// Returns the new re-encoded image data, file extension, and error.
// NOTE: Never use a user-set file extension (always validate we allow it).
func Validate(
	b []byte,
) ([]byte, string, error) {
	br := bytes.NewReader(b)

	// Check image header for config/format.
	cfg, format, err := image.DecodeConfig(br)
	if err != nil {
		slog.Error("Validate: Failed to DecodeConfig", "error", err)
		return []byte{}, "", errors.New("invalid or bad image")
	}
	slog.Debug("Validate", "cfg", cfg, "format", format)
	if int64(cfg.Width) > defMaxWidthHeight ||
		int64(cfg.Height) > defMaxWidthHeight {
		return []byte{}, "", errors.New("dimensions too large")
	}
	// Protect against images that max out the allowed width/height.
	// If we assume each pixel is 4 bytes, someone maxing out 8000x8000 would
	// mean ~235mb of data we need to decode into memory (i think!), but instead
	// of limiting the max width/height values too much, we can limit the max
	// amount of pixels to restrict the max size of the img pixels we'd allow.
	// Then weird aspect ratios are still allowed.
	// I'm definitely over-engineering this feature for a self-hosted movie list app lol.
	if int64(cfg.Width)*int64(cfg.Height) > defMaxPixels {
		return []byte{}, "", errors.New("i can't handle all those pixels")
	}

	// Seek back, we are reading again below for decode.
	if _, err = br.Seek(0, 0); err != nil {
		slog.Error("Validate: Seeking reader to start failed", "error", err)
		return []byte{}, "", err
	}

	// Decode image.
	// This should catch any malformed image files.
	var img image.Image
	switch format {
	case "png":
		img, err = png.Decode(br)
	case "jpeg":
		img, err = jpeg.Decode(br)
	case "gif":
		img, err = gif.Decode(br)
	default:
		return []byte{}, "", errors.New("unsupported image type")
	}
	if err != nil {
		slog.Error("full image decode failed", "error", err, "format", format)
		return []byte{}, "", errors.New("invalid or corrupt image")
	}

	// Re-encode the image from our decoded data (any extra included, possibly
	// malicious data not part of the image should be gone now).
	// exif data, etc should also be gone now too which is good.
	// We just re-encode as jpeg right now, but if we wanted to, in the future
	// it's possible to encode different formats based on if we want to perserve
	// transparency from png, etc. OR maybe using webp will be easier and we
	// can just use that format since it supports transparency and animations.
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 75}); err != nil {
		slog.Error("Validate: Failed to encode jpeg", "error", err)
		return []byte{}, "", errors.New("failed to encode image")
	}
	return buf.Bytes(), ".jpg", nil
}

// Creates the image file on disk.
func save(b []byte, subPath string) (string, string, error) {
	br := bytes.NewReader(b)
	// First we get a hash of the files contents.
	// Using the hash for filename has the benefit of us not storing duplicate
	// files just because their filename provided to us is different.
	h := sha256.New()
	if _, err := io.Copy(h, br); err != nil {
		slog.Error("save: Copy failed!", "error", err)
		return "", "", errors.New("copy failed")
	}
	hs := hex.EncodeToString(h.Sum(nil))
	slog.Debug("save: image hash calculated",
		"hash", hs,
		"first_letter", hs[0:1])

	// Validate the file.
	// We always validate, even from "trusted sources!"
	b, ext, err := Validate(b)
	if err != nil {
		slog.Error("save: Validate failed!", "error", err)
		return "", "", err
	}
	br = bytes.NewReader(b)

	// Create paths for file.
	imgPath := path.Join(
		// Always outputs to `img/` dir.
		"img/",
		// Any sub path for separating images.
		subPath,
		// Sub-separate images by the starting character of their hash.
		hs[0:1],
		// File name is whole hash then the file extension.
		hs+ext)
	fullOutPath := path.Join(config.DataPath, imgPath)
	slog.Debug("save: Built path", "path", imgPath)

	// Save file
	err = os.MkdirAll(path.Dir(fullOutPath), 0764)
	if err != nil {
		return "", "", err
	}
	out, err := os.Create(fullOutPath)
	if err != nil {
		return "", "", err
	}
	defer out.Close()
	_, err = io.Copy(out, br)
	if err != nil {
		return "", "", err
	}

	return hs, imgPath, nil
}

// Download an image from `url` and insert it.
func DownloadAndInsertFromUrl(
	db *gorm.DB,
	url string,
	imgSubPath string,
) (entity.Image, error) {
	slog.Debug("DownloadAndInsertFromUrl: Running.", "url", url)

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return entity.Image{}, err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return entity.Image{}, fmt.Errorf("bad status: %s", resp.Status)
	}

	return DownloadAndInsert(db, resp.Body, imgSubPath)
}

func DownloadAndInsert(
	db *gorm.DB,
	r io.Reader,
	imgSubPath string,
) (entity.Image, error) {
	slog.Debug("DownloadAndInsert: Running.")

	// Read all into memory.
	b, err := util.LimitedReadAll(r, defMaxSize)
	if err != nil {
		slog.Error("DownloadAndInsert: Failed to read response!", "error", err)
		return entity.Image{}, err
	}

	// Save the file.
	imgHash, imgPath, err := save(b, imgSubPath)
	if err != nil {
		slog.Error("DownloadAndInsert: Failed to save file!", "error", err)
		return entity.Image{}, err
	}

	// Insert image into db.
	imge, err := Insert(db, imgHash, imgPath, b)
	if err != nil {
		slog.Error("DownloadAndInsert: Insert into db failed!", "error", err)
		return entity.Image{}, err
	}

	return imge, nil
}
