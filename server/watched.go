package main

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type Watched struct {
	GormModel
	Finished  bool    `json:"watched"`
	UserID    uint    `json:"-"`
	ContentID int     `json:"-"`
	Content   Content `json:"content"`
}

func getWatched(db *gorm.DB, userId uint) []Watched {
	watched := new([]Watched)
	res := db.Model(&Watched{}).Preload("Content").Where("user_id = ?", userId).Find(&watched)
	if res.Error != nil {
		panic(res.Error)
	}
	return *watched
}

func addWatched(db *gorm.DB, userId uint, content Content) (bool, error) {
	if content.ID == 0 {
		return false, errors.New("content has no ID")
	}

	// Save the content in our db
	res := db.Create(&content)
	if res.Error != nil {
		// Error if anything but unique contraint error
		if !strings.Contains(res.Error.Error(), "UNIQUE") {
			println("Error creating content in database:", res.Error.Error())
			return false, errors.New("failed to cache content in database")
		}
	}
	println(res.RowsAffected)

	watched := Watched{Finished: true, UserID: userId, ContentID: content.ID}
	res = db.Create(&watched)
	if res.Error != nil {
		println("Error adding watched content to database:", res.Error.Error())
		return false, errors.New("failed adding content to database")
	}
	println(res.RowsAffected)
	fmt.Printf("%+v\n", watched)

	return true, nil
}
