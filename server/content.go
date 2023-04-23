package main

import (
	"errors"
)

type ContentType string

const (
	MOVIE ContentType = "movie"
	SHOW  ContentType = "tv"
)

// For storing cached content, so we can serve the basic local data for watched list to work
type Content struct {
	ID         int         `json:"id" gorm:"primaryKey;autoIncrement"`
	TmdbID     int         `json:"tmdbId" gorm:"uniqueIndex:contentidtotypeidx;not null"`
	Title      string      `json:"title"`
	PosterPath string      `json:"poster_path"`
	Overview   string      `json:"overview"`
	Type       ContentType `json:"type" gorm:"uniqueIndex:contentidtotypeidx;not null"`
}

func searchContent(query string) (TMDBSearchMultiResponse, error) {
	resp := new(TMDBSearchMultiResponse)
	err := tmdbRequest("/search/multi", map[string]string{"query": query, "page": "1"}, &resp)
	if err != nil {
		println("Failed to complete multi search request!", err.Error())
		return TMDBSearchMultiResponse{}, errors.New("failed to complete multi search request")
	}
	return *resp, nil
}

func movieDetails(id string) (TMDBMovieDetails, error) {
	resp := new(TMDBMovieDetails)
	err := tmdbRequest("/movie/"+id, map[string]string{}, &resp)
	if err != nil {
		println("Failed to complete movie details request!", err.Error())
		return TMDBMovieDetails{}, errors.New("failed to complete movie details request")
	}
	return *resp, nil
}

func tvDetails(id string) (TMDBShowDetails, error) {
	resp := new(TMDBShowDetails)
	err := tmdbRequest("/tv/"+id, map[string]string{}, &resp)
	if err != nil {
		println("Failed to complete tv details request!", err.Error())
		return TMDBShowDetails{}, errors.New("failed to complete tv details request")
	}
	return *resp, nil
}
