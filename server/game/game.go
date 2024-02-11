package game

import (
	"bytes"
	"context"
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

var tokenRefreshJobCancel context.CancelFunc

type IGDB struct {
	ClientID           *string   `json:"clientId,omitempty"`
	ClientSecret       *string   `json:"clientSecret,omitempty"`
	AccessToken        string    `json:"accessToken,omitempty"`
	AccessTokenExpires time.Time `json:"accessTokenExpires,omitempty"`
	onTokenRefreshed   *func()
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

func (i *IGDB) getNewAccessToken() (TwitchTokenResponse, error) {
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
		return TwitchTokenResponse{}, errors.New("token request failed, check client id and secret")
	}
	return resp, nil
}

func (i *IGDB) refreshToken(ctx context.Context) {
	var exp <-chan time.Time

	if i.AccessToken != "" && !i.AccessTokenExpires.IsZero() && i.AccessTokenExpires.Compare(time.Now()) == 1 {
		// Stored token not expired.. set exp time to token exp time - 1h.
		exp = time.After(i.AccessTokenExpires.Sub(time.Now().Add(60 * time.Second)))
	} else {
		// Token expired.. exp now..
		exp = time.After(100 * time.Millisecond)
	}

	slog.Info("refreshToken running")

	for {
		select {
		case <-exp:
			slog.Info("IGDB refreshToken: Token expired (or is near expiry date)")
			r, err := i.getNewAccessToken()
			if err != nil {
				slog.Error("IGDB refreshToken: Error refreshing token (retrying in 60s):", err)
				exp = time.After(60 * time.Second)
			} else {
				slog.Info("IGDB refreshToken: Token successfully refreshed")
				i.AccessToken = r.AccessToken
				i.AccessTokenExpires = time.Now().Add(time.Duration(r.ExpiresIn) * time.Second)
				exp = time.After(i.AccessTokenExpires.Sub(time.Now().Add(60 * time.Second)))
				// Call token refresh callback if set
				if i.onTokenRefreshed != nil {
					(*i.onTokenRefreshed)()
				}
			}

		case <-ctx.Done():
			slog.Info("refreshToken cancelled")
			return
		}
	}
}

func (i *IGDB) OnTokenRefreshed(tokenRefreshed func()) {
	i.onTokenRefreshed = &tokenRefreshed
}

// Get token and stuff
func (i *IGDB) Init() error {
	// Cancel existing refresh job if we have a cancel func for its context
	if tokenRefreshJobCancel != nil {
		slog.Debug("IGDB init: Refresh job running.. cancelling it.")
		tokenRefreshJobCancel()
		tokenRefreshJobCancel = nil
	}
	// Stop here if we have no client id or secret.
	if i.ClientID == nil || i.ClientSecret == nil {
		slog.Error("IGDB init client id and or secret not provided")
		return errors.New("client id and or secret not provided")
	}
	slog.Debug("IGDB init running.")
	ctx, cancel := context.WithCancel(context.Background())
	tokenRefreshJobCancel = cancel
	// Get and set first token if needed
	go i.refreshToken(ctx)
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
			videos.video_id,
			similar_games.id,
			similar_games.name,
			similar_games.summary,
			similar_games.cover.image_id,
			similar_games.first_release_date;
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
