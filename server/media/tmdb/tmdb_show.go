package tmdb

import (
	"errors"
	"log/slog"
	"time"

	"github.com/sbondCo/Watcharr/cache"
)

func (t *TMDB) ShowDetails(
	id string,
	country string,
	rParams map[string]string,
) (TMDBShowDetails, error) {
	cacheKey := cache.CreateCacheKey("ShowDetails", id, country, rParams)
	resp := new(TMDBShowDetails)
	if cache.GetCache(ContentStore, cacheKey, &resp) {
		slog.Debug("ShowDetails: Returning cache.")
		return *resp, nil
	}
	err := t.Request("/tv/"+id, rParams, &resp)
	if err != nil {
		slog.Error("ShowDetails: Request failed!", "error", err)
		return TMDBShowDetails{}, errors.New("request failed")
	}
	resp.WatchProvidersTransformed = transformProviders(&resp.WatchProviders, country)
	resp.WatchProviders = nil // We don't want this to linger around (in cache) since we have the transformed version now..
	go t.contentProvider.CacheContentTv(*resp, true)
	ContentStore.Set(cacheKey, resp, time.Hour*24)
	return *resp, nil
}

func (t *TMDB) ShowCredits(id string) (TMDBContentCredits, error) {
	resp := new(TMDBContentCredits)
	err := t.Request("/tv/"+id+"/credits", map[string]string{}, &resp)
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
	err := t.Request(
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
