package main

import (
	"errors"
	"log/slog"

	"github.com/sbondCo/Watcharr/arr"
)

type SonarrSettings struct {
	// ID              int    `json:"id,omitempty"`
	Name            string `json:"name,omitempty"`
	Host            string `json:"host,omitempty"`
	Key             string `json:"key,omitempty"`
	QualityProfile  int    `json:"qualityProfile,omitempty"`
	RootFolder      int    `json:"rootFolder,omitempty"`
	LanguageProfile int    `json:"languageProfile,omitempty"`
	AutomaticSearch bool   `json:"automaticSearch"`
	// TODO eventually separate profiles and root for anime content (i can see diff language profile being useful)
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
	edited := false
	for i, v := range Config.SONARR {
		if v.Name == s.Name {
			Config.SONARR[i] = s
			edited = true
		}
	}
	if !edited {
		return errors.New("can't edit server that does not exist")
	}
	Config.SONARR = append(Config.SONARR, s)
	writeConfig()
	return nil
}
