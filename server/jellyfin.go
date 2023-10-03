package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
)

type JellyfinItemSearchResponse struct {
	Items []JellyfinItems `json:"Items"`
}

type JellyfinItems struct {
	Type        string `json:"Type"`
	ServerID    string `json:"ServerId"`
	Id          string `json:"Id"`
	ProviderIds struct {
		Tmdb string `json:"Tmdb"`
	} `json:"ProviderIds"`
}

type JFContentFindResponse struct {
	HasContent bool   `json:"hasContent"`
	Url        string `json:"url"`
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
	if Config.JELLYFIN_HOST == "" {
		slog.Error("Request made to login via Jellyfin, but JELLYFIN_HOST has not been configured.")
		return JFContentFindResponse{}, errors.New("jellyfin login not enabled")
	}
	if userType != JELLYFIN_USER || userThirdPartyId == "" {
		slog.Error("User is not a jellyfin user..", "user_type", userType, "user_third_party_id", userThirdPartyId)
		return JFContentFindResponse{}, errors.New("not jellyfin user")
	}
	if userThirdPartyAuth == "" {
		slog.Error("User has no thirdPartyAuth token..")
		return JFContentFindResponse{}, errors.New("user has no jellyfin auth token")
	}
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
