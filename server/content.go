package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"gorm.io/gorm"
)

type Content struct {
	gorm.Model

	ID   int    `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}

type TMDBSearchMultiResponse struct {
	Page    int `json:"page"`
	Results []struct {
		Adult            bool     `json:"adult"`
		BackdropPath     string   `json:"backdrop_path"`
		ID               int      `json:"id"`
		Title            string   `json:"title,omitempty"`
		OriginalLanguage string   `json:"original_language"`
		OriginalTitle    string   `json:"original_title,omitempty"`
		Overview         string   `json:"overview"`
		PosterPath       string   `json:"poster_path"`
		MediaType        string   `json:"media_type"`
		GenreIds         []int    `json:"genre_ids"`
		Popularity       float64  `json:"popularity"`
		ReleaseDate      string   `json:"release_date,omitempty"`
		Video            bool     `json:"video,omitempty"`
		VoteAverage      float64  `json:"vote_average"`
		VoteCount        int      `json:"vote_count"`
		Name             string   `json:"name,omitempty"`
		OriginalName     string   `json:"original_name,omitempty"`
		FirstAirDate     string   `json:"first_air_date,omitempty"`
		OriginCountry    []string `json:"origin_country,omitempty"`
	} `json:"results"`
	TotalPages   int `json:"total_pages"`
	TotalResults int `json:"total_results"`
}

func tmdbRequest(ep string, p map[string]string, resp interface{}) error {
	base, err := url.Parse("https://api.themoviedb.org/3")
	if err != nil {
		return errors.New("failed to parse api uri")
	}

	// Path params
	base.Path += ep

	// Query params
	params := url.Values{}
	params.Add("api_key", "")
	params.Add("language", "en-US")
	for k, v := range p {
		params.Add(k, v)
	}

	// Add params to url
	base.RawQuery = params.Encode()

	// Run get request
	res, err := http.Get(base.String())
	if err != nil {
		return err
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(body), &resp)
	if err != nil {
		return err
	}
	return nil
}

func searchContent(query string) (TMDBSearchMultiResponse, error) {
	resp := new(TMDBSearchMultiResponse)
	err := tmdbRequest("/search/multi", map[string]string{"query": query, "page": "1"}, &resp)
	if err != nil {
		return TMDBSearchMultiResponse{}, errors.New("failed to complete multi search request")
	}
	return *resp, nil
}
