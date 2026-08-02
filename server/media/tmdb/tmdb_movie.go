package tmdb

import (
	"errors"
	"log/slog"
	"time"

	"github.com/sbondCo/Watcharr/cache"
)

type MovieDetailsOptions struct {
	// TMDB ID
	ID string
	// Country (currently used for watch providers)
	Country string
	// Request params map.
	Params map[string]string

	// If CacheContentMovie should be ran or not.
	// If the caller wants to do its own caching to the db, it can use this
	// to avoid multiple calls to CacheContentMovie.
	DontRunDBCache bool
}

func (t *TMDB) MovieDetails(o MovieDetailsOptions) (TMDBMovieDetails, error) {
	resp := new(TMDBMovieDetails)
	cacheKey := cache.CreateCacheKey(
		"MovieDetails",
		o.ID,
		o.Country,
		o.Params)
	if cache.GetCache(ContentStore, cacheKey, &resp) {
		slog.Debug("MovieDetails: Returning cache.")
		return *resp, nil
	}
	err := t.req("/movie/"+o.ID, o.Params, &resp)
	if err != nil {
		slog.Error("MovieDetails: Request failed!", "error", err)
		return TMDBMovieDetails{}, errors.New("request failed")
	}
	resp.WatchProvidersTransformed = transformProviders(
		&resp.WatchProviders,
		o.Country)
	// We don't want this to linger around (in cache) since we have the
	// transformed version now..
	resp.WatchProviders = nil
	if !o.DontRunDBCache {
		go t.contentProvider.CacheContentMovie(*resp, true)
	}
	ContentStore.Set(cacheKey, resp, time.Hour*24)
	return *resp, nil
}

func (t *TMDB) MovieCredits(id string) (TMDBContentCredits, error) {
	resp := new(TMDBContentCredits)
	err := t.req("/movie/"+id+"/credits", map[string]string{}, &resp)
	if err != nil {
		slog.Error("MovieCredits: Request failed!", "error", err)
		return TMDBContentCredits{}, errors.New("request failed")
	}
	return *resp, nil
}
