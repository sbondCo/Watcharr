package domain

import "time"

type ContentDetailsResponse struct {
	Media
}

type PersonCreditsResponse struct {
	Credits []Media `json:"credits,omitempty"`
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
	KnownCredits       int       `json:"knownCredits,omitempty"`
	// Earliest found credit.
	FirstKnownCredit PersonDetailsFirstKnownCredit `json:"firstKnownCredit,omitzero"`
	// NOTE: This is an untyped string slice because we don't really need
	// it to be typed right now, we don't do any specific processing that would
	// require it or be helpful.
	// If we want to add Departments string type, found this helpful forum post:
	// https://www.themoviedb.org/talk/598c3a70925141080100e601
	// We currently add "Acting" if the person has any 'cast' credits to
	// indicate to frontend that filtering for only cast credits is possible.
	CreditTypes []string `json:"creditTypes,omitempty"`
}

type PersonDetailsFirstKnownCredit struct {
	ID   int       `json:"id,omitempty"`
	Type string    `json:"type,omitempty"`
	Date time.Time `json:"date,omitzero"`
	// Age of person when credited.
	Age int `json:"age,omitzero"`
}

const CreditTypeActing = "Acting"
