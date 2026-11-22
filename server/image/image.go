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
	"github.com/sbondCo/Watcharr/database"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/util"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

// Removes images from db and disk when they become unused by any table.
func CleanupImages(db *gorm.DB) {
	slog.Info("CleanupImages: Running.")

	// Get list of columns referencing images.
	refs, err := database.ForeignKeyReferences(db, "images")
	if err != nil {
		slog.Error("CleanupImages: Failed to get references.", "error", err)
		return
	}
	if len(refs) <= 0 {
		// This check is important, otherwise no WHERE clauses would get added
		// to the query below when no refs (causing all images to be returned).
		// This shouldn't happen. If it does, something has gone very wrong.
		slog.Error("CleanupImages: No references returned! Something is wrong.")
		return
	}

	// Build query for unused images by checking if any are not used by any `refs`.
	var unusedImgs []entity.Image
	var whereAdded bool
	q := db.Model(&entity.Image{})
	for _, v := range refs {
		q = q.Where(
			"NOT EXISTS (SELECT 1 FROM ? AS r WHERE ? = ?)",
			clause.Table{Name: v.Table},
			clause.Column{Table: "r", Name: v.Column},
			clause.Column{Table: "images", Name: v.RefColumn},
		)
		whereAdded = true // feelin preddy paranoid rn
	}
	if !whereAdded {
		slog.Error("CleanupImages: No WHEREs added to query! Quitting to avoid catastrophic failure.")
		return
	}
	err = q.Find(&unusedImgs).Error
	if err != nil {
		slog.Error("CleanupImages: Unused images query failed!", "error", err)
		return
	}

	if len(unusedImgs) <= 0 {
		slog.Debug("CleanupImages: No unused images found.")
		return
	}
	slog.Info("CleanupImages: Found unused images.", "amount", len(unusedImgs))

	for _, v := range unusedImgs {
		slog.Debug("CleanupImages: Removing an image", "id", v.ID, "path", v.Path)

		if err := db.Where("id = ?", v.ID).Delete(&entity.Image{}).Error; err != nil {
			slog.Error("CleanupImages: Failed to remove row!",
				"id", v.ID, "error", err)
			continue
		}
		if err := os.Remove(path.Join(config.DataPath, v.Path)); err != nil {
			slog.Error("CleanupImages: Failed to remove file!", "error", err)
			continue
		}
	}

	// TODO There's a possible failure mode where the row is deleted, but the
	// file on disk wasn't deleted for any reason. Before the task finishes (or
	// maybe have another less occurring task) we could scan images on disk and
	// if they aren't referenced in the db, try to delete them again.
}
