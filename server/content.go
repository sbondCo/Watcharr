package main

import (
	"errors"
	"log/slog"
	"time"
)

type ContentType string

const (
	MOVIE ContentType = "movie"
	SHOW  ContentType = "tv"
)

// For storing cached content, so we can serve the basic local data for watched list to work
type Content struct {
	ID               int         `json:"id" gorm:"primaryKey;autoIncrement"`
	TmdbID           int         `json:"tmdbId" gorm:"uniqueIndex:contentidtotypeidx;not null"`
	Title            string      `json:"title"`
	PosterPath       string      `json:"poster_path"`
	Overview         string      `json:"overview"`
	Type             ContentType `json:"type" gorm:"uniqueIndex:contentidtotypeidx;not null"`
	ReleaseDate      time.Time   `json:"release_date"`
	Popularity       float32     `json:"popularity"`
	VoteAverage      float32     `json:"vote_average"`
	VoteCount        uint32      `json:"vote_count"`
	ImdbID           string      `json:"imdb_id"`
	Status           string      `json:"status"`
	Budget           uint32      `json:"budget"`
	Revenue          uint32      `json:"revenue"`
	Runtime          uint32      `json:"runtime"`
	NumberOfEpisodes uint32      `json:"numberOfEpisodes"`
	NumberOfSeasons  uint32      `json:"numberOfSeasons"`
}

func searchContent(query string) (TMDBSearchMultiResponse, error) {
	resp := new(TMDBSearchMultiResponse)
	err := tmdbRequest("/search/multi", map[string]string{"query": query, "page": "1"}, &resp)
	if err != nil {
		slog.Error("Failed to complete multi search request!", "error", err.Error())
		return TMDBSearchMultiResponse{}, errors.New("failed to complete multi search request")
	}
	return *resp, nil
}

func movieDetails(id string) (TMDBMovieDetails, error) {
	resp := new(TMDBMovieDetails)
	err := tmdbRequest("/movie/"+id, map[string]string{"append_to_response": "videos,watch/providers"}, &resp)
	if err != nil {
		slog.Error("Failed to complete movie details request!", "error", err.Error())
		return TMDBMovieDetails{}, errors.New("failed to complete movie details request")
	}
	return *resp, nil
}

func movieCredits(id string) (TMDBContentCredits, error) {
	resp := new(TMDBContentCredits)
	err := tmdbRequest("/movie/"+id+"/credits", map[string]string{}, &resp)
	if err != nil {
		slog.Error("Failed to complete movie cast request!", "error", err.Error())
		return TMDBContentCredits{}, errors.New("failed to complete movie cast request")
	}
	return *resp, nil
}

func tvDetails(id string) (TMDBShowDetails, error) {
	resp := new(TMDBShowDetails)
	err := tmdbRequest("/tv/"+id, map[string]string{"append_to_response": "videos,watch/providers"}, &resp)
	if err != nil {
		slog.Error("Failed to complete tv details request!", "error", err.Error())
		return TMDBShowDetails{}, errors.New("failed to complete tv details request")
	}
	return *resp, nil
}

func tvCredits(id string) (TMDBContentCredits, error) {
	resp := new(TMDBContentCredits)
	err := tmdbRequest("/tv/"+id+"/credits", map[string]string{}, &resp)
	if err != nil {
		slog.Error("Failed to complete tv cast request!", err.Error())
		return TMDBContentCredits{}, errors.New("failed to complete tv cast request")
	}
	return *resp, nil
}

func seasonDetails(tvId string, seasonNumber string) (TMDBSeasonDetails, error) {
	resp := new(TMDBSeasonDetails)
	err := tmdbRequest("/tv/"+tvId+"/season/"+seasonNumber, map[string]string{}, &resp)
	if err != nil {
		slog.Error("Failed to complete season details request!", "error", err.Error())
		return TMDBSeasonDetails{}, errors.New("failed to complete season details request")
	}
	return *resp, nil
}

func personDetails(id string) (TMDBPersonDetails, error) {
	resp := new(TMDBPersonDetails)
	err := tmdbRequest("/person/"+id, map[string]string{}, &resp)
	if err != nil {
		slog.Error("Failed to complete person details request!", "error", err.Error())
		return TMDBPersonDetails{}, errors.New("failed to complete person details request")
	}
	return *resp, nil
}

func personCredits(id string) (TMDBPersonCombinedCredits, error) {
	resp := new(TMDBPersonCombinedCredits)
	err := tmdbRequest("/person/"+id+"/combined_credits", map[string]string{}, &resp)
	if err != nil {
		slog.Error("Failed to complete person details request!", "error", err.Error())
		return TMDBPersonCombinedCredits{}, errors.New("failed to complete person details request")
	}
	return *resp, nil
}
