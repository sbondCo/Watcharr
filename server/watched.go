package main

import (
	"gorm.io/gorm"
)

type Watched struct {
	gorm.Model

	ID        int     `json:"id"`
	Finished  bool    `json:"watched"`
	UserID    uint    `json:"-"`
	ContentID int     `json:"-"`
	Content   Content `json:"content"`
}

func getWatched(db *gorm.DB) Watched {
	watched := new(Watched)
	res := db.Where("user_id = ?", 1).Find(&watched)
	if res.Error != nil {
		panic(res.Error)
	}
	return *watched
}
