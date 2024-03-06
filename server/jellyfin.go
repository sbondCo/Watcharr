package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
)

type JellyfinItemSearchResponse struct {
	Items []JellyfinItems `json:"Items"`
}

type JellyfinItems struct {
	Name        string `json:"Name"`
	Type        string `json:"Type"`
	ServerID    string `json:"ServerId"`
	Id          string `json:"Id"`
	ProviderIds struct {
		Tmdb string `json:"Tmdb"`
	} `json:"ProviderIds"`
	UserData struct {
		Rating                float64   `json:"Rating"`
		PlayedPercentage      float64   `json:"PlayedPercentage"`
		UnplayedItemCount     int64     `json:"UnplayedItemCount"`
		PlaybackPositionTicks int64     `json:"PlaybackPositionTicks"`
		PlayCount             int64     `json:"PlayCount"`
		IsFavorite            bool      `json:"IsFavorite"`
		Likes                 bool      `json:"Likes"`
		LastPlayedDate        time.Time `json:"LastPlayedDate"`
		Played                bool      `json:"Played"`
		Key                   string    `json:"Key"`
		ItemId                string    `json:"ItemId"`
	} `json:"UserData"`
	RecursiveItemCount int64 `json:"RecursiveItemCount"`
}

type JFContentFindResponse struct {
	HasContent bool   `json:"hasContent"`
	Url        string `json:"url"`
}

// Jellyfin access middleware, ensures user is a jellyfin user.
// To be ran after AuthRequired middleware with extra data.
func JellyfinAccessRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		slog.Debug("JellyfinAccessRequired middleware hit", "user_id", userId)
		userType := c.MustGet("userType").(UserType)
		userThirdPartyId := c.MustGet("userThirdPartyId").(string)
		userThirdPartyAuth := c.MustGet("userThirdPartyAuth").(string)
		if Config.JELLYFIN_HOST == "" {
			slog.Error("JellyfinAccessRequired: Request made to login via Jellyfin, but JELLYFIN_HOST has not been configured.")
			c.AbortWithStatus(401)
			return
		}
		if userType != JELLYFIN_USER || userThirdPartyId == "" {
			slog.Error("JellyfinAccessRequired: User is not a jellyfin user..", "user_type", userType, "user_third_party_id", userThirdPartyId)
			c.AbortWithStatus(401)
			return
		}
		if userThirdPartyAuth == "" {
			slog.Error("JellyfinAccessRequired: User has no thirdPartyAuth token..")
			c.AbortWithStatus(401)
			return
		}
	}
}

func jellyfinAPIRequest(method string, ep string, p map[string]string, username string, userToken string, resp interface{}) error {
	if Config.JELLYFIN_HOST == "" {
		slog.Error("jellyfinAPIRequest: JELLYFIN_HOST not configured.")
		return errors.New("jellyfin not enabled")
	}
	slog.Debug("jellyfinAPIRequest", "endpoint", ep, "params", p)
	base, err := url.Parse(Config.JELLYFIN_HOST)
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

	// Run get request
	client := &http.Client{}
	req, err := http.NewRequest(method, base.String(), bytes.NewBuffer([]byte{}))
	if err != nil {
		slog.Error("Creating request to jellyfin failed", "error", err)
		return errors.New("request failed")
	}
	req.Header.Add("Content-Type", "application/json")
	authHeader := "MediaBrowser Client=\"Watcharr\", Device=\"HTTP\", DeviceId=\"WatcharrFor" + username + "\", Version=\"10.8.0\""
	if userToken != "" {
		authHeader += ", Token=\"" + userToken + "\""
	}
	req.Header.Add("X-Emby-Authorization", authHeader)
	res, err := client.Do(req)
	if err != nil {
		slog.Error("making request to jellyfin for auth failed", "error", err)
		return errors.New("request failed")
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		slog.Error("Error reading jellyfin auth response", "error", err.Error())
		return err
	}
	if res.StatusCode != 200 {
		slog.Error("Jellyfin auth non 200 status code", "status_code", res.StatusCode, "error", string(body))
		return errors.New("incorrect details")
	}
	// Unmarshal response
	err = json.Unmarshal([]byte(body), &resp)
	if err != nil {
		return err
	}
	return nil
}

func jellyfinContentFind(
	userId uint,
	userType UserType,
	username string,
	userThirdPartyId string,
	userThirdPartyAuth string,
	contentType string,
	contentName string,
	contentTmdbId string,
) (JFContentFindResponse, error) {
	if contentType == "" || contentName == "" {
		slog.Error("Bad request", "content_type", contentType, "content_name", contentName)
		return JFContentFindResponse{}, errors.New("content type or name not provided")
	}
	if contentType == "tv" {
		contentType = "series"
	}

	resp := new(JellyfinItemSearchResponse)
	err := jellyfinAPIRequest(
		"GET",
		"/Users/"+userThirdPartyId+"/Items",
		map[string]string{
			"searchTerm":             contentName,
			"IncludePeople":          "false",
			"IncludeMedia":           "true",
			"IncludeGenres":          "false",
			"IncludeStudios":         "false",
			"IncludeArtists":         "false",
			"IncludeItemTypes":       contentType,
			"Limit":                  "5",
			"Fields":                 "ProviderIds",
			"Recursive":              "true",
			"EnableTotalRecordCount": "false",
			"ImageTypeLimit":         "0",
		},
		username,
		userThirdPartyAuth,
		&resp,
	)
	if err != nil {
		slog.Error("jellyfinContentFind: Jellyfin API request failed", "error", err)
		return JFContentFindResponse{}, errors.New("failed to get jellyfin response")
	}

	// Find true match from jf search results
	ret := new(JFContentFindResponse)
	ret.HasContent = false
	ret.Url = ""
	for _, i := range resp.Items {
		if i.ProviderIds.Tmdb == contentTmdbId {
			ret.HasContent = true
			ret.Url = Config.JELLYFIN_HOST + "/web/index.html#!/details?id=" + i.Id + "&serverId=" + i.ServerID
		}
	}
	return *ret, nil
}
