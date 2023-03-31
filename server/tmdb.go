package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
)

type TMDBSearchMultiResponse struct {
	Page         int                      `json:"page"`
	Results      []TMDBSearchMultiResults `json:"results"`
	TotalPages   int                      `json:"total_pages"`
	TotalResults int                      `json:"total_results"`
}

type TMDBSearchMultiResults struct {
	Adult            bool     `json:"adult"`
	BackdropPath     string   `json:"backdrop_path"`
	ID               int      `json:"id"`
	Title            string   `json:"title,omitempty"`
	OriginalLanguage string   `json:"original_language"`
	OriginalTitle    string   `json:"original_title,omitempty"`
	Overview         string   `json:"overview"`
	PosterPath       string   `json:"poster_path"`
	MediaType        string   `json:"media_type"`
	GenreIds         []int64  `json:"genre_ids"`
	Popularity       float64  `json:"popularity"`
	ReleaseDate      string   `json:"release_date,omitempty"`
	Video            bool     `json:"video,omitempty"`
	VoteAverage      float64  `json:"vote_average"`
	VoteCount        int      `json:"vote_count"`
	Name             string   `json:"name,omitempty"`
	OriginalName     string   `json:"original_name,omitempty"`
	FirstAirDate     string   `json:"first_air_date,omitempty"`
	OriginCountry    []string `json:"origin_country,omitempty"`
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
	params.Add("api_key", "d047fa61d926371f277e7a83c9c4ff2c")
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
