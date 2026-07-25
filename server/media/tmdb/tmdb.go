package tmdb

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
)

// TODO rewrite tmdb to work like how igdb package was made
// TODO The *WithWatched structs likely need to go in the watched package (or with go 1.25 can we
// fix needing so many extra structs for the *WithWatched types and functions)

type TMDB struct {
	Key  string
	Lang string
}

func NewTMDB(key string, lang string) *TMDB {
	return &TMDB{
		Key:  key,
		Lang: lang,
	}
}

func (t *TMDB) GetKey() string {
	if t.Key != "" {
		return t.Key //Config.TMDB_KEY
	}
	return "d047fa61d926371f277e7a83c9c4ff2c"
}

// GetLang returns the configured TMDB metadata language (ISO 639-1, optionally
// with a region, e.g. "fr-FR"), falling back to English when unset.
func (t *TMDB) GetLang() string {
	if t.Lang != "" {
		return t.Lang
	}
	return "en-US"
}

func (t *TMDB) APIRequest(ep string, p map[string]string) ([]byte, error) {
	slog.Debug("tmdbAPIRequest", "endpoint", ep, "params", p)
	base, err := url.Parse("https://api.themoviedb.org/3")
	if err != nil {
		return nil, errors.New("failed to parse api uri")
	}

	// Path params
	base.Path += ep

	// Query params
	params := url.Values{}
	params.Add("api_key", t.GetKey())
	params.Add("language", t.GetLang())
	for k, v := range p {
		params.Add(k, v)
	}

	// Add params to url
	base.RawQuery = params.Encode()

	// Run get request
	res, err := http.Get(base.String())
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		slog.Error("TMDB non 200 status code:", "status_code", res.StatusCode)
		return nil, errors.New(string(body))
	}
	return body, nil
}

func (t *TMDB) Request(ep string, p map[string]string, resp interface{}) error {
	body, err := t.APIRequest(ep, p)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(body), &resp)
	if err != nil {
		return err
	}
	return nil
}
