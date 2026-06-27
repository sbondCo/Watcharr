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
}
