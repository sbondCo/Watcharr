package image

import (
	"bytes"
	"errors"
	"image"
	"io"
	"log/slog"
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
