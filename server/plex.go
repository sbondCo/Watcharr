package main

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PlexLoginRequest struct {
	AuthToken        string `json:"token" binding:"required"`
	ClientIdentifier string `json:"clientIdentifier" binding:"required"`
}

type PlexUser struct {
	Id       uint64 `json:"id" binding:"required"`
	Uuid     string `json:"uuid" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Username string `json:"username" binding:"required"`
}

type PlexAccountResponse struct {
	User PlexUser `json:"user" binding:"required"`
}

type PlexIdentity struct {
	MediaContainer struct {
		MachineIdentifier string `json:"machineIdentifier"`
	} `json:"MediaContainer"`
}

type PlexHostConfigUpdateResponse struct {
	PLEX_MACHINE_ID string
}

// Plex get libraries response
// /library/sections
type PlexLibrariesResponse struct {
	MediaContainer struct {
		Size      int    `json:"size"`
		AllowSync bool   `json:"allowSync"`
		Title1    string `json:"title1"`
		Directory []struct {
			AllowSync        bool   `json:"allowSync"`
			Art              string `json:"art"`
			Composite        string `json:"composite"`
			Filters          bool   `json:"filters"`
			Refreshing       bool   `json:"refreshing"`
			Thumb            string `json:"thumb"`
			Key              string `json:"key"`
			Type             string `json:"type"`
			Title            string `json:"title"`
			Agent            string `json:"agent"`
			Scanner          string `json:"scanner"`
			Language         string `json:"language"`
			UUID             string `json:"uuid"`
			UpdatedAt        int    `json:"updatedAt"`
			CreatedAt        int    `json:"createdAt"`
			ScannedAt        int    `json:"scannedAt"`
			Content          bool   `json:"content"`
			Directory        bool   `json:"directory"`
			ContentChangedAt int64  `json:"contentChangedAt"`
			Hidden           int    `json:"hidden"`
			Location         []struct {
				ID   int    `json:"id"`
				Path string `json:"path"`
			} `json:"Location"`
		} `json:"Directory"`
	} `json:"MediaContainer"`
}

// Plex get all library items response (with includeGuids parameter)
// /library/sections/{sectionId}/all
type PlexLibraryItemsResponse struct {
	MediaContainer struct {
		Size                int    `json:"size"`
		AllowSync           bool   `json:"allowSync"`
		Art                 string `json:"art"`
		Identifier          string `json:"identifier"`
		LibrarySectionID    int    `json:"librarySectionID"`
		LibrarySectionTitle string `json:"librarySectionTitle"`
		LibrarySectionUUID  string `json:"librarySectionUUID"`
		MediaTagPrefix      string `json:"mediaTagPrefix"`
		MediaTagVersion     int    `json:"mediaTagVersion"`
		Thumb               string `json:"thumb"`
		Title1              string `json:"title1"`
		Title2              string `json:"title2"`
		ViewGroup           string `json:"viewGroup"`
		// ViewMode            int    `json:"viewMode"` // Causing string error, not used so commented out for now.
		Metadata []struct {
			RatingKey              string  `json:"ratingKey"`
			Key                    string  `json:"key"`
			GUID                   string  `json:"guid"` // Plex guid
			Slug                   string  `json:"slug"`
			Studio                 string  `json:"studio"`
			Type                   string  `json:"type"`
			Title                  string  `json:"title"`
			ContentRating          string  `json:"contentRating"`
			Summary                string  `json:"summary"`
			Rating                 float64 `json:"rating"`
			UserRating             float64 `json:"userRating"`
			AudienceRating         float64 `json:"audienceRating"`
			ViewCount              int     `json:"viewCount,omitempty"`
			LastViewedAt           int64   `json:"lastViewedAt,omitempty"`
			Year                   int     `json:"year"`
			Tagline                string  `json:"tagline"`
			Thumb                  string  `json:"thumb"`
			Art                    string  `json:"art"`
			Duration               int     `json:"duration"`
			OriginallyAvailableAt  string  `json:"originallyAvailableAt"`
			AddedAt                int     `json:"addedAt"`
			UpdatedAt              int     `json:"updatedAt"`
			AudienceRatingImage    string  `json:"audienceRatingImage"`
			HasPremiumPrimaryExtra string  `json:"hasPremiumPrimaryExtra"`
			RatingImage            string  `json:"ratingImage"`
			LeafCount              int     `json:"leafCount,omitempty"`
			ViewedLeafCount        int     `json:"viewedLeafCount,omitempty"`
			Media                  []struct {
				ID              int     `json:"id"`
				Duration        int     `json:"duration"`
				Bitrate         int     `json:"bitrate"`
				Width           int     `json:"width"`
				Height          int     `json:"height"`
				AspectRatio     float64 `json:"aspectRatio"`
				AudioChannels   int     `json:"audioChannels"`
				AudioCodec      string  `json:"audioCodec"`
				VideoCodec      string  `json:"videoCodec"`
				VideoResolution string  `json:"videoResolution"`
				Container       string  `json:"container"`
				VideoFrameRate  string  `json:"videoFrameRate"`
				VideoProfile    string  `json:"videoProfile"`
				Part            []struct {
					ID           int    `json:"id"`
					Key          string `json:"key"`
					Duration     int    `json:"duration"`
					File         string `json:"file"`
					Size         int64  `json:"size"`
					Container    string `json:"container"`
					HasThumbnail string `json:"hasThumbnail"`
					VideoProfile string `json:"videoProfile"`
				} `json:"Part"`
			} `json:"Media"`
			// External ids
			Guid []struct {
				ID string `json:"id"`
			} `json:"Guid"`
			Genre []struct {
				Tag string `json:"tag"`
			} `json:"Genre"`
			Country []struct {
				Tag string `json:"tag"`
			} `json:"Country"`
			Director []struct {
				Tag string `json:"tag"`
			} `json:"Director"`
			Writer []struct {
				Tag string `json:"tag"`
			} `json:"Writer"`
			Role []struct {
				Tag string `json:"tag"`
			} `json:"Role"`
			ChapterSource string `json:"chapterSource,omitempty"`
		} `json:"Metadata"`
	} `json:"MediaContainer"`
}

// Plex get all children (seasons) for a serie.
// /library/metadata/{ratingKey}/children
type PlexLibraryItemSeasonsResponse struct {
	MediaContainer struct {
		Size                int    `json:"size"`
		AllowSync           bool   `json:"allowSync"`
		Art                 string `json:"art"`
		Identifier          string `json:"identifier"`
		Key                 string `json:"key"`
		LibrarySectionID    int    `json:"librarySectionID"`
		LibrarySectionTitle string `json:"librarySectionTitle"`
		LibrarySectionUUID  string `json:"librarySectionUUID"`
		MediaTagPrefix      string `json:"mediaTagPrefix"`
		MediaTagVersion     int    `json:"mediaTagVersion"`
		Nocache             bool   `json:"nocache"`
		ParentIndex         int    `json:"parentIndex"`
		ParentTitle         string `json:"parentTitle"`
		ParentYear          int    `json:"parentYear"`
		Summary             string `json:"summary"`
		Theme               string `json:"theme"`
		Thumb               string `json:"thumb"`
		Title1              string `json:"title1"`
		Title2              string `json:"title2"`
		ViewGroup           string `json:"viewGroup"`
		// ViewMode            int    `json:"viewMode"` // Causing string error, not used so commented out for now.
		Directory []struct {
			LeafCount       int    `json:"leafCount"`
			Thumb           string `json:"thumb"`
			ViewedLeafCount int    `json:"viewedLeafCount"`
			Key             string `json:"key"`
			Title           string `json:"title"`
		} `json:"Directory"`
		Metadata []struct {
			RatingKey       string  `json:"ratingKey"`
			Key             string  `json:"key"`
			ParentRatingKey string  `json:"parentRatingKey"`
			GUID            string  `json:"guid"`
			ParentGUID      string  `json:"parentGuid"`
			ParentSlug      string  `json:"parentSlug"`
			ParentStudio    string  `json:"parentStudio"`
			Type            string  `json:"type"`
			Title           string  `json:"title"`
			ParentKey       string  `json:"parentKey"`
			ParentTitle     string  `json:"parentTitle"`
			Summary         string  `json:"summary"`
			Index           int     `json:"index"`
			ParentIndex     int     `json:"parentIndex"`
			UserRating      float64 `json:"userRating,omitempty"`
			LastRatedAt     int64   `json:"lastRatedAt,omitempty"`
			ParentYear      int     `json:"parentYear"`
			Thumb           string  `json:"thumb"`
			Art             string  `json:"art"`
			ParentThumb     string  `json:"parentThumb"`
			ParentTheme     string  `json:"parentTheme"`
			LeafCount       int     `json:"leafCount"`
			ViewedLeafCount int     `json:"viewedLeafCount"`
			AddedAt         int     `json:"addedAt"`
			UpdatedAt       int     `json:"updatedAt"`
			ViewCount       int     `json:"viewCount,omitempty"`
			LastViewedAt    int64   `json:"lastViewedAt,omitempty"`
		} `json:"Metadata"`
	} `json:"MediaContainer"`
}

// Plex get all children (episodes) for a season.
// /library/metadata/{ratingKey}/allLeaves
type PlexLibraryItemEpisodesResponse struct {
	MediaContainer struct {
		Size                int    `json:"size"`
		AllowSync           bool   `json:"allowSync"`
		Art                 string `json:"art"`
		Identifier          string `json:"identifier"`
		Key                 string `json:"key"`
		LibrarySectionID    int    `json:"librarySectionID"`
		LibrarySectionTitle string `json:"librarySectionTitle"`
		LibrarySectionUUID  string `json:"librarySectionUUID"`
		MediaTagPrefix      string `json:"mediaTagPrefix"`
		MediaTagVersion     int    `json:"mediaTagVersion"`
		MixedParents        bool   `json:"mixedParents"`
		Nocache             bool   `json:"nocache"`
		ParentIndex         int    `json:"parentIndex"`
		ParentTitle         string `json:"parentTitle"`
		ParentYear          int    `json:"parentYear"`
		Theme               string `json:"theme"`
		Title1              string `json:"title1"`
		Title2              string `json:"title2"`
		ViewGroup           string `json:"viewGroup"`
		// ViewMode            int    `json:"viewMode"` // Causing string error, not used so commented out for now.
		Metadata []struct {
			RatingKey             string  `json:"ratingKey"`
			Key                   string  `json:"key"`
			ParentRatingKey       string  `json:"parentRatingKey"`
			GrandparentRatingKey  string  `json:"grandparentRatingKey"`
			GUID                  string  `json:"guid"`
			GrandparentSlug       string  `json:"grandparentSlug"`
			Studio                string  `json:"studio"`
			Type                  string  `json:"type"`
			Title                 string  `json:"title"`
			GrandparentKey        string  `json:"grandparentKey"`
			ParentKey             string  `json:"parentKey"`
			GrandparentTitle      string  `json:"grandparentTitle"`
			ParentTitle           string  `json:"parentTitle"`
			ContentRating         string  `json:"contentRating"`
			Summary               string  `json:"summary"`
			Index                 int     `json:"index"`
			ParentIndex           int     `json:"parentIndex"`
			AudienceRating        float64 `json:"audienceRating"`
			Year                  int     `json:"year"`
			Thumb                 string  `json:"thumb"`
			Art                   string  `json:"art"`
			ParentThumb           string  `json:"parentThumb"`
			GrandparentThumb      string  `json:"grandparentThumb"`
			GrandparentArt        string  `json:"grandparentArt"`
			GrandparentTheme      string  `json:"grandparentTheme"`
			Duration              int     `json:"duration"`
			OriginallyAvailableAt string  `json:"originallyAvailableAt"`
			AddedAt               int     `json:"addedAt"`
			UpdatedAt             int     `json:"updatedAt"`
			AudienceRatingImage   string  `json:"audienceRatingImage"`
			ChapterSource         string  `json:"chapterSource"`
			TitleSort             string  `json:"titleSort,omitempty"`
			UserRating            float64 `json:"userRating,omitempty"`
			ViewCount             int     `json:"viewCount,omitempty"`
			LastViewedAt          int64   `json:"lastViewedAt,omitempty"`
			LastRatedAt           int64   `json:"lastRatedAt,omitempty"`
		} `json:"Metadata"`
	} `json:"MediaContainer"`
}

// Response from clients.plex.tv/api/v2/resources
// Used to get users auth token for home plex server.
type PlexClientResources []struct {
	Name                   string      `json:"name"`
	Product                string      `json:"product"`
	ProductVersion         string      `json:"productVersion"`
	Platform               string      `json:"platform"`
	PlatformVersion        string      `json:"platformVersion"`
	Device                 string      `json:"device"`
	ClientIdentifier       string      `json:"clientIdentifier"`
	CreatedAt              time.Time   `json:"createdAt"`
	LastSeenAt             time.Time   `json:"lastSeenAt"`
	Provides               string      `json:"provides"`
	OwnerID                interface{} `json:"ownerId"`
	SourceTitle            interface{} `json:"sourceTitle"`
	PublicAddress          string      `json:"publicAddress"`
	AccessToken            string      `json:"accessToken"`
	Owned                  bool        `json:"owned"`
	Home                   bool        `json:"home"`
	Synced                 bool        `json:"synced"`
	Relay                  bool        `json:"relay"`
	Presence               bool        `json:"presence"`
	HTTPSRequired          bool        `json:"httpsRequired"`
	PublicAddressMatches   bool        `json:"publicAddressMatches"`
	DNSRebindingProtection bool        `json:"dnsRebindingProtection"`
	NatLoopbackSupported   bool        `json:"natLoopbackSupported"`
	Connections            []struct {
		Protocol string `json:"protocol"`
		Address  string `json:"address"`
		Port     int    `json:"port"`
		URI      string `json:"uri"`
		Local    bool   `json:"local"`
		Relay    bool   `json:"relay"`
		IPv6     bool   `json:"IPv6"`
	} `json:"connections"`
}

// Plex access middleware, ensures user is a Plex user.
// To be ran after AuthRequired middleware with extra data.
func PlexAccessRequired(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		slog.Debug("PlexAccessRequired middleware hit", "user_id", userId)
		userType := c.MustGet("userType").(UserType)
		if Config.PLEX_HOST == "" || Config.PLEX_MACHINE_ID == "" {
			slog.Error("PlexAccessRequired: Plex has not been configured.", "user_id", userId)
			c.AbortWithStatus(401)
			return
		}
		if userType != PLEX_USER {
			slog.Error("PlexAccessRequired: User is not a Plex user..", "user_id", userId, "user_type", userType)
			c.AbortWithStatus(401)
			return
		}
		userPlexService := new(UserServices)
		if res := db.Where("user_id = ? AND name = ?", userId, "plex").Take(&userPlexService); res.Error != nil {
			slog.Error("PlexAccessRequired: Failed when attempting to get users plex service integration..", "user_id", userId, "user_type", userType)
			c.AbortWithStatus(401)
			return
		}
		if userPlexService.ClientID == "" || userPlexService.AuthToken == "" || userPlexService.AuthToken2 == "" {
			slog.Error("PlexAccessRequired: User has missing details from service (clientId, authToken or authToken2)..", "user_id", userId, "client_id", userPlexService.ClientID)
			c.AbortWithStatus(401)
			return
		}
		c.Set("plexAuthToken", userPlexService.AuthToken)
		c.Set("plexLocalAuthToken", userPlexService.AuthToken2)
	}
}

func getPlexIdentity(host string) (PlexIdentity, error) {
	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", host+"/identity", nil)
	if err != nil {
		return PlexIdentity{}, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return PlexIdentity{}, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return PlexIdentity{}, err
	}
	defer resp.Body.Close()
	var pi PlexIdentity
	err = json.Unmarshal(body, &pi)
	if err != nil {
		return PlexIdentity{}, err
	}
	return pi, nil
}

func fetchPlexAccountFromToken(token string) (PlexUser, error) {
	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", "https://plex.tv/users/account.json", nil)
	if err != nil {
		return PlexUser{}, err
	}
	req.Header.Set("X-Plex-Token", token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return PlexUser{}, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return PlexUser{}, err
	}
	defer resp.Body.Close()
	var pa PlexAccountResponse
	err = json.Unmarshal(body, &pa)
	if err != nil {
		return PlexUser{}, err
	}
	return pa.User, nil
}

// Update plex host setting
func updateConfigPlexHost(v string) (PlexHostConfigUpdateResponse, error) {
	Config.PLEX_HOST = v
	if Config.PLEX_HOST != "" {
		pi, err := getPlexIdentity(Config.PLEX_HOST)
		if err != nil {
			slog.Error("updateConfigPlexHost: Failed to get plex server identity!", "error", err)
			return PlexHostConfigUpdateResponse{}, errors.New("failed to get Plex server identity. Please try setting the Plex Host again or setting PLEX_MACHINE_ID manually in your config file")
		}
		if pi.MediaContainer.MachineIdentifier == "" {
			slog.Error("updateConfigPlexHost: Plex server identity response had no machine id!", "response", pi)
			return PlexHostConfigUpdateResponse{}, errors.New("got Plex server identity, but no machine id was found")
		}
		Config.PLEX_MACHINE_ID = pi.MediaContainer.MachineIdentifier
	} else {
		Config.PLEX_MACHINE_ID = ""
	}
	if err := writeConfig(); err != nil {
		slog.Error("updateConfigPlexHost: Failed to write updated config to file!", "err", err)
		return PlexHostConfigUpdateResponse{}, errors.New("failed to write config")
	}
	return PlexHostConfigUpdateResponse{PLEX_MACHINE_ID: Config.PLEX_MACHINE_ID}, nil
}

func getPlexLibraries(plexAuth string) (PlexLibrariesResponse, error) {
	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", Config.PLEX_HOST+"/library/sections", nil)
	if err != nil {
		return PlexLibrariesResponse{}, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Set("X-Plex-Token", plexAuth)
	resp, err := httpClient.Do(req)
	if err != nil {
		return PlexLibrariesResponse{}, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return PlexLibrariesResponse{}, err
	}
	defer resp.Body.Close()
	var pl PlexLibrariesResponse
	err = json.Unmarshal(body, &pl)
	if err != nil {
		return PlexLibrariesResponse{}, err
	}
	return pl, nil
}

func getPlexLibraryItems(plexAuth string, libraryKey string) (PlexLibraryItemsResponse, error) {
	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", Config.PLEX_HOST+"/library/sections/"+libraryKey+"/all?includeGuids=1", nil)
	if err != nil {
		return PlexLibraryItemsResponse{}, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Set("X-Plex-Token", plexAuth)
	resp, err := httpClient.Do(req)
	if err != nil {
		return PlexLibraryItemsResponse{}, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return PlexLibraryItemsResponse{}, err
	}
	defer resp.Body.Close()
	var pl PlexLibraryItemsResponse
	err = json.Unmarshal(body, &pl)
	if err != nil {
		return PlexLibraryItemsResponse{}, err
	}
	return pl, nil
}

func getPlexLibraryItemSeasons(plexAuth string, ratingKey string) (PlexLibraryItemSeasonsResponse, error) {
	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", Config.PLEX_HOST+"/library/metadata/"+ratingKey+"/children", nil)
	if err != nil {
		return PlexLibraryItemSeasonsResponse{}, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Set("X-Plex-Token", plexAuth)
	resp, err := httpClient.Do(req)
	if err != nil {
		return PlexLibraryItemSeasonsResponse{}, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return PlexLibraryItemSeasonsResponse{}, err
	}
	defer resp.Body.Close()
	var pl PlexLibraryItemSeasonsResponse
	err = json.Unmarshal(body, &pl)
	if err != nil {
		return PlexLibraryItemSeasonsResponse{}, err
	}
	return pl, nil
}

func getPlexLibraryItemEpisodes(plexAuth string, ratingKey string) (PlexLibraryItemEpisodesResponse, error) {
	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", Config.PLEX_HOST+"/library/metadata/"+ratingKey+"/allLeaves", nil)
	if err != nil {
		return PlexLibraryItemEpisodesResponse{}, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Set("X-Plex-Token", plexAuth)
	resp, err := httpClient.Do(req)
	if err != nil {
		return PlexLibraryItemEpisodesResponse{}, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return PlexLibraryItemEpisodesResponse{}, err
	}
	defer resp.Body.Close()
	var pl PlexLibraryItemEpisodesResponse
	err = json.Unmarshal(body, &pl)
	if err != nil {
		return PlexLibraryItemEpisodesResponse{}, err
	}
	return pl, nil
}

// Gets users auth token for local plex server,
// so they can authenticate against it for api requests.
// If no auth token is returned or errored, assume user doesn't have access to home plex server library.
func getPlexHomeServerAuthToken(plexAuth string, userClientId string) (string, error) {
	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", "https://clients.plex.tv/api/v2/resources", nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Set("X-Plex-Token", plexAuth)
	req.Header.Set("X-Plex-Client-Identifier", userClientId)
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var pl PlexClientResources
	err = json.Unmarshal(body, &pl)
	if err != nil {
		return "", err
	}
	authToken := ""
	for _, v := range pl {
		if v.ClientIdentifier == Config.PLEX_MACHINE_ID {
			slog.Debug("getPlexHomeServerAuthToken: Found entry with clientIdentifier matching home server machine id.")
			if v.AccessToken == "" {
				slog.Error("getPlexHomeServerAuthToken: Matching entry has no AccessToken!")
				continue
			}
			authToken = v.AccessToken
		}
	}
	if authToken == "" {
		slog.Error("getPlexHomeServerAuthToken: No authToken retrieved!")
	}
	return authToken, nil
}
