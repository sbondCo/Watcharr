package tmdb

import (
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/sbondCo/Watcharr/cache"
)

func (t *TMDB) PersonDetails(id string) (TMDBPersonDetails, error) {
	resp := new(TMDBPersonDetails)
	err := t.req("/person/"+id, map[string]string{}, &resp)
	if err != nil {
		slog.Error("PersonDetails: Request failed!", "error", err)
		return TMDBPersonDetails{}, errors.New("request failed")
	}
	return *resp, nil
}

func (t *TMDB) PersonCredits(id string) (TMDBPersonCombinedCredits, error) {
	cacheKey := cache.CreateCacheKey("PersonCredits", id)
	resp := new(TMDBPersonCombinedCredits)
	if cache.GetCache(ContentStore, cacheKey, &resp) {
		slog.Debug("PersonCredits: Returning cache.")
		return *resp, nil
	}
	err := t.req(
		"/person/"+id+"/combined_credits",
		map[string]string{},
		&resp)
	if err != nil {
		slog.Error("PersonCredits: Request failed!", "error", err)
		return TMDBPersonCombinedCredits{}, errors.New("request failed")
	}
	ContentStore.Set(cacheKey, resp, time.Hour*24)
	return *resp, nil
}

func (t *TMDB) PopularPeople(pageNum int) (TMDBPopularPeople, error) {
	cacheKey := cache.CreateCacheKey("PopularPeople", pageNum)
	resp := new(TMDBPopularPeople)
	if cache.GetCache(ContentStore, cacheKey, &resp) {
		slog.Debug("PopularPeople: Returning cache.")
		return *resp, nil
	}
	err := t.req(
		"/person/popular",
		map[string]string{"page": strconv.Itoa(pageNum)},
		&resp)
	if err != nil {
		slog.Error("PopularPeople: Request failed!", "error", err)
		return TMDBPopularPeople{}, errors.New("request failed")
	}
	ContentStore.Set(cacheKey, resp, time.Hour*24)
	return *resp, nil
}
