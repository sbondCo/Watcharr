package entity

import "github.com/sbondCo/Watcharr/database/dbmodel"

// UniqueIndex applied between WatchedID, SeasonNum and EpisodeNum to avoid duplicates incase logic fails.
//
// Episodes on tmdb are only queried by season number + episode number, not possible via episode id,
// since episodes can be removed and re-added. For this reason we store season and episodes nums instead
// of just the episode id.
type WatchedEpisode struct {
	dbmodel.GormModelNoDel
	UserID        uint          `json:"-" gorm:"not null"`
	User          User          `json:"-"`
	WatchedID     uint          `json:"-" gorm:"uniqueIndex:we_watched_to_ens;not null"`
	SeasonNumber  int           `json:"seasonNumber" gorm:"uniqueIndex:we_watched_to_ens;not null"`
	EpisodeNumber int           `json:"episodeNumber" gorm:"uniqueIndex:we_watched_to_ens;not null"`
	Status        WatchedStatus `json:"status"`
	Rating        int8          `json:"rating"`
}
