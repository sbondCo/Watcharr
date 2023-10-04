package main

import (
	"log/slog"
	"time"

	"gorm.io/gorm"
)

// Setup recurring tasks (eg cleanup every 5 mins)
func setupTasks(db *gorm.DB) {
	taskRunInterval := 5 * time.Second
	ticker := time.NewTicker(taskRunInterval)
	defer ticker.Stop()

	for range ticker.C {
		slog.Info("TIMER GOING OFFFFFFFF")
		// Runs funcs that are in the place where we are cleaning.
		// Bit cleaner and we can keep the related code close to its home.
		cleanupTokens(db)
	}
}
