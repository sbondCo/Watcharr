package entity

import "github.com/sbondCo/Watcharr/database/dbmodel"

// A remembered choice of which content a name from an import file refers to.
//
// Import files often only give us a title, and a title on its own can match
// more than one thing (or nothing exactly), which means the user is asked to
// pick. Saving that choice means re-importing the same file does not ask
// again.
//
// Mappings are per user, so one users choice can never decide what another
// users import turns into.
type ImportMapping struct {
	dbmodel.GormModel
	// ID of the user this mapping belongs to.
	UserID uint `json:"-" gorm:"not null;uniqueIndex:usernameidx"`
	// The name from the import file, lowercased so lookups are not
	// case sensitive. The original casing is not needed, we only ever
	// match on it.
	Name string `json:"name" gorm:"not null;uniqueIndex:usernameidx"`
	// Content type the name was imported as. Part of the key because the
	// same name can legitimately be both a movie and a show.
	Type string `json:"type" gorm:"not null;uniqueIndex:usernameidx"`
	// The chosen tmdb id (movies and shows).
	TmdbID int `json:"tmdbId"`
	// The chosen igdb id (games).
	IgdbID int `json:"igdbId"`
}
