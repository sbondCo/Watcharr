package tmdb

import (
	"errors"
	"log/slog"
)

func (t *TMDB) Regions() (TMDBRegions, error) {
	resp := new(TMDBRegions)
	err := t.req(
		"/watch/providers/regions",
		map[string]string{},
		&resp)
	if err != nil {
		slog.Error("Regions: Request failed", "error", err)
		return TMDBRegions{}, errors.New("request failed")
	}
	return *resp, nil
}
