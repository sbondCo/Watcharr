package game

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

const (
	igdbHost       = "https://api.igdb.com/v4"
	tokenGrantType = "client_credentials"
)

// Config (for admins to set) required to be set for use of igdb.
// type IGDBConfig struct {

// }

type IGDB struct {
	ClientID           *string   `json:"clientId,omitempty"`
	ClientSecret       *string   `json:"clientSecret,omitempty"`
	AccessToken        string    `json:"accessToken,omitempty"`
	AccessTokenExpires time.Time `json:"accessTokenExpires,omitempty"`
}

func New(cfg *IGDB) *IGDB {
	// cfg.init()
	return cfg
}

func (i *IGDB) req(host string, ep string, p map[string]string, b string, resp interface{}) error {
	// if using igdb host and we have no access token, error before running req
	if host == igdbHost && (i.ClientID == nil || i.AccessToken == "") {
		return errors.New("using igdbHost without a clientID or accessToken")
	}

	base, err := url.Parse(host)
	if err != nil {
		return errors.New("failed to parse api uri")
	}

	// Path params
	base.Path += ep

	// Query params
	params := url.Values{}
	for k, v := range p {
		params.Add(k, v)
	}

	// Add params to url
	base.RawQuery = params.Encode()

	slog.Info("req", "base", base.String())

	req, err := http.NewRequest("POST", base.String(), bytes.NewBuffer([]byte(b)))
	if err != nil {
		return err
	}

	// Add igdb auth headers
	if host == igdbHost {
		req.Header.Add("Client-ID", *i.ClientID)
		req.Header.Add("Authorization", "Bearer "+i.AccessToken)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return err
	}
	if !(res.StatusCode >= 200 && res.StatusCode <= 299) {
		slog.Error("game non 2xx status code:", "status_code", res.StatusCode)
		return errors.New(string(body))
	}
	err = json.Unmarshal([]byte(body), &resp)
	if err != nil {
		return err
	}
	return nil
}

// Get token and stuff
func (i *IGDB) Init() error {
	if i.ClientID == nil || i.ClientSecret == nil {
		slog.Error("IGDB init client id and or secret not provided")
		return errors.New("client id and or secret not provided")
	}
	slog.Debug("IGDB init running.", "client_id", *i.ClientID, "client_secret", *i.ClientSecret)
	// If we have an unexpired access token already, use that instead of requesting a new one.
	if i.AccessToken != "" && !i.AccessTokenExpires.IsZero() {
		if i.AccessTokenExpires.Compare(time.Now()) == 1 {
			slog.Debug("IGDB init current access token hasn't expired. Will continue using that one.")
			return nil
		}
		slog.Debug("IGDB init current access token has expired.. fetching a new one.")
	}
	var resp TwitchTokenResponse
	err := i.req(
		"https://id.twitch.tv/oauth2",
		"/token",
		map[string]string{"client_id": *i.ClientID, "client_secret": *i.ClientSecret, "grant_type": tokenGrantType},
		"",
		&resp,
	)
	if err != nil {
		slog.Error("IGDB init token request failed", "error", err)
		return errors.New("token request failed, check client id and secret")
	}
	i.AccessToken = resp.AccessToken
	i.AccessTokenExpires = time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second)
	slog.Debug("IGDB init token response", "resp", resp, "token_expires", i.AccessTokenExpires)
	return nil
}

func (i *IGDB) Search(q string) (GameSearchResponse, error) {
	slog.Debug("IGDB Search called", "query", q)
	var resp GameSearchResponse
	err := i.req(
		igdbHost,
		"/games",
		map[string]string{},
		"fields name, cover.image_id, version_title, summary, first_release_date; search \""+q+"\";",
		&resp,
	)
	if err != nil {
		slog.Error("IGDB Search request failed!", "error", err)
		return GameSearchResponse{}, errors.New("request failed")
	}
	return resp, nil
}

func (i *IGDB) GameDetails(id string) (GameDetailsResponse, error) {
	slog.Debug("IGDB GameDetails called", "id", id)
	var resp []GameDetailsResponse
	err := i.req(
		igdbHost,
		"/games",
		map[string]string{},
		`fields 
			name,
			cover.image_id,
			version_title,
			summary,
			storyline,
			first_release_date,
			artworks.width,
			artworks.height,
			artworks.image_id,
			category,
			platforms.name,
			game_modes.name,
			genres.id,
			genres.name,
			involved_companies.developer,
			involved_companies.publisher,
			involved_companies.porting,
			involved_companies.supporting,
			involved_companies.company.name,
			involved_companies.company.description,
			involved_companies.company.slug,
			involved_companies.company.websites.category,
			involved_companies.company.websites.trusted,
			involved_companies.company.websites.url,
			rating,
			rating_count,
			status,
			url,
			websites.trusted,
			websites.category,
			websites.url,
			videos.name,
			videos.video_id;
		where id = `+id+";",
		&resp,
	)
	if err != nil {
		slog.Error("IGDB GameDetails request failed!", "error", err)
		return GameDetailsResponse{}, errors.New("request failed")
	}
	if len(resp) > 0 {
		return resp[0], nil
	}
	return GameDetailsResponse{}, errors.New("no game details recieved")
}

// Basic game details for when we are using them only to update our cache.
// In these cases, it's a waste to ask for everything, when we don't need it.
func (i *IGDB) GameDetailsBasic(id string) (GameDetailsBasicResponse, error) {
	slog.Debug("IGDB GameDetails called", "id", id)
	var resp []GameDetailsBasicResponse
	err := i.req(
		igdbHost,
		"/games",
		map[string]string{},
		`fields 
			name,
			cover.image_id,
			summary,
			storyline,
			first_release_date,
			category,
			platforms.name,
			game_modes.name,
			genres.name,
			rating,
			rating_count,
			status;
		where id = `+id+";",
		&resp,
	)
	if err != nil {
		slog.Error("IGDB GameDetails request failed!", "error", err)
		return GameDetailsBasicResponse{}, errors.New("request failed")
	}
	if len(resp) > 0 {
		return resp[0], nil
	}
	return GameDetailsBasicResponse{}, errors.New("no game details recieved")
}
