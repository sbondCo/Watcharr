package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PlexLoginRequest struct {
	AuthToken string `json:"token" binding:"required"`
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

type PlexUsersResponse struct {
	XMLName           xml.Name `xml:"MediaContainer"`
	Text              string   `xml:",chardata"`
	FriendlyName      string   `xml:"friendlyName,attr"`
	Identifier        string   `xml:"identifier,attr"`
	MachineIdentifier string   `xml:"machineIdentifier,attr"`
	TotalSize         string   `xml:"totalSize,attr"`
	Size              string   `xml:"size,attr"`
	User              []struct {
		Text                      string `xml:",chardata"`
		ID                        string `xml:"id,attr"`
		Title                     string `xml:"title,attr"`
		Username                  string `xml:"username,attr"`
		Email                     string `xml:"email,attr"`
		RecommendationsPlaylistId string `xml:"recommendationsPlaylistId,attr"`
		Thumb                     string `xml:"thumb,attr"`
		Protected                 string `xml:"protected,attr"`
		Home                      string `xml:"home,attr"`
		AllowTuners               string `xml:"allowTuners,attr"`
		AllowSync                 string `xml:"allowSync,attr"`
		AllowCameraUpload         string `xml:"allowCameraUpload,attr"`
		AllowChannels             string `xml:"allowChannels,attr"`
		AllowSubtitleAdmin        string `xml:"allowSubtitleAdmin,attr"`
		FilterAll                 string `xml:"filterAll,attr"`
		FilterMovies              string `xml:"filterMovies,attr"`
		FilterMusic               string `xml:"filterMusic,attr"`
		FilterPhotos              string `xml:"filterPhotos,attr"`
		FilterTelevision          string `xml:"filterTelevision,attr"`
		Restricted                string `xml:"restricted,attr"`
		Server                    []struct {
			Text              string `xml:",chardata"`
			ID                string `xml:"id,attr"`
			ServerId          string `xml:"serverId,attr"`
			MachineIdentifier string `xml:"machineIdentifier,attr"`
			Name              string `xml:"name,attr"`
			LastSeenAt        string `xml:"lastSeenAt,attr"`
			NumLibraries      string `xml:"numLibraries,attr"`
			AllLibraries      string `xml:"allLibraries,attr"`
			Owned             string `xml:"owned,attr"`
			Pending           string `xml:"pending,attr"`
		} `xml:"Server"`
	} `xml:"User"`
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
		ViewMode            int    `json:"viewMode"`
		Metadata            []struct {
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
		ViewMode            int    `json:"viewMode"`
		Directory           []struct {
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
		ViewMode            int    `json:"viewMode"`
		Metadata            []struct {
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

// Plex access middleware, ensures user is a Plex user.
// To be ran after AuthRequired middleware with extra data.
func PlexAccessRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		slog.Debug("PlexAccessRequired middleware hit", "user_id", userId)
		userType := c.MustGet("userType").(UserType)
		userThirdPartyId := c.MustGet("userThirdPartyId").(string)
		userThirdPartyAuth := c.MustGet("userThirdPartyAuth").(string)
		if Config.PLEX_HOST == "" || Config.PLEX_MACHINE_ID == "" {
			slog.Error("PlexAccessRequired: Plex has not been configured.", "user_id", userId)
			c.AbortWithStatus(401)
			return
		}
		if userType != PLEX_USER || userThirdPartyId == "" {
			slog.Error("PlexAccessRequired: User is not a Plex user..", "user_id", userId, "user_type", userType, "user_third_party_id", userThirdPartyId)
			c.AbortWithStatus(401)
			return
		}
		if userThirdPartyAuth == "" {
			slog.Error("PlexAccessRequired: User has no thirdPartyAuth token..", "user_id", userId)
			c.AbortWithStatus(401)
			return
		}
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
	slog.Info(string(body))
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

// If a plex user has access to our home plex server (PLEX_HOST).
func plexUserHasAccessToPlexHost(token string) error {
	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", "https://plex.tv/api/users", nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Plex-Token", token)
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var pa PlexUsersResponse
	err = xml.Unmarshal(body, &pa)
	if err != nil {
		return err
	}

	if len(pa.User) <= 0 {
		return errors.New("found no users in response")
	}

	// Now check if any of the users servers include our home server machine id
	homeServerFound := false
userLoop:
	for _, user := range pa.User {
		for _, server := range user.Server {
			if server.MachineIdentifier == Config.PLEX_MACHINE_ID {
				slog.Debug("plexUserHasAccessToPlexHost: Processing a server.", "server", server.MachineIdentifier)
				homeServerFound = true
				break userLoop
			}
		}
		if homeServerFound {
			break
		}
	}
	if homeServerFound {
		return nil
	}
	return errors.New("user does not have access to home plex server")
}

func getPlexLibraries(userThirdPartyAuth string) (PlexLibrariesResponse, error) {
	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", Config.PLEX_HOST+"/library/sections", nil)
	if err != nil {
		return PlexLibrariesResponse{}, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Set("X-Plex-Token", userThirdPartyAuth)
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

func getPlexLibraryItems(userThirdPartyAuth string, libraryKey string) (PlexLibraryItemsResponse, error) {
	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", Config.PLEX_HOST+"/library/sections/"+libraryKey+"/all?includeGuids=1", nil)
	if err != nil {
		return PlexLibraryItemsResponse{}, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Set("X-Plex-Token", userThirdPartyAuth)
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

func getPlexLibraryItemSeasons(userThirdPartyAuth string, ratingKey string) (PlexLibraryItemSeasonsResponse, error) {
	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", Config.PLEX_HOST+"/library/metadata/"+ratingKey+"/children", nil)
	if err != nil {
		return PlexLibraryItemSeasonsResponse{}, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Set("X-Plex-Token", userThirdPartyAuth)
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

func getPlexLibraryItemEpisodes(userThirdPartyAuth string, ratingKey string) (PlexLibraryItemEpisodesResponse, error) {
	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", Config.PLEX_HOST+"/library/metadata/"+ratingKey+"/allLeaves", nil)
	if err != nil {
		return PlexLibraryItemEpisodesResponse{}, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Set("X-Plex-Token", userThirdPartyAuth)
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
