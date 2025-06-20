package main

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
)

// addAPIKeyRoutes registers /api_keys CRUD operations. Requires normal JWT auth.
func (b *BaseRouter) addAPIKeyRoutes() {
    api := b.rg.Group("/api_keys").Use(AuthRequired(b.db))

    // List keys
    api.GET("", func(c *gin.Context) {
        userId := c.MustGet("userId").(uint)
        keys, err := getAPIKeys(b.db, userId)
        if err != nil {
            c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
            return
        }
        c.JSON(http.StatusOK, keys)
    })

    // Create new key (returns plaintext once)
    api.POST("", func(c *gin.Context) {
        userId := c.MustGet("userId").(uint)
        key, err := createAPIKey(b.db, userId)
        if err != nil {
            c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{"apiKey": key})
    })

    // Revoke (delete) key
    api.DELETE(":id", func(c *gin.Context) {
        userId := c.MustGet("userId").(uint)
        id, err := strconv.Atoi(c.Param("id"))
        if err != nil {
            c.Status(http.StatusBadRequest)
            return
        }
        if err := revokeAPIKey(b.db, userId, uint(id)); err != nil {
            c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
            return
        }
        c.Status(http.StatusOK)
    })
}
