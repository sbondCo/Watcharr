package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type BaseRouter struct {
	db *bun.DB
	rg *gin.Engine
}

func newBaseRouter(db *bun.DB, rg *gin.Engine) *BaseRouter {
	return &BaseRouter{
		db: db,
		rg: rg,
	}
}

func (b *BaseRouter) addContentRoutes() {
	content := b.rg.Group("/content")

	content.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, getContent(b.db))
	})
}
