package main

import (
	"fmt"
	"log"
	"net/http"
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

var AvailableAuthProviders = []string{}

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

	// Check if we want to be in DEV or PROD
	isProd := true
	if os.Getenv("MODE") == "DEV" {
		isProd = false
	}

	db, err := gorm.Open(sqlite.Open("./data/watcharr.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}

	err = db.AutoMigrate(&User{}, &Content{}, &Watched{})
	if err != nil {
		log.Fatal("Failed to auto migrate database:", err)
	}

	if isProd {
		gin.SetMode(gin.ReleaseMode)
	}
	gine := gin.Default()
	gine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "accept", "origin", "Cache-Control", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	if isProd {
		gine.NoRoute(gin.WrapH(http.FileServer(gin.Dir("/ui", false))))
	}
	br := newBaseRouter(db, gine.Group("/api"))
	br.addAuthRoutes()
	br.addContentRoutes()
	br.addWatchedRoutes()
	br.rg.Static("/img", "./data/img")

	gine.Run("0.0.0.0:3080")
}

// Ensure all required environment variables are set.
func ensureEnv() {
	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("JWT_SECRET env var missing!")
	}

	if os.Getenv("JELLYFIN_HOST") != "" {
		AvailableAuthProviders = append(AvailableAuthProviders, "jellyfin")
	}
}

func ensureDirExists(dir string) error {
	err := os.MkdirAll(dir, 0764)
	if err != nil {
		return err
	}
	return nil
}
