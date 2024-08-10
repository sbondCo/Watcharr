package main

import (
	"errors"
	"log/slog"

	"gorm.io/gorm"
)

// I think tags will be private for the user.
// If the user wants to make a public list, they should make a custom view.

type Tag struct {
	GormModel
	// ID of user that own this tag.
	UserID uint `json:"-" gorm:"not null"`
	// Name of the tag.
	Name string `json:"name" gorm:"not null"`
	// Hex of text color.
	Color string `json:"color"`
	// Hex of background color.
	BgColor string `json:"bgColor"`
}

type TagAddRequest struct {
	Name    string `json:"name" binding:"required"`
	Color   string `json:"color"`
	BgColor string `json:"bgColor"`
}

func getTags(db *gorm.DB, userId uint) ([]Tag, error) {
	tags := new([]Tag)
	res := db.Model(&Tag{}).Where("user_id = ?", userId).Find(&tags)
	if res.Error != nil {
		slog.Error("Failed getting tags from database", "error", res.Error.Error())
		return []Tag{}, errors.New("failed getting tags")
	}
	return *tags, nil
}

// Let user create a tag.
func addTag(db *gorm.DB, userId uint, tr TagAddRequest) (Tag, error) {
	if tr.Name == "" {
		return Tag{}, errors.New("tag must have a name")
	}
	tag := Tag{UserID: userId, Name: tr.Name, Color: tr.Color, BgColor: tr.BgColor}
	res := db.Create(&tag)
	if res.Error != nil {
		slog.Error("Error adding tag to database", "error", res.Error.Error())
		return Tag{}, errors.New("failed adding new tag to database")
	}
	slog.Debug("Adding tag", "added_tag", tag)
	return tag, nil
}
