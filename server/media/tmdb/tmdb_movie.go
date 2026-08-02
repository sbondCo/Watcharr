package tmdb

import (
	"errors"
	"log/slog"
	"time"

	"github.com/sbondCo/Watcharr/cache"
)

func (t *TMDB) MovieDetails(
	id string,
	country string,
	rParams map[string]string,
) (TMDBMovieDetails, error) {
	resp := new(TMDBMovieDetails)
	cacheKey := cache.CreateCacheKey("MovieDetails", id, country, rParams)
	if cache.GetCache(ContentStore, cacheKey, &resp) {
		slog.Debug("MovieDetails: Returning cache.")
		return *resp, nil
	}
	err := t.Request("/movie/"+id, rParams, &resp)
	if err != nil {
		slog.Error("MovieDetails: Request failed!", "error", err)
		return TMDBMovieDetails{}, errors.New("request failed")
	}
	resp.WatchProvidersTransformed = transformProviders(&resp.WatchProviders, country)
	resp.WatchProviders = nil // We don't want this to linger around (in cache) since we have the transformed version now..
	go t.contentProvider.CacheContentMovie(*resp, true)
	ContentStore.Set(cacheKey, resp, time.Hour*24)
	return *resp, nil
}

func (t *TMDB) MovieCredits(id string) (TMDBContentCredits, error) {
	resp := new(TMDBContentCredits)
	err := t.Request("/movie/"+id+"/credits", map[string]string{}, &resp)
	if err != nil {
		slog.Error("MovieCredits: Request failed!", "error", err)
		return TMDBContentCredits{}, errors.New("request failed")
	}
	return *resp, nil
}
