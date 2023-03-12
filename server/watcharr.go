package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

func main() {
	fmt.Println("Watcharr Starting")

	sqldb, err := sql.Open(sqliteshim.ShimName, "./watcharr.db")
	if err != nil {
		panic(err)
	}

	db := bun.NewDB(sqldb, sqlitedialect.New())

	// Create tables if they don't exist
	db.NewCreateTable().Model((*User)(nil)).IfNotExists().Exec(context.TODO())
	db.NewCreateTable().Model((*List)(nil)).IfNotExists().Exec(context.TODO())

	gin := gin.Default()
	br := newBaseRouter(db, gin)
	br.addAuthRoutes()
	br.addContentRoutes()

	gin.Run("localhost:3080")
}
