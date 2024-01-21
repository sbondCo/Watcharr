package game

import (
	"encoding/json"
	"time"
)

type TwitchTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// So we can unmarshall the unix timestamps returned from igdb into time.Time.
type UnixTime struct {
	time.Time
}

func (u *UnixTime) UnmarshalJSON(b []byte) error {
	var timestamp int64
	err := json.Unmarshal(b, &timestamp)
	if err != nil {
		return err
	}
	u.Time = time.Unix(timestamp, 0)
	return nil
}

// Only the fields we request
type GameSearchResponse []struct {
	ID    int `json:"id"`
	Cover struct {
		ID      int    `json:"id"`
		ImageID string `json:"image_id"`
	} `json:"cover"`
	FirstReleaseDate UnixTime `json:"first_release_date"`
	Name             string   `json:"name"`
	Summary          string   `json:"summary,omitempty"`
	VersionTitle     string   `json:"version_title,omitempty"`
}
