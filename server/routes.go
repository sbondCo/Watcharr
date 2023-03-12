package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

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

func (b *BaseRouter) addAuthRoutes() {
	content := b.rg.Group("/auth")

	// Login
	content.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, getContent(b.db))
	})

	// Register
	content.POST("/", func(c *gin.Context) {
		var user User
		if c.ShouldBindJSON(&user) == nil {
			println(user.Username)
			println(user.Password)
			response, err := register(&user, b.db)
			if err != nil {
				c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
				return
			}
			c.JSON(http.StatusOK, response)
			return
		}
		c.Status(400)
	})
}
