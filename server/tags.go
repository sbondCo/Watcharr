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

// Add watched content to a tag (user must own the tag and watched entry).
func addWatchedToTag(db *gorm.DB, userId uint, tagId uint, watchedId uint) error {
	slog.Debug("addWatchedToTag: Adding", "userId", userId, "watchedID", watchedId, "tagId", tagId)
	// 1. Make sure watched item exists and is owned by this user
	var w Watched
	if resp := db.Where("id = ? AND user_id = ?", watchedId, userId).Preload("Tags").Find(&w); resp.Error != nil {
		slog.Error("addWatchedToTag: failed to get watched item from db", "error", resp.Error)
		return errors.New("failed when retrieving watched item")
	}
	if w.ID == 0 {
		slog.Error("addWatchedToTag", "error", "watched item does not exist in db", "watchedID", watchedId)
		return errors.New("watched entry does not exist")
	}
	// 2. Make sure tag exists
	var t Tag
	if resp := db.Where("id = ? AND user_id = ?", tagId, userId).Find(&t); resp.Error != nil {
		slog.Error("addWatchedToTag: Failed to get tag from db", "error", resp.Error)
		return errors.New("failed when retrieving tag")
	}
	if t.ID == 0 {
		slog.Error("addWatchedToTag", "error", "tag does not exist in db", "tagId", tagId)
		return errors.New("tag does not exist")
	}
	// 3. Save relation (unique restraint will fail if it already exists)
	w.Tags = append(w.Tags, t)
	resp := db.Save(&w)
	if resp.Error != nil {
		slog.Error("addWatchedToTag: Failed to tag watched item", "error", resp.Error)
		return errors.New("failed to tag watched item")
	}
	slog.Debug("addWatchedToTag: watched content successfully linked to tag", "watchedID", watchedId, "tagId", tagId)
	return nil
}
