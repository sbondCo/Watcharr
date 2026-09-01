package entity

import (
	"time"
)

type Book struct {
	ID int `json:"id" gorm:"primaryKey;autoIncrement"`

	// open library ID of the book
	OLID string `json:"olid" gorm:"uniqueIndex"`

	// List of edition ISBNs, separated by "|"
	ISBN          string  `json:"isbn"`
	Title         string  `json:"title"`
	Storyline     string  `json:"storyline"`
	RatingAverage float64 `json:"ratingAverage"`
	RatingCount   int     `json:"ratingCount"`
	// list of genres, separated by "|"
	Genres string `json:"genres"`

	// TODO: proper database normalization? or just keep it this way, because it's also done this way at other places

	// list of author names, separated by "|"
	AuthorNames string `json:"authorNames"`
	// list of author IDs (same order as author names), separated by "|"
	AuthorIDs string `json:"authorIds"`

	// optional properties
	ReleaseDate *time.Time `json:"releaseDate"`

	// ID to poster image row (cached game cover)
	CoverID *uint  `json:"-"`
	Cover   *Image `json:"cover,omitempty"`
}
