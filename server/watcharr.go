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

type GormModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
}

func main() {
	fmt.Println("Watcharr Starting")

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load vars from .env file:", err)
	}
	ensureEnv()

	// Ensure data dir exists
	err = ensureDirExists("./data")
	if err != nil {
		log.Fatal("Failed to create data dir:", err)
	}

	db, err := gorm.Open(sqlite.Open("./data/watcharr.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}

	db.AutoMigrate(&User{}, &Content{}, &Watched{})

	gin := gin.Default()
	gin.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "accept", "origin", "Cache-Control", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	br := newBaseRouter(db, gin)
	br.addAuthRoutes()
	br.addContentRoutes()
	br.addWatchedRoutes()
	br.rg.Static("/img", "./data/img")

	gin.Run("0.0.0.0:3080")
}

// Ensure all required environment variables are set.
func ensureEnv() {
	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("JWT_SECRET env var missing!")
	}
}

func ensureDirExists(dir string) error {
	err := os.MkdirAll(dir, 0764)
	if err != nil {
		return err
	}
	return nil
}
