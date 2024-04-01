package main

import (
	"time"

	"gorm.io/gorm"
)

// Setup recurring tasks (eg cleanup every x mins)
func setupTasks(db *gorm.DB) {
	go setupMinutelyTasks(db)
	go setupDailyTasks(db)
}

func setupMinutelyTasks(db *gorm.DB) {
	taskRunInterval := 1 * time.Minute
	ticker := time.NewTicker(taskRunInterval)
	defer ticker.Stop()

	for range ticker.C {
		// Runs funcs that are in the place where we are cleaning.
		// Bit cleaner and we can keep the related code close to its home.
		cleanupTokens(db)
		refreshArrQueues()
	}
}

func setupDailyTasks(db *gorm.DB) {
	taskRunInterval := 24 * time.Hour
	ticker := time.NewTicker(taskRunInterval)
	defer ticker.Stop()

	for range ticker.C {
		cleanupImages(db)
	}
}
