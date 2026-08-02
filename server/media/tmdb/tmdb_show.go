package tmdb

import (
	"errors"
	"log/slog"
	"time"

	"github.com/sbondCo/Watcharr/cache"
)

type ShowDetailsOptions struct {
	// TMDB ID
	ID string
	// Country (currently used for watch providers)
	Country string
	// Request params map.
	Params map[string]string

	// If CacheContentShow should be ran or not.
	// If the caller wants to do its own caching to the db, it can use this
	// to avoid multiple calls to CacheContentShow.
	DontRunDBCache bool
}

func (t *TMDB) ShowDetails(o ShowDetailsOptions) (TMDBShowDetails, error) {
	cacheKey := cache.CreateCacheKey(
		"ShowDetails",
		o.ID,
		o.Country,
		o.Params)
	resp := new(TMDBShowDetails)
	if cache.GetCache(ContentStore, cacheKey, &resp) {
		slog.Debug("ShowDetails: Returning cache.")
		return *resp, nil
	}
	err := t.req("/tv/"+o.ID, o.Params, &resp)
	if err != nil {
		slog.Error("ShowDetails: Request failed!", "error", err)
		return TMDBShowDetails{}, errors.New("request failed")
	}
	resp.WatchProvidersTransformed = transformProviders(
		&resp.WatchProviders,
		o.Country)
	// We don't want this to linger around (in cache) since we have the
	// transformed version now..
	resp.WatchProviders = nil
	if !o.DontRunDBCache {
		go t.contentProvider.CacheContentShow(*resp, true)
	}
	ContentStore.Set(cacheKey, resp, time.Hour*24)
	return *resp, nil
}

func (t *TMDB) ShowCredits(id string) (TMDBContentCredits, error) {
	resp := new(TMDBContentCredits)
	err := t.req("/tv/"+id+"/credits", map[string]string{}, &resp)
	if err != nil {
		slog.Error("ShowCredits: Request failed!", "error", err)
		return TMDBContentCredits{}, errors.New("request failed")
	}
	return *resp, nil
}

func (t *TMDB) SeasonDetails(
	showId string,
	seasonNumber string,
) (TMDBSeasonDetails, error) {
	cacheKey := cache.CreateCacheKey("SeasonDetails", showId, seasonNumber)
	resp := new(TMDBSeasonDetails)
	if cache.GetCache(ContentStore, cacheKey, &resp) {
		slog.Debug("SeasonDetails: Returning cache.")
		return *resp, nil
	}
	err := t.req(
		"/tv/"+showId+"/season/"+seasonNumber,
		map[string]string{},
		&resp)
	if err != nil {
		slog.Error("SeasonDetails: Request failed!", "error", err)
		return TMDBSeasonDetails{}, errors.New("request failed")
	}
	ContentStore.Set(cacheKey, resp, time.Hour*24)
	return *resp, nil
}
