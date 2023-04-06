package main

import (
	"errors"

	"github.com/lib/pq"
)

type ContentType string

const (
	MOVIE ContentType = "movie"
	SHOW  ContentType = "tv"
)

type Content struct {
	Adult            bool           `json:"adult"`
	BackdropPath     string         `json:"backdrop_path"`
	ID               int            `json:"id" gorm:"primaryKey,unique"`
	Title            string         `json:"title,omitempty"`
	OriginalLanguage string         `json:"original_language"`
	OriginalTitle    string         `json:"original_title,omitempty"`
	Overview         string         `json:"overview"`
	PosterPath       string         `json:"poster_path"`
	MediaType        string         `json:"media_type"`
	GenreIds         pq.Int64Array  `json:"genre_ids" gorm:"type:int[]"`
	Popularity       float64        `json:"popularity"`
	ReleaseDate      string         `json:"release_date,omitempty"`
	Video            bool           `json:"video,omitempty"`
	VoteAverage      float64        `json:"vote_average"`
	VoteCount        int            `json:"vote_count"`
	Name             string         `json:"name,omitempty"`
	OriginalName     string         `json:"original_name,omitempty"`
	FirstAirDate     string         `json:"first_air_date,omitempty"`
	OriginCountry    pq.StringArray `json:"origin_country,omitempty" gorm:"type:text[]"`
}

func searchContent(query string) (TMDBSearchMultiResponse, error) {
	resp := new(TMDBSearchMultiResponse)
	err := tmdbRequest("/search/multi", map[string]string{"query": query, "page": "1"}, &resp)
	if err != nil {
		return TMDBSearchMultiResponse{}, errors.New("failed to complete multi search request")
	}
	return *resp, nil
}
