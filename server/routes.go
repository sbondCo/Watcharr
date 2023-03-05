package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func addContentRoutes(rg *gin.Engine) {
	content := rg.Group("/content")

	content.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, getContent())
	})
}
