package domain

import (
	"errors"
	"time"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/util"
)

type ImportContentType string

const (
	ImportContentTypeMovie       ImportContentType = "movie"
	ImportContentTypeShow        ImportContentType = "tv"
	ImportContentTypeShowEpisode ImportContentType = "tv_episode"
	ImportContentTypeGame        ImportContentType = "game"
)

func ImportContentTypeToSearchType(t ImportContentType) SearchType {
	switch t {
	case ImportContentTypeMovie:
		return SearchTypeMovie
	case ImportContentTypeShow:
		return SearchTypeShow
	case ImportContentTypeGame:
		return SearchTypeGame
	}
	// Empty string should be caught as an error.
	return ""
}

type ImportResponseType string

var (
	// Successful import
	IMPORT_SUCCESS ImportResponseType = "IMPORT_SUCCESS"
	// Import failed for reasons user cant fix
	IMPORT_FAILED ImportResponseType = "IMPORT_FAILED"
	// Import query returned multiple results, user must decide
	IMPORT_MULTI ImportResponseType = "IMPORT_MULTI"
	// Import query returned zero results, user must provide more info
	IMPORT_NOTFOUND ImportResponseType = "IMPORT_NOTFOUND"
	// Item already exists so couldn't import (unique constraint hit when adding)
	IMPORT_EXISTS ImportResponseType = "IMPORT_EXISTS"
	// User has chosen to skip this name, now and on any future import
	IMPORT_IGNORED ImportResponseType = "IMPORT_IGNORED"
)

type ImportRequest struct {
	TmdbID int    `json:"tmdbId"`
	ImdbID string `json:"imdbId"`
	IgdbID int    `json:"igdbId"`

	Name             string                  `json:"name"`
	Year             int                     `json:"year"`
	Type             ImportContentType       `json:"type"`
	Rating           float64                 `json:"rating" binding:"max=10"`
	RatingCustomDate *time.Time              `json:"ratingCustomDate"`
	Status           entity.WatchedStatus    `json:"status"`
	Thoughts         string                  `json:"thoughts"`
	DatesWatched     []time.Time             `json:"datesWatched"`
	Activity         []entity.Activity       `json:"activity"`
	WatchedEpisodes  []entity.WatchedEpisode `json:"watchedEpisodes"`
	WatchedSeason    []entity.WatchedSeason  `json:"watchedSeasons"`
	Tags             []TagAddRequest         `json:"tags"`

	// Skip any previously saved mapping for this name, so the user is asked
	// to pick again. Lets a wrong saved choice be corrected, and allows
	// re-picking everything when that is what the user wants.
	IgnoreSavedMatches bool `json:"ignoreSavedMatches"`

	// Don't import this name, now or on any future import. Set when the user
	// gives up on matching an entry, which saves an ignored mapping instead
	// of a match.
	IgnoreThisItem bool `json:"ignoreThisItem"`
}

// Internal struct given to the SuccessfulImport function.
type SuccessfulImportProps struct {
	TmdbID      int
	IgdbID      int
	ContentType util.SupportedMedia
}

func NewSuccessfulImportPropsFromMedia(m *Media) (SuccessfulImportProps, error) {
	p := SuccessfulImportProps{ContentType: m.GetMediaType()}
	switch p.ContentType {
	case util.SupportedMediaMovie, util.SupportedMediaShow:
		p.TmdbID = m.IDs.TMDB
	case util.SupportedMediaGame:
		p.IgdbID = m.IDs.IGDB
	default:
		return p, errors.New("unsupported content type on media")
	}
	return p, nil
}

// Change which content a saved mapping points at.
type ImportMappingUpdateRequest struct {
	TmdbID int `json:"tmdbId"`
	IgdbID int `json:"igdbId"`
	// Content type of the newly picked content. Only needed for mappings
	// saved without one (an ignored name the import file gave no type to),
	// which would otherwise hold an id with nothing to say what it is an id
	// of. Ignored when empty, so an existing type is never lost.
	Type ImportContentType `json:"type"`
}

type ImportResponse struct {
	Type    ImportResponseType `json:"type"`
	Results []Media            `json:"results,omitempty"`
	Match   Media              `json:"match,omitzero"`
	// On success this will be filled with the new watched entry
	WatchedEntry entity.Watched `json:"watchedEntry,omitzero"`
}
