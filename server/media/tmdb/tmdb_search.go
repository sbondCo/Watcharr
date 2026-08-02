package tmdb

import (
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/sbondCo/Watcharr/cache"
)

// This file has all tmdb search methods.
// Each helper has their own Options struct.
// Each Options struct has an `AsParamsMap` function attached that converts
// the properties of the struct into a query params map we can pass to the
// requests (in a tmdb supported fashion).

// **NOTE:** Ensure any new options are also added to the cache keys of the
// search functions using them!

// Options that all the Search structs support.
type SearchUniversalOptions struct {
	Query string
	Page  int
	Adult bool
}

// Check if SearchUniversalOptions is valid.
// Fixes `Page` to equal `1` if it is `0` (unset).
func (o *SearchUniversalOptions) Valid() bool {
	if o.Query == "" {
		// A query is necessary!
		return false
	}
	if o.Page == 0 {
		o.Page = 1
	}
	return true
}

func (o *SearchUniversalOptions) AsParamsMap() map[string]string {
	m := map[string]string{
		"query": o.Query,
		"page":  strconv.Itoa(o.Page),
	}
	if o.Adult {
		m["include_adult"] = "true"
	}
	return m
}

func (t *TMDB) SearchMulti(
	o SearchUniversalOptions,
) (TMDBSearchMultiResponse, error) {
	resp := new(TMDBSearchMultiResponse)
	if !o.Valid() {
		return *resp, errors.New("request is invalid")
	}
	cacheKey := cache.CreateCacheKey(
		"SearchMulti",
		o.Query,
		o.Page,
		o.Adult)
	if cache.GetCache(ContentStore, cacheKey, &resp) {
		slog.Debug("SearchMulti: Returning cache.")
		return *resp, nil
	}
	err := t.req(
		"/search/multi",
		o.AsParamsMap(),
		&resp)
	if err != nil {
		slog.Error("SearchMulti: Request failed!", "error", err)
		return TMDBSearchMultiResponse{}, errors.New("request failed")
	}
	ContentStore.Set(cacheKey, resp, time.Hour*24)
	return *resp, nil
}

type SearchMoviesOptions struct {
	SearchUniversalOptions
	Year        int
	PrimaryYear int
}

func (o *SearchMoviesOptions) AsParamsMap() map[string]string {
	m := o.SearchUniversalOptions.AsParamsMap()
	if o.Year != 0 {
		m["year"] = strconv.Itoa(o.Year)
	}
	if o.PrimaryYear != 0 {
		m["primary_release_year"] = strconv.Itoa(o.PrimaryYear)
	}
	return m
}

func (t *TMDB) SearchMovies(
	o SearchMoviesOptions,
) (TMDBSearchMoviesResponse, error) {
	resp := new(TMDBSearchMoviesResponse)
	if !o.Valid() {
		return *resp, errors.New("request is invalid")
	}
	cacheKey := cache.CreateCacheKey(
		"SearchMovies",
		o.Query,
		o.Page,
		o.Adult,
		o.Year,
		o.PrimaryYear)
	if cache.GetCache(ContentStore, cacheKey, &resp) {
		slog.Debug("SearchMovies: Returning cache.")
		return *resp, nil
	}
	err := t.req(
		"/search/movie",
		o.AsParamsMap(),
		&resp)
	if err != nil {
		slog.Error("SearchMovies: Request failed!", "error", err)
		return TMDBSearchMoviesResponse{}, errors.New("request failed")
	}
	for i := range resp.Results {
		resp.Results[i].MediaType = "movie"
	}
	ContentStore.Set(cacheKey, resp, time.Hour*24)
	return *resp, nil
}

type SearchShowsOptions struct {
	SearchUniversalOptions
	Year        int
	PrimaryYear int
}

func (o *SearchShowsOptions) AsParamsMap() map[string]string {
	m := o.SearchUniversalOptions.AsParamsMap()
	if o.Year != 0 {
		m["year"] = strconv.Itoa(o.Year)
	}
	if o.PrimaryYear != 0 {
		m["first_air_date_year"] = strconv.Itoa(o.PrimaryYear)
	}
	return m
}

func (t *TMDB) SearchShows(
	o SearchShowsOptions,
) (TMDBSearchShowsResponse, error) {
	resp := new(TMDBSearchShowsResponse)
	if !o.Valid() {
		return *resp, errors.New("request is invalid")
	}
	cacheKey := cache.CreateCacheKey(
		"SearchShows",
		o.Query,
		o.Page,
		o.Adult,
		o.Year,
		o.PrimaryYear)
	if cache.GetCache(ContentStore, cacheKey, &resp) {
		slog.Debug("SearchShows: Returning cache.")
		return *resp, nil
	}
	err := t.req(
		"/search/tv",
		o.AsParamsMap(),
		&resp)
	if err != nil {
		slog.Error("SearchShows: Request failed!", "error", err)
		return TMDBSearchShowsResponse{}, errors.New("request failed")
	}
	for i := range resp.Results {
		resp.Results[i].MediaType = "tv"
	}
	ContentStore.Set(cacheKey, resp, time.Hour*24)
	return *resp, nil
}

func (t *TMDB) SearchPeople(
	o SearchUniversalOptions,
) (TMDBSearchPeopleResponse, error) {
	resp := new(TMDBSearchPeopleResponse)
	if !o.Valid() {
		return *resp, errors.New("request is invalid")
	}
	cacheKey := cache.CreateCacheKey(
		"SearchPeople",
		o.Query,
		o.Page,
		o.Adult)
	if cache.GetCache(ContentStore, cacheKey, &resp) {
		slog.Debug("SearchPeople: Returning cache.")
		return *resp, nil
	}
	err := t.req(
		"/search/person",
		o.AsParamsMap(),
		&resp)
	if err != nil {
		slog.Error("SearchPeople: Request failed!", "error", err)
		return TMDBSearchPeopleResponse{}, errors.New("request failed")
	}
	for i := range resp.Results {
		resp.Results[i].MediaType = "person"
	}
	ContentStore.Set(cacheKey, resp, time.Hour*24)
	return *resp, nil
}

// Search for content by an external id (imdb, etc).
// Defaults to imdb if no source if provided (probably most common).
func (t *TMDB) SearchByExternalId(
	id string,
	source string,
) (TMDBSearchMultiResponse, error) {
	resp := new(TMDBFindByExternalIdResponse)
	if source == "" {
		source = "imdb"
	}
	cacheKey := cache.CreateCacheKey("SearchByExternalId", id, source)
	if cache.GetCache(ContentStore, cacheKey, &resp) {
		slog.Debug("SearchByExternalId: Got cache.")
	} else {
		// If not found in cache, request data from tmdb.
		err := t.req(
			"/find/"+id,
			map[string]string{"external_source": source + "_id"},
			&resp)
		if err != nil {
			slog.Error("Failed to complete find/external_id request!",
				"error", err.Error())
			return TMDBSearchMultiResponse{},
				errors.New("failed to complete find/external_id request")
		}
		ContentStore.Set(cacheKey, resp, time.Hour*24)
	}
	comb := []TMDBSearchMultiResult{}
	comb = append(comb, resp.MovieResults...)
	comb = append(comb, resp.TvResults...)
	comb = append(comb, resp.PersonResults...)
	comb = append(comb, resp.TvSeasonResults...)
	comb = append(comb, resp.TvEpisodeResults...)
	return TMDBSearchMultiResponse{
			TMDBSearchResponse: TMDBSearchResponse[TMDBSearchMultiResult]{
				Results: comb,
				TMDBPageFields: TMDBPageFields{
					TotalResults: len(comb),
					// Just providing these so we don't break frontend pagination logic.
					TotalPages: 1,
					Page:       1,
				},
			},
		},
		nil
}
