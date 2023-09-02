package main

import (
	"errors"

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
	Name string `json:"name"`
	Year string `json:"year"`
}

type ImportResponse struct {
	Type    ImportResponseType       `json:"type"`
	Results []TMDBSearchMultiResults `json:"results"`
}

// TODO
// - add slogging
// - filter people out
// - use year if provided in request body
func importContent(db *gorm.DB, userId uint, ar ImportRequest) (ImportResponse, error) {
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
