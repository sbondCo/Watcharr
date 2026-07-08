// testutil is for testing code that we want to reuse for tests.
package testutil

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/sbondCo/Watcharr/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SetupLogging will configure the default slog logger.
// Since our main app uses `slog`, this is useful for getting
// debug logs to show, or hiding all, etc.
// Controlled by env var `WTEST_LOG_LEVEL` (accepts: `debug` or `error`), if not
// set, default Info log level is used.
func SetupLogging() {
	level := slog.LevelInfo
	switch os.Getenv("WTEST_LOG_LEVEL") {
	case "debug":
		level = slog.LevelDebug
	case "error":
		level = slog.LevelError
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(
		os.Stdout, &slog.HandlerOptions{Level: level})))
}

// Setup a fresh database for testing.
// Exits test by using t.Fatalf if something fails.
func SetupDB(t *testing.T) *gorm.DB {
	t.Helper()

	// Open our test db.
	// Note: Could have used inmemory db, but it breaks our WAL migration and
	// errors out, and I don't wanna mess with prod code simply so I can use
	// an inmem db for testing, so we make a temporary file db.
	db, err := gorm.Open(
		sqlite.Open(filepath.Join(t.TempDir(), "test-watcharr.db")),
		&gorm.Config{TranslateError: true},
	)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	// Setup the db same as we do for prod.
	if err := database.Setup(db); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}

	return db
}
