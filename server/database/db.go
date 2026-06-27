package database

import (
	"log/slog"
	"path"

	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/database/migrate"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Create a new database connection.
// Also runs migrations, etc, before returning connection.
// Any error returned from this func should always make our app Exit (caller
// handled).
func New() (*gorm.DB, error) {
	slog.Info("New: Opening new database connection")
	// Open the database.
	db, err := gorm.Open(
		sqlite.Open(path.Join(config.DataPath, "watcharr.db")),
		&gorm.Config{TranslateError: true},
	)
	if err != nil {
		slog.Error("New: Opening database failed.")
		return nil, err
	}
		return nil, err
	}
	// Perform auto migration.
	slog.Info("New: AutoMigrating")
	err = db.AutoMigrate(
		&migrate.MigrationRecord{},
		&entity.User{},
		&entity.UserServices{},
		&entity.Content{},
		&entity.Watched{},
		&entity.WatchedSeason{},
		&entity.WatchedEpisode{},
		&entity.Activity{},
		&entity.Token{},
		&entity.Follow{},
		&entity.Image{},
		&entity.Game{},
		&entity.ArrRequest{},
		&entity.Tag{},
	)
	if err != nil {
		slog.Error("New: Auto migration failed.")
		return nil, err
	}
	slog.Info("New: AutoMigrated")
	// Perform our manual migrations.
	if err := migrate.Now(db); err != nil {
		slog.Error("New: Manual migrations failed.", "error", err)
		return nil, err
	}
		return nil, err
	}
	return db, nil
}
