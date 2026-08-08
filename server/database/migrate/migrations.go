package migrate

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/sbondCo/Watcharr/database/entity"
	"gorm.io/gorm"
)

// NOTE: For obvious reasons, once a migration is created and in production,
// it is set in stone, so there should be almost no reason to change an existing
// migration, create a new one instead!
// If it's not obvious, changing an existing migration won't apply for people
// who already have applied it and only apply for people who haven't yet,
// so we are risking splitting the consistency of everyones databases as a
// whole. I can't forsee any circumstance that would require doing so..
// We should still delete older migrations once they can't safely run on new
// schemas at all in the future (we can only truly support upgrading Watcharr
// to the next and not skipping version in between.).

var migrations = []Migration{
	{
		// Backfilling `plays` data from users Activity.
		// (we have just created the 'count_as_play' column, instead of starting
		// existing data from 0 plays, we can check what existing activities
		// should count, and count them).
		ID: "202603201715_0001",
		Up: func(tx *gorm.DB) error {
			migID := "202603201715_0001"
			slog.Info("Migration is starting.", "mig", migID)
			// For ADDED_WATCHED and STATUS_CHANGED activities where the data
			// holds something saying FINISHED somewhere, count as a play.
			res := tx.
				Model(&entity.Activity{}).
				Where(
					`type IN ? AND data LIKE "%FINISHED%"`,
					[]entity.ActivityType{
						entity.ADDED_WATCHED,
						entity.STATUS_CHANGED,
					},
				).
				Update("count_as_play", 1)
			if res.Error != nil {
				slog.Error("First step failed!", "mig", migID,
					"error", res.Error)
				return res.Error
			}
			slog.Info("First step succeeded, continuing.", "mig", migID)
			// For IMPORTED_ADDED_WATCHED* activities, we don't need to check
			// data (since there isn't any). We know these should always count
			// as a play, so count them.
			res = tx.
				Model(&entity.Activity{}).
				Where("type IN ?", []entity.ActivityType{
					entity.IMPORTED_ADDED_WATCHED,
					entity.IMPORTED_ADDED_WATCHED_JF,
					entity.IMPORTED_ADDED_WATCHED_PLEX,
				}).
				Update("count_as_play", 1)
			if res.Error != nil {
				slog.Error("Second step failed!", "mig", migID,
					"error", res.Error)
				return res.Error
			}
			slog.Info("Second step succeeded, continuing.", "mig", migID)
			return nil
		},
	},
	{
		// Dropping `deleted_at` columns for `watched_seasons` and
		// `watched_episodes` tables since we do not use them.
		ID: "202604142234_0002",
		Up: func(tx *gorm.DB) error {
			migID := "202604142234_0002"
			slog.Info("Migration is starting.", "mig", migID)

			// Drop deleted_at for watched_seasons
			err := tx.Migrator().DropColumn(&entity.WatchedSeason{}, "deleted_at")
			if err != nil {
				slog.Error("watched_seasons migration failed!", "mig", migID)
				return err
			}
			slog.Info("watched_seasons migration succeeded.", "mig", migID)

			// Drop deleted_at for watched_episodes
			err = tx.Migrator().DropColumn(&entity.WatchedEpisode{}, "deleted_at")
			if err != nil {
				slog.Error("watched_episodes migration failed!", "mig", migID)
				return err
			}
			slog.Info("watched_episodes migration succeeded.", "mig", migID)

			slog.Info("Migration complete.", "mig", migID)
			return nil
		},
	},
	{
		// Moving to using WAL journal_mode for our sqlite database, which
		// should grant us improvements in all areas.
		// https://sqlite.org/pragma.html#pragma_journal_mode
		ID: "202604162229_0003",
		// NOTE: We are using `unsafe` so that this migration isn't ran inside
		// of a transaction (can't change into WAL from within one), so we
		// MUST ENSURE we are only doing one thing!
		UNSAFE: true,
		Up: func(db *gorm.DB) error {
			migID := "202604162229_0003"
			slog.Info("Migration is starting.", "mig", migID)

			var mode string
			res := db.Raw("PRAGMA journal_mode=WAL").Scan(&mode)
			if res.Error != nil {
				slog.Error("Setting journal_mode=WAL failed!", "mig", migID)
				return res.Error
			}
			// Setting journal_mode might not return an error if it fails,
			// it always returns the current journal_mode of the db, which
			// will be WAL if it succeeds OR the "old" journal_mode if it
			// wasn't changed.
			// If the mode returned isn't WAL, then something has failed, so
			// we'll error to stop here and prevent the migration record from
			// being created, allowing the user to try again.
			// Note: I was able to test this code by opening the db like this
			// `sqlite.Open("file:data/watcharr.db?immutable=true")` and
			// commenting out other code so we get right to this migration
			// without failing at AutoMigration, etc.
			slog.Info("journal_mode response.", "mode", mode)
			if strings.ToLower(mode) != "wal" {
				slog.Error("Setting journal_mode=WAL failed silently!")
				return errors.New("Database is not in WAL mode after setting journal_mode=WAL")
			}
			slog.Info("WAL journal_mode migration succeeded.", "mig", migID)

			slog.Info("Migration complete.", "mig", migID)
			return nil
		},
	},
	{
		// Migrating third_party_{id,auth} data from Users table to the
		// user_services table. Only Jellyfin users have their auth using the
		// third_party columns so we can safely assume this.
		ID: "202608070037_0004",
		Up: func(tx *gorm.DB) error {
			migID := "202608070037_0004"
			slog.Info("Migration is starting.", "mig", migID)

			// If the columns to migrate dont exist in this db, we can assume
			// this migration has succeeded before OR this is a fresh db that
			// doesn't need this migration to run.
			if !tx.Migrator().HasColumn(&entity.User{}, "third_party_id") ||
				!tx.Migrator().HasColumn(&entity.User{}, "third_party_auth") {
				slog.Info("Current user table doesn't have third_party_id or third_party_auth columns. Skipping.",
					"mig", migID)
				return nil
			}

			// Migrate the data to user_services table.
			err := tx.Exec(`
				INSERT INTO user_services (user_id, name, client_id, auth_token, created_at, updated_at)
				SELECT id, 'jellyfin', third_party_id, third_party_auth, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
				FROM users
				WHERE third_party_id IS NOT NULL AND third_party_id != ''
			`).Error
			if err != nil {
				slog.Error("migration query failed!", "mig", migID)
				return err
			}

			slog.Info("Migration complete.", "mig", migID)
			return nil
		},
	},
}
