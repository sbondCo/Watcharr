package main

type WatchedSeason struct {
	GormModel
	WatchedID    uint          `json:"watchedId" gorm:"not null"`
	SeasonNumber int           `json:"seasonNumber" gorm:"not null"`
	Status       WatchedStatus `json:"status"`
	Rating       int8          `json:"rating"`
}
