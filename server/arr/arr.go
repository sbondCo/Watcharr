// Sonarr and Radarr

package arr

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
)

type ArrType string

var (
	SONARR ArrType = "SONARR"
	RADARR ArrType = "RADARR"
)

type Arr struct {
	// Type of Arr we want to use.
	// Each servars api might differ, so this will
	// allow us to make those changes when needed.
	Type ArrType
	// Hostname to the Arr sever
	Host *string
	// Api key for the Arr server.
	Key *string
}

func New(t ArrType, host *string, key *string) *Arr {
	return &Arr{
		Type: t,
		Host: host,
		Key:  key,
	}
}

func (a *Arr) GetQualityProfiles() ([]QualityProfile, error) {
	slog.Info("GetQualityProfiles", "type", a.Type, "host", *a.Host, "key", *a.Key)
	var resp []QualityProfile
	err := request(*a.Host, "/qualityprofile", map[string]string{"apikey": *a.Key}, &resp)
	if err != nil {
		slog.Error("GetQualityProfiles request failed", "service", a.Type, "error", err)
		return []QualityProfile{}, errors.New("request to service failed")
	}
	return resp, nil
}

func request(host string, ep string, p map[string]string, resp interface{}) error {
	slog.Debug("tmdbAPIRequest", "endpoint", ep, "params", p)
	base, err := url.Parse(host)
	if err != nil {
		return errors.New("failed to parse api uri")
	}

	// Path params
	base.Path += "/api/v3" + ep

	// Query params
	params := url.Values{}
	for k, v := range p {
		params.Add(k, v)
	}

	// Add params to url
	base.RawQuery = params.Encode()

	// Run get request
	res, err := http.Get(base.String())
	if err != nil {
		return err
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		slog.Error("arr non 200 status code:", "status_code", res.StatusCode)
		return errors.New(string(body))
	}
	// slog.Info("", "body", body)
	err = json.Unmarshal([]byte(body), &resp)
	if err != nil {
		return err
	}
	return nil
}
