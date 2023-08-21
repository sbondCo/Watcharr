package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gopkg.in/natefinch/lumberjack.v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type GormModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
}

var (
	AvailableAuthProviders = []string{}
	TMDBKey                = "d047fa61d926371f277e7a83c9c4ff2c"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load vars from .env file:", err)
	}

	multiw := setupLogging()
	slog.Info("Watcharr Starting")

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

	err = db.AutoMigrate(&User{}, &Content{}, &Watched{}, &Activity{})
	if err != nil {
		log.Fatal("Failed to auto migrate database:", err)
	}

	if isProd {
		go runUI()
		gin.SetMode(gin.ReleaseMode)
	}
	gin.DefaultWriter = multiw
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
		// Proxy NoRoute requests to UI server
		gine.NoRoute(func(c *gin.Context) {
			director := func(req *http.Request) {
				req.URL.Scheme = "http"
				req.URL.Host = "127.0.0.1:3000"
			}
			proxy := &httputil.ReverseProxy{Director: director}
			proxy.ServeHTTP(c.Writer, c.Request)
		})
	}
	br := newBaseRouter(db, gine.Group("/api"))
	br.addAuthRoutes()
	br.addContentRoutes()
	br.addWatchedRoutes()
	br.addActivityRoutes()
	br.addProfileRoutes()
	br.addJellyfinRoutes()
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

	if os.Getenv("TMDB_KEY") != "" {
		slog.Info("Default TMDBKey being overriden by TMDB_KEY.")
		TMDBKey = os.Getenv("TMDB_KEY")
	}
}

// Setup slog defaults
func setupLogging() io.Writer {
	level := slog.LevelInfo
	if os.Getenv("DEBUG") == "true" {
		level = slog.LevelDebug
	}
	multiw := io.MultiWriter(&lumberjack.Logger{
		Filename:   "./data/watcharr.log",
		MaxSize:    1, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
		Compress:   false,
	}, os.Stdout)
	slog.SetDefault(slog.New(
		slog.NewTextHandler(multiw, &slog.HandlerOptions{Level: level}),
	))
	slog.Info("Logging level set", "logging_level", level)
	return multiw
}

// Run UI server
func runUI() {
	cmd := exec.Command("node", "ui/index.js")
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal("UI ERR ", err)
	}
	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	if err := cmd.Start(); err != nil {
		log.Fatal("UI ERR ", err)
	}
	if err := cmd.Wait(); err != nil {
		log.Fatal("UI ERR ", err)
	}
}

func ensureDirExists(dir string) error {
	err := os.MkdirAll(dir, 0764)
	if err != nil {
		return err
	}
	return nil
}
