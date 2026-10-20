package database

import (
	"log/slog"

	"gorm.io/gorm"
)

type fkRef struct {
	// The referencing table.
	Table string
	// The column in `Table` that holds the reference.
	Column string
	// The column in the referenceD table that `Column` is referencing.
	RefColumn string
}

// Get list of columns in the db REFERENCESing `table`.
func ForeignKeyReferences(db *gorm.DB, table string) ([]fkRef, error) {
	var refs []fkRef
	err := db.
		Raw(`
			SELECT
				s.name AS "table",
				f."from" AS "column",
				COALESCE(f."to", 'id') AS "ref_column"
			FROM sqlite_schema s, pragma_foreign_key_list(s.name) f
			WHERE s.type = 'table' AND f."table" = ? COLLATE NOCASE
		`, table).
		Scan(&refs).
		Error
	slog.Debug("ForeignKeyReferences", "refs", refs, "err", err)
	return refs, err
}
