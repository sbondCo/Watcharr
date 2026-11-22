package entity

import "time"

// NOTE: When making a new reference to an Image, ensure the CleanupImages task
// still works correctly!
type Image struct {
	ID        uint      `gorm:"primarykey" json:"-"`
	CreatedAt time.Time `json:"createdAt"`
	Hash      string    `gorm:"uniqueIndex;not null" json:"-"`
	BlurHash  string    `json:"blurHash"`
	// Path constructable from hash alone, but I can't decide
	// if I should have this or not so I figure it's easier
	// to remove it later than to add it later....... -_-
	Path string `gorm:"not null" json:"path"`
}
