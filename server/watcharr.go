package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("Watcharr Starting")

	err := godotenv.Load()
	if err != nil {
		panic("Failed to load vars from .env file")
	}
	ensureEnv()

	db, err := gorm.Open(sqlite.Open("./watcharr.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}

	db.AutoMigrate(&User{})

	gin := gin.Default()
	gin.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "accept", "origin", "Cache-Control", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	br := newBaseRouter(db, gin)
	br.addAuthRoutes()
	// br.addContentRoutes()

	gin.Run("localhost:3080")
}

// Ensure all required environment variables are set.
func ensureEnv() {
	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("JWT_SECRET env var missing!")
	}
}
