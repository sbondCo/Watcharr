package tmdb

import (
	"errors"
	"log/slog"
)

func (t *TMDB) Regions() (Regions, error) {
	resp := new(Regions)
	err := t.req(
		"/watch/providers/regions",
		map[string]string{},
		&resp)
	if err != nil {
		slog.Error("Regions: Request failed", "error", err)
		return Regions{}, errors.New("request failed")
	}
	return *resp, nil
}
