package igdb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	gocache "github.com/robfig/go-cache"

	"github.com/sbondCo/Watcharr/cache"
)

// inmemory content cache
var GameStore = gocache.New(time.Hour*24, time.Minute)

const (
	igdbHost       = "https://api.igdb.com/v4"
	tokenGrantType = "client_credentials"

	// Fields that we select for "search" services.
	fieldsForSearch = `name,
cover.image_id,
version_title,
summary,
first_release_date`
)

var tokenRefreshJobCancel context.CancelFunc

type IGDB struct {
	ClientID           *string   `json:"clientId,omitempty"`
	ClientSecret       *string   `json:"clientSecret,omitempty"`
	AccessToken        string    `json:"accessToken,omitempty"`
	AccessTokenExpires time.Time `json:"accessTokenExpires,omitzero"`
	onTokenRefreshed   *func()
}

func (i *IGDB) req(host string, ep string, p map[string]string, b string, resp interface{}) error {
	// if using igdb host and we have no access token, error before running req
	if host == igdbHost && (i.ClientID == nil || i.AccessToken == "") {
		return errors.New("using igdbHost without a clientID or accessToken")
	}

	slog.Debug("IGDB->req: Creating a request.", "ep", ep, "body", b)

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
	slog.Debug("Search:", "query", q)
	var resp GameSearchResponse
	cacheKey := cache.CreateCacheKey("Search", q)
	if cache.GetCache(GameStore, cacheKey, &resp) {
		slog.Debug("Search: Returning cache.")
		return resp, nil
	}
	err := i.req(
		igdbHost,
		"/games",
		map[string]string{},
		"fields "+fieldsForSearch+"; search \""+q+"\"; limit 40;",
		&resp,
	)
	if err != nil {
		slog.Error("Search: request failed!", "error", err)
		return GameSearchResponse{}, errors.New("request failed")
	}
	GameStore.Set(cacheKey, resp, time.Hour*24)
	return resp, nil
}

// Should return same details as `Search`, we both are for search page only minimal details required.
func (i *IGDB) SearchById(id string) (GameSearchResponse, error) {
	slog.Debug("SearchById:", "id", id)
	var resp GameSearchResponse
	cacheKey := cache.CreateCacheKey("SearchById", id)
	if cache.GetCache(GameStore, cacheKey, &resp) {
		slog.Debug("SearchById: Returning cache.")
		return resp, nil
	}
	err := i.req(
		igdbHost,
		"/games",
		map[string]string{},
		"fields "+fieldsForSearch+"; where id = "+id+";",
		&resp,
	)
	if err != nil {
		slog.Error("SearchById: request failed!", "error", err)
		return GameSearchResponse{}, errors.New("request failed")
	}
	GameStore.Set(cacheKey, resp, time.Hour*24)
	return resp, nil
}

// Should return same details as `Search`, both are for search page only minimal details required.
func (i *IGDB) SearchBySlug(slug string) (GameSearchResponse, error) {
	slog.Debug("SearchBySlug: Will search.", "slug", slug)
	var resp GameSearchResponse
	cacheKey := cache.CreateCacheKey("SearchBySlug", slug)
	if cache.GetCache(GameStore, cacheKey, &resp) {
		slog.Debug("SearchBySlug: Returning cache.")
		return resp, nil
	}
	err := i.req(
		igdbHost,
		"/games",
		map[string]string{},
		"fields "+fieldsForSearch+"; where slug = \""+slug+"\";",
		&resp,
	)
	if err != nil {
		slog.Error("SearchBySlug: Request failed!", "error", err)
		return GameSearchResponse{}, errors.New("request failed")
	}
	GameStore.Set(cacheKey, resp, time.Hour*24)
	return resp, nil
}

func (i *IGDB) GameDetails(id string) (GameDetailsResponse, error) {
	slog.Debug("GameDetails:", "id", id)
	var resp []GameDetailsResponse
	cacheKey := cache.CreateCacheKey("GameDetails", id)
	if cache.GetCache(GameStore, cacheKey, &resp) {
		slog.Debug("GameDetails: Returning cache.")
		if len(resp) > 0 {
			return resp[0], nil
		}
		return GameDetailsResponse{}, errors.New("no game details recieved")
	}
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
			game_type,
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
			involved_companies.company.websites.type,
			involved_companies.company.websites.trusted,
			involved_companies.company.websites.url,
			rating,
			rating_count,
			status,
			url,
			websites.trusted,
			websites.type,
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
		slog.Error("GameDetails: Request failed!", "error", err)
		return GameDetailsResponse{}, errors.New("request failed")
	}
	GameStore.Set(cacheKey, resp, time.Hour*24)
	if len(resp) > 0 {
		return resp[0], nil
	}
	return GameDetailsResponse{}, errors.New("no game details recieved")
}

// Basic game details for when we are using them only to update our cache.
// In these cases, it's a waste to ask for everything, when we don't need it.
func (i *IGDB) GameDetailsBasic(id string) (GameDetailsBasicResponse, error) {
	slog.Debug("GameDetailsBasic:", "id", id)
	var resp []GameDetailsBasicResponse
	cacheKey := cache.CreateCacheKey("GameDetailsBasic", id)
	if cache.GetCache(GameStore, cacheKey, &resp) {
		slog.Debug("GameDetailsBasic: Returning cache.")
		if len(resp) > 0 {
			return resp[0], nil
		}
		return GameDetailsBasicResponse{}, errors.New("no game details recieved")
	}
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
	GameStore.Set(cacheKey, resp, time.Hour*24)
	if len(resp) > 0 {
		return resp[0], nil
	}
	return GameDetailsBasicResponse{}, errors.New("no game details recieved")
}

// Getting top visited games in last 24hrs with popscore.
// There doesn't seem to be a way to get game details directly in this request
// so it must be combined with another.
func (i *IGDB) GetTopVisitedGameIds() (PopularityPrimitivesGameIdsResponse, error) {
	slog.Debug("GetTopVisitedGameIds: Running.")
	var resp PopularityPrimitivesGameIdsResponse
	cacheKey := cache.CreateCacheKey("GetTopVisitedGameIds")
	if cache.GetCache(GameStore, cacheKey, &resp) {
		slog.Debug("GetTopVisitedGameIds: Returning cache.")
		return resp, nil
	}
	err := i.req(
		igdbHost,
		"/popularity_primitives",
		map[string]string{},
		`fields game_id;
sort value desc;
where popularity_type = 1;
limit 40;`,
		&resp,
	)
	if err != nil {
		slog.Error("GetTopVisitedGameIds: request failed!", "error", err)
		return resp, errors.New("request failed")
	}
	GameStore.Set(cacheKey, resp, time.Hour*24)
	if len(resp) <= 0 {
		slog.Error("GetTopVisitedGameIds: We got zero results!"+
			"Something must be wrong.",
			"error", err)
		return resp, errors.New("got zero results")
	}
	return resp, nil
}

// Trending games using popscores (which suck btw.. they dont expose that nice
// Popular Right Now list you can see on igdb homepage :( uh)
func (i *IGDB) Trending() (GameSearchResponse, error) {
	slog.Debug("Trending: Running.")
	var resp GameSearchResponse

	// Get game ids.
	ids, err := i.GetTopVisitedGameIds()
	if err != nil {
		return resp, errors.New("failed to get game ids")
	}

	// Get ids in a string we can pass to igdb.
	idsStr := ""
	for _, v := range ids {
		idsStr = idsStr + strconv.Itoa(v.GameID) + ","
	}
	idsStr = strings.TrimSuffix(idsStr, ",")

	cacheKey := cache.CreateCacheKey("Trending")
	if cache.GetCache(GameStore, cacheKey, &resp) {
		slog.Debug("Trending: Returning cache.")
		return resp, nil
	}
	err = i.req(
		igdbHost,
		"/games",
		map[string]string{},
		"fields "+fieldsForSearch+"; where id = ("+idsStr+"); limit 40;",
		&resp,
	)
	if err != nil {
		slog.Error("Trending: request failed!", "error", err)
		return resp, errors.New("request failed")
	}
	GameStore.Set(cacheKey, resp, time.Hour*24)
	return resp, nil
}

// Most hyped upcoming games
func (i *IGDB) Upcoming() (GameSearchResponse, error) {
	slog.Debug("Upcoming: Running.")
	var resp GameSearchResponse
	cacheKey := cache.CreateCacheKey("Upcoming")
	if cache.GetCache(GameStore, cacheKey, &resp) {
		slog.Debug("Upcoming: Returning cache.")
		return resp, nil
	}
	epochNow := strconv.Itoa(int(time.Now().Unix()))
	err := i.req(
		igdbHost,
		"/games",
		map[string]string{},
		"fields "+fieldsForSearch+`;
sort hypes desc;
where first_release_date > `+epochNow+`;
limit 40;`,
		&resp,
	)
	if err != nil {
		slog.Error("Upcoming: request failed!", "error", err)
		return resp, errors.New("request failed")
	}
	GameStore.Set(cacheKey, resp, time.Hour*24)
	return resp, nil
}
