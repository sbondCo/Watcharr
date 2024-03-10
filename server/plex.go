package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

type PlexLibraryType string

var (
	MOVIES   PlexLibraryType = "movie"
	TV_SHOWS PlexLibraryType = "show"
)

type PlexResponse struct {
	MediaContainer PlexMediaContainer `json:"MediaContainer"`
}

type PlexMediaContainer struct {
	Directory []PlexDirectory `json:"Directory"`
	Metadata  []PlexMetadata  `json:"Metadata"`
}

type PlexDirectory struct {
	Key  string `json:"key"`
	Type string `json:"type"`
}

type PlexMetadata struct {
	Title        string  `json:"title"`
	Year         int32   `json:"year"`
	ViewCount    int32   `json:"viewCount"`
	LastViewedAt int32   `json:"lastViewedAt"`
	UserRating   float32 `json:"userRating"`
}

// Plex access middleware, ensures user is a Plex user.
// To be ran after AuthRequired middleware with extra data.
func PlexAccessRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		slog.Debug("PlexAccessRequired middleware hit", "user_id", userId)
		userType := c.MustGet("userType").(UserType)
		userThirdPartyAuth := c.MustGet("userThirdPartyAuth").(string)
		if Config.PLEX_OAUTH_ID == "" {
			slog.Error("PlexAccessRequired: Request made to make Plex API call, but PLEX_OAUTH_ID has not been configured.")
			c.AbortWithStatus(401)
			return
		}
		if Config.PLEX_HOST == "" {
			slog.Error("PlexAccessRequired: Request made to make Plex API call, but PLEX_HOST has not been configured.")
			c.AbortWithStatus(401)
			return
		}
		if userType != PLEX_USER {
			slog.Error("PlexAccessRequired: User is not a Plex user..", "user_type", userType)
			c.AbortWithStatus(401)
			return
		}
		if userThirdPartyAuth == "" {
			slog.Error("PlexAccessRequired: User has no thirdPartyAuth token..")
			c.AbortWithStatus(401)
			return
		}
	}
}

func plexAPIRequest(method string, ep string, p map[string]string, userToken string, resp interface{}) error {
	if Config.PLEX_HOST == "" {
		slog.Error("plexAPIRequest: PLEX_HOST not configured.")
		return errors.New("plex sync not enabled")
	}
	slog.Debug("plexAPIRequest", "endpoint", ep, "params", p)
	base, err := url.Parse(Config.PLEX_HOST)
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
		slog.Error("Creating request to plex failed", "error", err)
		return errors.New("request failed")
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("X-Pler-Token", userToken)
	res, err := client.Do(req)
	if err != nil {
		slog.Error("making request to plex failed", "error", err)
		return errors.New("request failed")
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		slog.Error("Error reading plex response", "error", err.Error())
		return err
	}
	if res.StatusCode != 200 {
		slog.Error("Plex non 200 status code", "status_code", res.StatusCode, "error", string(body))
		return errors.New("plex non 200 status code")
	}
	// Unmarshal response
	slog.Info(string(body))
	err = json.Unmarshal([]byte(body), &resp)
	if err != nil {
		return err
	}
	return nil
}

func plexGetLibraries(userThirdPartyAuth string) ([]PlexDirectory, error) {
	resp := new(PlexResponse)
	err := plexAPIRequest("GET", "/library/sections", nil, userThirdPartyAuth, &resp)
	if err != nil {
		slog.Error("plexGetLibraries: Plex Libraries API request failed", "error", err)
		return nil, errors.New("failed to get plex libraries")
	}
	return resp.MediaContainer.Directory, nil
}

func plexGetLibraryItems(userThirdPartyAuth string, library string) ([]PlexMetadata, error) {
	resp := new(PlexResponse)
	err := plexAPIRequest("GET", "/library/sections/"+library+"/all", nil, userThirdPartyAuth, &resp)
	if err != nil {
		slog.Error("plexGetLibraryItems: Plex Library Items API request failed", "error", err)
		return nil, errors.New("failed to get plex library items")
	}

	return resp.MediaContainer.Metadata, nil
}
