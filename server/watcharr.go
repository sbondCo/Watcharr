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

var ServerInSetup = false

func main() {
	err := godotenv.Load()
	if err != nil {
		// Do not fail if file does not exist
		if !os.IsNotExist(err) {
			log.Fatal("Failed to load vars from .env file:", err)
		}
	}

	multiw, logLevel := setupLogging()
	slog.Info("Watcharr Starting")

	if err = readConfig(); err != nil {
		log.Fatal("Failed to read server config!", err)
	}

	if Config.DEBUG {
		logLevel.Set(slog.LevelDebug)
	}
	slog.Info("Logging level set", "logging_level", logLevel)

	// Ensure data dir exists
	err = ensureDirExists("./data")
	if err != nil {
		log.Fatal("Failed to create data dir:", err)
	}

	// Check if we want to be in DEV or PROD
	isProd := true
	if os.Getenv("MODE") == "DEV" {
		slog.Info("Starting in DEV mode")
		isProd = false
	}

	db, err := gorm.Open(sqlite.Open("./data/watcharr.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	err = db.AutoMigrate(&User{}, &Content{}, &Watched{}, &Activity{}, &Token{})
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
	// Only add setup routes if there are no users found in db.
	var userCount int64
	if uresp := db.Model(&User{}).Count(&userCount); uresp.Error == nil {
		if userCount != 0 {
			slog.Debug("registered users found.. skipped creating setup routes.")
		} else {
			slog.Info("No users found.. creating setup routes.")
			ServerInSetup = true
			br.addSetupRoutes()
		}
	} else {
		slog.Error("Failed to check if any users exist.. not registering setup routes", "error", uresp.Error)
	}
	br.addAuthRoutes()
	br.addContentRoutes()
	br.addWatchedRoutes()
	br.addActivityRoutes()
	br.addProfileRoutes()
	br.addJellyfinRoutes()
	br.addUserRoutes()
	br.addImportRoutes()
	br.rg.Static("/img", "./data/img")

	go setupTasks(db)

	gine.Run("0.0.0.0:3080")
}

// Setup slog defaults
func setupLogging() (io.Writer, *slog.LevelVar) {
	logLevel := new(slog.LevelVar)
	multiw := io.MultiWriter(&lumberjack.Logger{
		Filename:   "./data/watcharr.log",
		MaxSize:    1, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
		Compress:   false,
	}, os.Stdout)
	slog.SetDefault(slog.New(
		slog.NewTextHandler(multiw, &slog.HandlerOptions{Level: logLevel}),
	))
	return multiw, logLevel
}

// Run UI server
func runUI() {
	cmd := exec.Command("node", "ui/index.js")
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal("UI ERR @ get stdout read pipe: ", err)
	}
	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	if err := cmd.Start(); err != nil {
		log.Fatal("UI ERR @ command start: ", err)
	}
	if err := cmd.Wait(); err != nil {
		log.Fatal("UI ERR @ command wait: ", err)
	}
}

func ensureDirExists(dir string) error {
	err := os.MkdirAll(dir, 0764)
	if err != nil {
		return err
	}
	return nil
}
