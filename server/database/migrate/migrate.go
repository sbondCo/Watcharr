package migrate

import (
	"log/slog"
	"time"

	"gorm.io/gorm"
)

type Migration struct {
	// ID of migration, stick to YYYYMMDDHHMM.
	ID string
	// Apply migration func.
	Up func(tx *gorm.DB) error
	// When `true`, the migration is not run inside of a transaction.
	// You should only use this when is required by sqlite that the command
	// we need to run for eg cannot be ran from within a transaction!
	// YOU SHOULD ENSURE YOU ONLY RUN ONE COMMAND PER MIGRATION WHEN USING
	// THIS WITH `TRUE` TO AVOID BEING LEFT IN A BAD OR INCOMPLETE STATE!!!
	//
	// ALSO: All statements used with this should take into account that since,
	// it isn't inside of a transaction, it's possible the migration succeeds,
	// but creating the record of it doesn't. If a user starts the server again
	// after we error in this case, the migration will run again, so it must
	// not break data integrity or make any assumptions of it being the first
	// time running!
	UNSAFE bool
}

// Start our migrations.
// NOTE: This is only to be ran after GORM's AutoMigrate.
func Now(db *gorm.DB) error {
	slog.Info("Starting migrations.")

	for _, mig := range migrations {
		slog.Debug("Processing migration.", "id", mig.ID)

		migRecord := MigrationRecord{ID: mig.ID}

		// First ensure that the migration hasn't already been applied.
		var alreadyApplied int64
		res := db.
			Model(&MigrationRecord{}).
			Where(&migRecord).
			Count(&alreadyApplied)
		if res.Error != nil {
			slog.Error("already applied check failed!")
			return res.Error
		}
		if alreadyApplied > 0 {
			// If record exists in our table, then migration was applied
			// already, so skip processing it.
			slog.Debug("Migration already applied.", "id", mig.ID)
			continue
		}

		slog.Info("Migration has NOT been applied before.. applying.",
			"id", mig.ID)

		// Timing the migration.
		timeBeforeMig := time.Now()

		// Apply the migration.
		if mig.UNSAFE {
			// Unsafe migrations are not ran inside of a transaction
			// and are only used when required by sqlite engine.
			if err := mig.Up(db); err != nil {
				slog.Error("Migration failed!", "id", mig.ID, "error", err)
				return err
			}
			// Record the migration record.
			if res := db.Create(&migRecord); res.Error != nil {
				slog.Error("Unsafe migration succeeded, but we failed to create the record of it!",
					"id", mig.ID, "error", res.Error)
				return res.Error
			}
		} else {
			// Migrations go through a transaction wrapper.
			err := db.Transaction(func(tx *gorm.DB) error {
				if err := mig.Up(tx); err != nil {
					// Errored.. rollback any changes made.
					return err
				}
				// Migration succeeded.. record it.
				// If the Create succeeds, all will be committed.
				return tx.Create(&migRecord).Error
			})
			if err != nil {
				// If any migration fails, we return here.
				slog.Error("Migration failed!", "id", mig.ID, "error", err)
				return err
			}
		}

		slog.Info("Migration applied successfully.",
			"id", mig.ID,
			"duration", time.Since(timeBeforeMig))
	}

	slog.Info("Done processing all migrations.")

	return nil
}
