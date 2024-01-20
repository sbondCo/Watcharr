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
	ClientID           *string `json:"clientId,omitempty"`
	ClientSecret       *string `json:"clientSecret,omitempty"`
	accessToken        string
	accessTokenExpires time.Time
}

func New(cfg *IGDB) *IGDB {
	// cfg.init()
	return cfg
}

func (i *IGDB) req(host string, ep string, p map[string]string, b map[string]interface{}, resp interface{}) error {
	// TODO if using igdb host and we have no access token, error before running req

	base, err := url.Parse(host)
	if err != nil {
		return errors.New("failed to parse api uri")
	}

	// Path params
	base.Path += ep

	var res *http.Response

	// Query params
	params := url.Values{}
	for k, v := range p {
		params.Add(k, v)
	}

	// Add params to url
	base.RawQuery = params.Encode()

	slog.Info("req", "base", base.String())

	if len(b) > 0 {
		jsonp, err := json.Marshal(b)
		if err != nil {
			return err
		}
		res, err = http.Post(base.String(), "application/json", bytes.NewBuffer(jsonp))
		if err != nil {
			return err
		}
	} else {
		res, err = http.Post(base.String(), "application/json", bytes.NewBuffer(nil))
		if err != nil {
			return err
		}
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
	var resp TwitchTokenResponse
	err := i.req(
		"https://id.twitch.tv/oauth2",
		"/token",
		map[string]string{"client_id": *i.ClientID, "client_secret": *i.ClientSecret, "grant_type": tokenGrantType},
		map[string]interface{}{},
		&resp,
	)
	if err != nil {
		slog.Error("IGDB init token request failed", "error", err)
		return errors.New("token request failed, check client id and secret")
	}
	i.accessToken = resp.AccessToken
	i.accessTokenExpires = time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second)
	slog.Debug("IGDB init token response", "resp", resp, "token_expires", i.accessTokenExpires)
	return nil
}

func (i *IGDB) Search() {
	slog.Debug("IGDB Search called")
}
