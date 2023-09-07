package main

import (
	"errors"
	"log/slog"
	"strconv"
	"strings"

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
	TmdbID int         `json:"tmdbId"`
	Type   ContentType `json:"type"`
}

type ImportResponse struct {
	Type    ImportResponseType       `json:"type"`
	Results []TMDBSearchMultiResults `json:"results"`
	Match   TMDBSearchMultiResults   `json:"match"`
}

// TODO
// - add slogging
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
			slog.Debug("import: by tmdbid of movie", "cr", cr)
			return ImportResponse{Type: IMPORT_SUCCESS, Match: TMDBSearchMultiResults{ID: cr.ID, Title: cr.Title, ReleaseDate: cr.ReleaseDate, MediaType: string(MOVIE)}}, nil
		} else if ar.Type == SHOW {
			cr, err := tvDetails(tid, "")
			if err != nil {
				return ImportResponse{}, errors.New("tv details request failed")
			}
			slog.Debug("import: by tmdbid of tv", "cr", cr)
			return ImportResponse{Type: IMPORT_SUCCESS, Match: TMDBSearchMultiResults{ID: cr.ID, Name: cr.Name, FirstAirDate: cr.FirstAirDate, MediaType: string(SHOW)}}, nil
		}
	}
	sr, err := searchContent(ar.Name)
	if err != nil {
		return ImportResponse{}, errors.New("Content search failed")
	}
	pMatches := []TMDBSearchMultiResults{}
	for _, r := range sr.Results {
		if r.MediaType != "person" {
			pMatches = append(pMatches, r)
		}
	}
	resLen := len(pMatches)
	if resLen <= 0 {
		return ImportResponse{Type: IMPORT_NOTFOUND}, nil
	} else if resLen > 1 {
		slog.Debug("import: multiple results found")
		// If there are multiple responses, but only one item
		// from the results is a 100% match for the imported
		// items name, then consider successful match with that.
		var perfectMatch TMDBSearchMultiResults
		for _, r := range pMatches {
			itemName := r.Name
			if itemName == "" {
				itemName = r.Title
			}
			if strings.EqualFold(itemName, ar.Name) {
				slog.Debug("import: multiple results processing: found a perfectMatch", "match", r)
				if perfectMatch.ID != 0 {
					// If perfect match has been set before..
					// quit looking and just show all results.
					slog.Debug("import: multiple results processing: Second perfectMatch found.. returning all results")
					return ImportResponse{Type: IMPORT_MULTI, Results: pMatches}, nil
				}
				perfectMatch = r
			}
		}
		// If one perfect match found, import it
		if perfectMatch.ID != 0 {
			return ImportResponse{Type: IMPORT_SUCCESS, Match: perfectMatch}, nil
		}
		return ImportResponse{Type: IMPORT_MULTI, Results: pMatches}, nil
	} else {
		return ImportResponse{Type: IMPORT_SUCCESS, Match: pMatches[0]}, nil
	}
}
