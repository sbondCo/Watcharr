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
}
