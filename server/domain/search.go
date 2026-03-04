package domain

import (
	"github.com/go-playground/validator/v10"
	"github.com/sbondCo/Watcharr/util"
)

type SearchType string

const (
	// Search for **all available media types**.
	SearchTypeMulti SearchType = "multi"
	// Search for a **movie**.
	SearchTypeMovie SearchType = "movie"
	// Search for a **show**.
	SearchTypeShow SearchType = "show"
	// Search for a **person** (actor).
	SearchTypePerson SearchType = "person"
	// Search for a **game**.
	SearchTypeGame SearchType = "game"
)

type SearchRequest struct {
	// The type of content we are searching for.
	// SearchTypeMulti encompasses all types of media in the results.
	Type SearchType `form:"type" binding:"validsearchtype"`
	// The search term.
	Query string `form:"query"`
	// Prefer a search of our watched list first and return that if any results.
	PreferMyList bool `form:"preferMyList"`
}

type SearchResponse struct {
	util.PaginationResponse[Media, SearchResponseMeta]
}

// Search response metadata
type SearchResponseMeta struct {
	// When true, indicates that results were found from our list.
	FromMyList bool `json:"fromMyList,omitempty"`
}

var ValidSearchType validator.Func = func(fl validator.FieldLevel) bool {
	st, ok := fl.Field().Interface().(SearchType)
	if ok {
		switch st {
		case SearchTypeMulti,
			SearchTypeMovie,
			SearchTypeShow,
			SearchTypePerson,
			SearchTypeGame:
			return true
		}
	}
	return false
}
