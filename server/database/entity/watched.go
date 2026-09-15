package entity

import "github.com/sbondCo/Watcharr/database/dbmodel"

type WatchedStatus string

const (
	FINISHED WatchedStatus = "FINISHED"
	WATCHING WatchedStatus = "WATCHING"
	PLANNED  WatchedStatus = "PLANNED"
	HOLD     WatchedStatus = "HOLD"
	DROPPED  WatchedStatus = "DROPPED"
)

func (r WatchedStatus) IsValid() bool {
	switch r {
	case FINISHED,
		WATCHING,
		PLANNED,
		HOLD,
		DROPPED:
		return true
	}
	return false
}

type Watched struct {
	dbmodel.GormModel
	Status WatchedStatus `json:"status"`
	// float so we can support decimal ratings.
	// Ratings should still always be saved as out of 10.0,
	// so they can be viewed with any ratings setting in the client.
	Rating          float64          `json:"rating" gorm:"type:numeric(2,1)"`
	Thoughts        string           `json:"thoughts"`
	Pinned          bool             `json:"pinned" gorm:"default:false;not null"`
	UserID          uint             `json:"-" gorm:"uniqueIndex:usernctnidx;uniqueIndex:userngamidx;uniqueIndex:usernbookidx"`
	ContentID       *int             `json:"-" gorm:"uniqueIndex:usernctnidx"`
	Content         *Content         `json:"content,omitempty"`
	GameID          *int             `json:"-" gorm:"uniqueIndex:userngamidx"`
	Game            *Game            `json:"game,omitempty"`
	BookID          *int             `json:"-" gorm:"uniqueIndex:usernbookidx"`
	Book            *Book            `json:"book,omitempty"`
	Activity        []Activity       `json:"activity"`
	WatchedSeasons  []WatchedSeason  `json:"watchedSeasons,omitempty"`  // For shows
	WatchedEpisodes []WatchedEpisode `json:"watchedEpisodes,omitempty"` // For shows
	Tags            []Tag            `json:"tags,omitempty" gorm:"many2many:watched_tags;"`
	// The last season that was viewed by the user for this watched entry.
	// Only applies to tv shows of course.
	LastViewedSeason *int `json:"lastViewedSeason,omitempty"`
}
