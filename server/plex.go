package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"log/slog"
	"net/http"
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
