// Sonarr and Radarr

package arr

import (
	"bytes"
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

type ArrRequest struct {
	ServerName      string `json:"serverName"`
	QualityProfile  int    `json:"qualityProfile"` // id
	RootFolder      string `json:"rootFolder"`     // path
	AutomaticSearch bool   `json:"automaticSearch"`
	Title           string `json:"title"` // content name
	Year            int    `json:"year"`  // content year
	TMDBID          int    `json:"tmdbId"`
}

type SonarrRequest struct {
	ArrRequest
	TVDBID          int             `json:"tvdbId"`
	LanguageProfile int             `json:"languageProfile"` // id
	SeriesType      string          `json:"seriesType"`
	Seasons         []SonarrSeasons `json:"seasons"`
}

type RadarrRequest struct {
	ArrRequest
}

type SonarrSeasons struct {
	SeasonNumber int  `json:"seasonNumber"`
	Monitored    bool `json:"monitored"`
}

func New(t ArrType, host *string, key *string) *Arr {
	return &Arr{
		Type: t,
		Host: host,
		Key:  key,
	}
}

func (a *Arr) GetQualityProfiles() ([]QualityProfile, error) {
	slog.Debug("GetQualityProfiles", "type", a.Type, "host", *a.Host, "key", *a.Key)
	var resp []QualityProfile
	err := request(*a.Host, "/qualityprofile", map[string]string{"apikey": *a.Key}, &resp)
	if err != nil {
		slog.Error("GetQualityProfiles request failed", "service", a.Type, "error", err)
		return []QualityProfile{}, errors.New("request to service failed")
	}
	return resp, nil
}

func (a *Arr) GetRootFolders() ([]RootFolder, error) {
	slog.Debug("GetRootFolders", "type", a.Type, "host", *a.Host, "key", *a.Key)
	var resp []RootFolder
	err := request(*a.Host, "/rootfolder", map[string]string{"apikey": *a.Key}, &resp)
	if err != nil {
		slog.Error("GetRootFolders request failed", "service", a.Type, "error", err)
		return []RootFolder{}, errors.New("request to service failed")
	}
	return resp, nil
}

func (a *Arr) GetLangaugeProfiles() ([]LanguageProfile, error) {
	slog.Debug("GetLangaugeProfiles", "type", a.Type, "host", *a.Host, "key", *a.Key)
	var resp []LanguageProfile
	// languageprofile supposedly deprecated.. but new language endpoint doesnt seem to work.. note probs to switch soon
	// TODO languages are handled diffferently in Sonarr now, I think we can remove all language stuff, now controlled per profile.
	err := request(*a.Host, "/languageprofile", map[string]string{"apikey": *a.Key}, &resp)
	if err != nil {
		slog.Error("GetLangaugeProfiles request failed", "service", a.Type, "error", err)
		return []LanguageProfile{}, errors.New("request to service failed")
	}
	return resp, nil
}

func (a *Arr) RunCommand(name string) (CommandResponse, error) {
	slog.Debug("RunCommand", "name", name, "type", a.Type, "host", *a.Host, "key", *a.Key)
	var resp CommandResponse
	err := requestPost(*a.Host, "/command", *a.Key, map[string]interface{}{"name": name}, &resp)
	if err != nil {
		slog.Error("RunCommand request failed", "name", name, "service", a.Type, "error", err)
		return CommandResponse{}, errors.New("request to service failed")
	}
	return resp, nil
}

// arrId = movieId/seriesId on radarr/sonarr
func (a *Arr) GetQueueDetails(arrId string, resp interface{}) error {
	slog.Debug("GetQueueDetails", "arrId", arrId, "type", a.Type, "host", *a.Host, "key", *a.Key)
	p := map[string]string{"apikey": *a.Key}
	if a.Type == RADARR {
		p["movieId"] = arrId
	} else if a.Type == SONARR {
		p["seriesId"] = arrId
	} else {
		return errors.New("invalid arr type")
	}
	err := request(*a.Host, "/queue/details", p, resp)
	if err != nil {
		slog.Error("GetQueueDetails request failed", "arrId", arrId, "service", a.Type, "error", err)
		return errors.New("request to service failed")
	}
	return nil
}

// Get movie/show
func (a *Arr) GetContent(arrId string) (MovieSerie, error) {
	slog.Debug("GetContent", "arrId", arrId, "type", a.Type, "host", *a.Host, "key", *a.Key)
	e := "movie"
	if a.Type == SONARR {
		e = "series"
	}
	var resp MovieSerie
	err := request(*a.Host, "/"+e+"/"+arrId, map[string]string{"apikey": *a.Key}, &resp)
	if err != nil {
		slog.Error("GetContent request failed", "arrId", arrId, "service", a.Type, "error", err)
		return MovieSerie{}, errors.New("request to service failed")
	}
	return resp, nil
}

func (a *Arr) BuildAddShowBody(r SonarrRequest) map[string]interface{} {
	req := map[string]interface{}{
		"title":             r.Title,
		"year":              r.Year,
		"qualityProfileId":  r.QualityProfile,
		"languageProfileId": r.LanguageProfile,
		"seasonFolder":      true,
		"monitored":         true,
		"tvdbId":            r.TVDBID,
		"seriesType":        r.SeriesType,
		"seasons":           r.Seasons,
		"addOptions": map[string]interface{}{
			"ignoreEpisodesWithFiles":  true,
			"searchForMissingEpisodes": r.AutomaticSearch,
		},
		"rootFolderPath": r.RootFolder,
	}
	return req
}

func (a *Arr) BuildAddMovieBody(r RadarrRequest) map[string]interface{} {
	req := map[string]interface{}{
		"title":            r.Title,
		"year":             r.Year,
		"qualityProfileId": r.QualityProfile,
		"monitored":        true,
		"tmdbId":           r.TMDBID,
		"addOptions": map[string]interface{}{
			"searchForMovie": r.AutomaticSearch,
		},
		"rootFolderPath": r.RootFolder,
	}
	return req
}

func (a *Arr) AddContent(b map[string]interface{}) (map[string]interface{}, error) {
	ep := "series"
	if a.Type == RADARR {
		ep = "movie"
	}
	slog.Debug("AddContent", "type", ep, "body", b)
	var resp map[string]interface{}
	err := requestPost(*a.Host, "/"+ep, *a.Key, b, &resp)
	if err != nil {
		slog.Error("AddContent request failed", "service", a.Type, "error", err)
		return map[string]interface{}{}, errors.New("request to service failed")
	}
	slog.Debug("AddContent", "type", ep, "created_id", resp["id"])
	return resp, nil
}

func request(host string, ep string, p map[string]string, resp interface{}) error {
	slog.Debug("arrAPIRequest", "endpoint", ep, "params", p)
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

func requestPost(host string, ep string, key string, p map[string]interface{}, resp interface{}) error {
	base, err := url.Parse(host)
	if err != nil {
		return errors.New("failed to parse api uri")
	}

	// Path params
	base.Path += "/api/v3" + ep

	var res *http.Response

	// Query params
	params := url.Values{}
	params.Add("apikey", key)

	// Add params to url
	base.RawQuery = params.Encode()

	jsonp, err := json.Marshal(p)
	if err != nil {
		return err
	}
	res, err = http.Post(base.String(), "application/json", bytes.NewBuffer(jsonp))
	if err != nil {
		return err
	}

	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return err
	}
	if !(res.StatusCode >= 200 && res.StatusCode <= 299) {
		slog.Error("arr non 2xx status code:", "status_code", res.StatusCode)
		return errors.New(string(body))
	}
	err = json.Unmarshal([]byte(body), &resp)
	if err != nil {
		return err
	}
	return nil
}
