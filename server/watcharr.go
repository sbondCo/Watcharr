package main

import (
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

	gin := gin.Default()
	br := newBaseRouter(db, gin)
	br.addContentRoutes()

	gin.Run("localhost:3080")
}
