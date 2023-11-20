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
	// Item already exists so couldn't import (unique constraint hit when adding)
	IMPORT_EXISTS ImportResponseType = "IMPORT_EXISTS"
)

type ImportRequest struct {
	Name   string      `json:"name"`
	TmdbID int         `json:"tmdbId"`
	Type   ContentType `json:"type"`
	Rating int8        `json:"rating"`
}

type ImportResponse struct {
	Type    ImportResponseType       `json:"type"`
	Results []TMDBSearchMultiResults `json:"results"`
	Match   TMDBSearchMultiResults   `json:"match"`
	// On success this will be filled with the new watched entry
	WatchedEntry Watched `json:"watchedEntry"`
}

func importContent(db *gorm.DB, userId uint, ar ImportRequest) (ImportResponse, error) {
	// If tmdbId and type passed in request body
	// we dont need to use a search tmdb request.
	// Retrieve the details directly.
	if ar.TmdbID != 0 && (ar.Type == MOVIE || ar.Type == SHOW) {
		tid := strconv.Itoa(ar.TmdbID)
		if ar.Type == MOVIE {
			cr, err := movieDetails(tid, "", map[string]string{})
			if err != nil {
				return ImportResponse{}, errors.New("movie details request failed")
			}
			slog.Debug("import: by tmdbid of movie", "cr", cr)
			return successfulImport(db, userId, cr.ID, MOVIE, ar.Rating)
		} else if ar.Type == SHOW {
			cr, err := tvDetails(tid, "", map[string]string{})
			if err != nil {
				return ImportResponse{}, errors.New("tv details request failed")
			}
			slog.Debug("import: by tmdbid of tv", "cr", cr)
			return successfulImport(db, userId, cr.ID, SHOW, ar.Rating)
		}
	}
	sr, err := searchContent(ar.Name)
	if err != nil {
		slog.Error("import: content search failed", "error", err)
		return ImportResponse{}, errors.New("Content search failed")
	}
	pMatches := []TMDBSearchMultiResults{}
	for _, r := range sr.Results {
		if r.MediaType != "person" {
			pMatches = append(pMatches, r)
		}
	}
	resLen := len(pMatches)
	slog.Debug("import: potential matches", "num_found", resLen)
	if resLen <= 0 {
		slog.Debug("import: returning IMPORT_NOTFOUND")
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
			slog.Debug("import: importing from perfect match")
			return successfulImport(db, userId, perfectMatch.ID, ContentType(perfectMatch.MediaType), ar.Rating)
		}
		return ImportResponse{Type: IMPORT_MULTI, Results: pMatches}, nil
	} else {
		slog.Debug("import: success.. only found one result")
		return successfulImport(db, userId, pMatches[0].ID, ContentType(pMatches[0].MediaType), ar.Rating)
	}
}

func successfulImport(db *gorm.DB, userId uint, contentId int, contentType ContentType, rating int8) (ImportResponse, error) {
	w, err := addWatched(db, userId, WatchedAddRequest{
		Status:      FINISHED,
		ContentID:   contentId,
		ContentType: contentType,
		Rating:      rating,
	}, IMPORTED_WATCHED)
	if err != nil {
		if err.Error() == "content already on watched list" {
			slog.Error("successfulImport: unique constraint hit.. show must already be on watch list", "error", err)
			return ImportResponse{Type: IMPORT_EXISTS}, nil
		}
		slog.Error("successfulImport: Failed to add content as watched", "error", err)
		return ImportResponse{Type: IMPORT_FAILED}, nil
	}
	return ImportResponse{Type: IMPORT_SUCCESS, WatchedEntry: w}, nil
}
