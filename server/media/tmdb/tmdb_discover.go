package tmdb

import (
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/sbondCo/Watcharr/cache"
)

func (t *TMDB) DiscoverMovies(
	o DiscoverOptions,
	pageNum int,
	region string,
) (TMDBDiscoverMovies, error) {
	resp := new(TMDBDiscoverMovies)
	reqParams := map[string]string{
		"page":   strconv.Itoa(pageNum),
		"region": region,
	}
	t.applyDiscoverOptionsToMap(true, o, reqParams)
	cacheKey := cache.CreateCacheKey(
		"DiscoverMovies",
		pageNum,
		reqParams)
	if cache.GetCache(ContentStore, cacheKey, &resp) {
		slog.Debug("DiscoverMovies: Returning cache.")
		return *resp, nil
	}
	err := t.req("/discover/movie", reqParams, &resp)
	if err != nil {
		slog.Error("DiscoverMovies: Request failed!", "error", err)
		return TMDBDiscoverMovies{}, errors.New("request failed")
	}
	ContentStore.Set(cacheKey, resp, time.Hour*24)
	return *resp, nil
}

func (t *TMDB) DiscoverShows(
	o DiscoverOptions,
	pageNum int,
	region string,
) (TMDBDiscoverShows, error) {
	resp := new(TMDBDiscoverShows)
	reqParams := map[string]string{
		"page":   strconv.Itoa(pageNum),
		"region": region,
	}
	t.applyDiscoverOptionsToMap(false, o, reqParams)
	cacheKey := cache.CreateCacheKey(
		"DiscoverShows",
		pageNum,
		reqParams)
	if cache.GetCache(ContentStore, cacheKey, &resp) {
		slog.Debug("DiscoverShows: Returning cache.")
		return *resp, nil
	}
	err := t.req("/discover/tv", reqParams, &resp)
	if err != nil {
		slog.Error("DiscoverShows: Request failed!", "error", err)
		return TMDBDiscoverShows{}, errors.New("request failed")
	}
	ContentStore.Set(cacheKey, resp, time.Hour*24)
	return *resp, nil
}

func (t *TMDB) applyDiscoverOptionsToMap(
	// Some properties are named differently for sorting the same thing as far
	// as we care, so we need to differenciate to name them properly.
	forMovie bool,
	o DiscoverOptions,
	m map[string]string,
) {
	releaseDateMinKey := "release_date.gte"
	releaseDateMaxKey := "release_date.lte"
	withReleaseTypeKey := "with_release_type"
	if !forMovie {
		// Replace with names for equivalent tv filters
		releaseDateMinKey = "first_air_date.gte"
		releaseDateMaxKey = "first_air_date.lte"
		withReleaseTypeKey = "with_type"
	}
	if !o.ReleaseDateMin.IsZero() {
		m[releaseDateMinKey] = o.ReleaseDateMin.Format("2006-01-02")
	}
	if !o.ReleaseDateMax.IsZero() {
		m[releaseDateMaxKey] = o.ReleaseDateMax.Format("2006-01-02")
	}
	if o.WithReleaseType != "" {
		m[withReleaseTypeKey] = o.WithReleaseType
	}
}
