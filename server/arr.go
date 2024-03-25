package main

import (
	"errors"
	"log/slog"

	"github.com/sbondCo/Watcharr/arr"
)

type ArrSettings struct {
	Name string `json:"name,omitempty"`
	Host string `json:"host,omitempty"`
	Key  string `json:"key,omitempty"`
}

type SonarrSettings struct {
	ArrSettings
	QualityProfile  int  `json:"qualityProfile,omitempty"`
	RootFolder      int  `json:"rootFolder,omitempty"`
	LanguageProfile int  `json:"languageProfile,omitempty"`
	AutomaticSearch bool `json:"automaticSearch"`
	// TODO eventually separate profiles and root for anime content (i can see diff language profile being useful)
}

func (s *SonarrSettings) safe() SonarrSettings {
	s.Key = ""
	return *s
}

type RadarrSettings struct {
	ArrSettings
	QualityProfile  int  `json:"qualityProfile,omitempty"`
	RootFolder      int  `json:"rootFolder,omitempty"`
	AutomaticSearch bool `json:"automaticSearch"`
}

func (s *RadarrSettings) safe() RadarrSettings {
	s.Key = ""
	return *s
}

type ArrTestParams struct {
	Host string `json:"host,omitempty"`
	Key  string `json:"key,omitempty"`
}

type SonarrTestResponse struct {
	QualityProfiles  []arr.QualityProfile  `json:"qualityProfiles"`
	RootFolders      []arr.RootFolder      `json:"rootFolders"`
	LanguageProfiles []arr.LanguageProfile `json:"languageProfiles"`
}

type RadarrTestResponse struct {
	QualityProfiles  []arr.QualityProfile  `json:"qualityProfiles"`
	RootFolders      []arr.RootFolder      `json:"rootFolders"`
	LanguageProfiles []arr.LanguageProfile `json:"languageProfiles"`
}

// Response given to users with PERM_REQUEST_CONTENT - should never include sensitive info
func testSonarr(p ArrTestParams) (SonarrTestResponse, error) {
	sonarr := arr.New(arr.SONARR, &p.Host, &p.Key)
	qps, err := sonarr.GetQualityProfiles()
	if err != nil {
		slog.Error("testSonarr failed to get quality profiles!", "error", err)
		return SonarrTestResponse{}, errors.New("failed to get quality profiles")
	}
	rfs, err := sonarr.GetRootFolders()
	if err != nil {
		slog.Error("testSonarr failed to get root folders!", "error", err)
		return SonarrTestResponse{}, errors.New("failed to get root folders")
	}
	lps, err := sonarr.GetLangaugeProfiles()
	if err != nil {
		slog.Error("testSonarr failed to get language profiles!", "error", err)
		return SonarrTestResponse{}, errors.New("failed to get language profiles")
	}
	return SonarrTestResponse{QualityProfiles: qps, RootFolders: rfs, LanguageProfiles: lps}, nil
}

// Response given to users with PERM_REQUEST_CONTENT - should never include sensitive info
func testRadarr(p ArrTestParams) (RadarrTestResponse, error) {
	radarr := arr.New(arr.RADARR, &p.Host, &p.Key)
	qps, err := radarr.GetQualityProfiles()
	if err != nil {
		slog.Error("testRadarr failed to get quality profiles!", "error", err)
		return RadarrTestResponse{}, errors.New("failed to get quality profiles")
	}
	rfs, err := radarr.GetRootFolders()
	if err != nil {
		slog.Error("testRadarr failed to get root folders!", "error", err)
		return RadarrTestResponse{}, errors.New("failed to get root folders")
	}
	return RadarrTestResponse{QualityProfiles: qps, RootFolders: rfs}, nil
}

// TODO any way to simplify (deduplicate/reuse) these methods (and the whole file tbh) would be very good

// Add sonarr server to config
func addSonarr(s SonarrSettings) error {
	for _, v := range Config.SONARR {
		if v.Name == s.Name {
			// Server exists with this name...
			return errors.New("server with that name already exists")
		}
	}
	Config.SONARR = append(Config.SONARR, s)
	writeConfig()
	return nil
}

// Edit sonarr server in config
func editSonarr(s SonarrSettings) error {
	for i, v := range Config.SONARR {
		if v.Name == s.Name {
			Config.SONARR[i] = s
			writeConfig()
			return nil
		}
	}
	return errors.New("can't edit server that does not exist")
}

func rmSonarr(name string) error {
	for i, v := range Config.SONARR {
		if v.Name == name {
			Config.SONARR = append(Config.SONARR[:i], Config.SONARR[i+1:]...)
			writeConfig()
			return nil
		}
	}
	return errors.New("can't remove a server that does not exist")
}

func getSonarr(name string) (SonarrSettings, error) {
	for i, v := range Config.SONARR {
		if v.Name == name {
			return Config.SONARR[i], nil
		}
	}
	return SonarrSettings{}, errors.New("server not found")
}

// Get list of sonarr servers without api keys.
// Regular users with access to adding to sonarr will request this.
func getSonarrsSafe() []SonarrSettings {
	s := []SonarrSettings{}
	for _, v := range Config.SONARR {
		s = append(s, v.safe())
	}
	return s
}

// Add radarr server to config
func addRadarr(s RadarrSettings) error {
	for _, v := range Config.RADARR {
		if v.Name == s.Name {
			// Server exists with this name...
			return errors.New("server with that name already exists")
		}
	}
	Config.RADARR = append(Config.RADARR, s)
	writeConfig()
	return nil
}

// Edit radarr server in config
func editRadarr(s RadarrSettings) error {
	for i, v := range Config.RADARR {
		if v.Name == s.Name {
			Config.RADARR[i] = s
			writeConfig()
			return nil
		}
	}
	return errors.New("can't edit server that does not exist")
}

func rmRadarr(name string) error {
	for i, v := range Config.RADARR {
		if v.Name == name {
			Config.RADARR = append(Config.RADARR[:i], Config.RADARR[i+1:]...)
			writeConfig()
			return nil
		}
	}
	return errors.New("can't remove a server that does not exist")
}

func getRadarr(name string) (RadarrSettings, error) {
	for i, v := range Config.RADARR {
		if v.Name == name {
			return Config.RADARR[i], nil
		}
	}
	return RadarrSettings{}, errors.New("server not found")
}

// Get list of radarr servers without api keys.
// Regular users with access to adding to radarr will request this.
func getRadarrsSafe() []RadarrSettings {
	s := []RadarrSettings{}
	for _, v := range Config.RADARR {
		s = append(s, v.safe())
	}
	return s
}
