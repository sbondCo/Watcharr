package main

import (
	"errors"
	"log/slog"
	"strconv"

	"gorm.io/gorm"
)

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
)

type ImportRequest struct {
	Name   string      `json:"name"`
	Year   string      `json:"year"`
	TmdbID int         `json:"tmdbId"`
	Type   ContentType `json:"type"`
}

type ImportResponse struct {
	Type    ImportResponseType       `json:"type"`
	Results []TMDBSearchMultiResults `json:"results"`
}

// TODO
// - add slogging
// - filter people out
// - use year if provided in request body
// - actually import the content
func importContent(db *gorm.DB, userId uint, ar ImportRequest) (ImportResponse, error) {
	// If tmdbId and type passed in request body
	// we dont need to use a search tmdb request.
	// Retrieve the details directly.
	if ar.TmdbID != 0 && (ar.Type == MOVIE || ar.Type == SHOW) {
		tid := strconv.Itoa(ar.TmdbID)
		if ar.Type == MOVIE {
			cr, err := movieDetails(tid, "")
			if err != nil {
				return ImportResponse{}, errors.New("movie details request failed")
			}
			slog.Info("", "cr", cr)
			return ImportResponse{Type: IMPORT_SUCCESS}, nil
		} else if ar.Type == SHOW {
			cr, err := tvDetails(tid, "")
			if err != nil {
				return ImportResponse{}, errors.New("tv details request failed")
			}
			slog.Info("", "cr", cr)
			return ImportResponse{Type: IMPORT_SUCCESS}, nil
		}
	}
	sr, err := searchContent(ar.Name)
	if err != nil {
		return ImportResponse{}, errors.New("Content search failed")
	}
	resLen := len(sr.Results)
	if resLen <= 0 {
		return ImportResponse{Type: IMPORT_NOTFOUND}, nil
	} else if resLen > 1 {
		return ImportResponse{Type: IMPORT_MULTI, Results: sr.Results}, nil
	} else {
		return ImportResponse{Type: IMPORT_SUCCESS}, nil
	}
}
