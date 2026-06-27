package entity

import "github.com/sbondCo/Watcharr/database/dbmodel"

// UniqueIndex applied between WatchedID and SeasonNumber to avoid duplicates incase logic fails.
type WatchedSeason struct {
	dbmodel.GormModelNoDel
	UserID       uint          `json:"-" gorm:"not null"`
	User         User          `json:"-"`
	WatchedID    uint          `json:"-" gorm:"uniqueIndex:ws_watched_to_season_num;not null"`
	SeasonNumber int           `json:"seasonNumber" gorm:"uniqueIndex:ws_watched_to_season_num;not null"`
	Status       WatchedStatus `json:"status"`
	Rating       int8          `json:"rating"`
}
