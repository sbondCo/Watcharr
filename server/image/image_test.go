package image

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/internal/testutil"
)

func TestDownloadAndInsertFromUrl(t *testing.T) {
	testutil.SetupLogging()
	db := testutil.SetupDB(t)

	i, err := NewSaver(db, "test", ValidateOptions{}).
		DownloadAndInsertFromUrl(
			"https://github.com/sbondCo/Watcharr/raw/dev/screenshot/homepage.png")
	if err != nil {
		t.Fatalf("DownloadAndInsert call failed: %v", err)
	}

	if i.ID == 0 || i.Path == "" || i.BlurHash == "" {
		t.Fatal("returned entity.Image doesn't have certain fields!",
			"id", i.ID, "path", i.Path, "blurhash", i.BlurHash)
	}

	fullImgDataPath := path.Join(config.DataPath, i.Path)

	// Verify file exists and looks right.
	if fi, err := os.Stat(fullImgDataPath); err != nil {
		t.Fatalf("os.stat failed %v", err)
	} else if fi.Size() <= 1 {
		t.Fatalf("image file size doesn't seem right: %v", fi.Size())
	} else if filepath.Ext(fi.Name()) != ".jpg" {
		t.Fatalf("image file name doesn't have .jpg ext: %s", fi.Name())
	}

	if err := os.Remove(fullImgDataPath); err != nil {
		// Not a fatal error because this is extra logic only for testing,
		// but failing test because something in the real logic might possibly
		// have something to do with it failing and we should probably know.
		t.Errorf("removing the image file errored: %v", err)
	}

	// Verify image is in db
	var c int64
	if res := db.
		Model(&entity.Image{}).
		Where(&entity.Image{ID: i.ID}).
		Count(&c); res.Error != nil {
		t.Fatalf("verification query failed: %v", res.Error)
	}
	if c != 1 {
		t.Fatalf("count of images in db doesn't look right: %v", c)
	}
}
