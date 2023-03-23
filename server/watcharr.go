package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

func main() {
	fmt.Println("Watcharr Starting")

	err := godotenv.Load()
	if err != nil {
		panic("Failed to load vars from .env file")
	}
	ensureEnv()

	sqldb, err := sql.Open(sqliteshim.ShimName, "./watcharr.db")
	if err != nil {
		panic(err)
	}

	db := bun.NewDB(sqldb, sqlitedialect.New())

	// Create tables if they don't exist
	db.NewCreateTable().Model((*User)(nil)).IfNotExists().Exec(context.TODO())
	db.NewCreateTable().Model((*Content)(nil)).IfNotExists().Exec(context.TODO())
	db.NewCreateTable().Model((*List)(nil)).IfNotExists().Exec(context.TODO())

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
	br.addContentRoutes()

	gin.Run("localhost:3080")
}

// Ensure all required environment variables are set.
func ensureEnv() {
	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("JWT_SECRET env var missing!")
	}
}
