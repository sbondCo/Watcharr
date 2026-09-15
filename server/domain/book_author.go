package domain

import (
	"time"

	"github.com/sbondCo/Watcharr/util"
)

// Data class for the overview page about an author.
type Author struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Biography string `json:"biography"`

	// Optional fields
	Homepage  *string    `json:"homepage"`
	Photo     *string    `json:"photo"`
	BirthDate *time.Time `json:"birthDate"`
	DeathDate *time.Time `json:"deathDate"`
}

func (a *Author) AsPersonDetailsResponse() PersonDetailsResponse {
	m := PersonDetailsResponse{
		Name:      a.Name,
		Biography: a.Biography,
	}
	if a.Homepage != nil {
		m.Homepage = *a.Homepage
	}
	if a.Photo != nil {
		m.ExtPosterPath = *a.Photo
	}
	if a.BirthDate != nil {
		m.Birthday = *a.BirthDate
	}
	if a.DeathDate != nil {
		m.Deathday = *a.DeathDate
	}
	m.Age = util.GetAge(m.Birthday, m.Deathday)
	return m
}
