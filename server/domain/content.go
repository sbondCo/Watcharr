package domain

import "time"

type ContentDetailsResponse struct {
	Media
}

type PersonCreditsResponse struct {
	// Acting credits.
	Credits []Media `json:"credits,omitempty"`
	// Credits behind the camera by TMDB department ("Directing", "Writing"…),
	// one entry per title. Without these a director's page was near empty.
	Crew map[string][]Media `json:"crew,omitempty"`
}

type PersonDetailsResponse struct {
	Name               string    `json:"name,omitempty"`
	Birthday           time.Time `json:"birthday,omitzero"`
	Deathday           time.Time `json:"deathday,omitzero"`
	Age                int       `json:"age,omitempty"`
	PlaceOfBirth       string    `json:"placeOfBirth,omitempty"`
	KnownForDepartment string    `json:"knownForDepartment,omitempty"`
	Biography          string    `json:"biography,omitempty"`
	ExtPosterPath      string    `json:"extPosterPath,omitempty"`
	Homepage           string    `json:"homepage,omitempty"`
}
