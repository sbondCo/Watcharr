package main

import (
	"errors"
	"log/slog"

	"github.com/sbondCo/Watcharr/arr"
)

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
