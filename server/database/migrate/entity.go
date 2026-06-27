package migrate

import "time"

// Record of all our applied migrations.
type MigrationRecord struct {
	// Migration ID
	ID string `gorm:"primarykey"`
	// When migration was applied on this db.
	CreatedAt time.Time
}
