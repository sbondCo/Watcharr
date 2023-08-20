package main

import (
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type BaseRouter struct {
	db *gorm.DB
	rg *gin.RouterGroup
}

func newBaseRouter(db *gorm.DB, rg *gin.RouterGroup) *BaseRouter {
	return &BaseRouter{
		db: db,
		rg: rg,
	}
}

func (b *BaseRouter) addContentRoutes() {
	content := b.rg.Group("/content").Use(AuthRequired(nil))

	// Get trending content
	// content.GET("/", func(c *gin.Context) {
	// 	c.JSON(http.StatusOK, getWatched(b.db))
	// })

	// Search for content
	content.GET("/:query", func(c *gin.Context) {
		println(c.Param("query"))
		if c.Param("query") == "" {
			c.Status(400)
			return
		}
		content, err := searchContent(c.Param("query"))
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, content)
	})

	// Get movie details (for movie page)
	content.Use(WhereaboutsRequired()).GET("/movie/:id", func(c *gin.Context) {
		if c.Param("id") == "" {
			c.Status(400)
			return
		}
		content, err := movieDetails(c.Param("id"), c.MustGet("userCountry").(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, content)
	})

	// Get movie cast
	content.GET("/movie/:id/credits", func(c *gin.Context) {
		if c.Param("id") == "" {
			c.Status(400)
			return
		}
		content, err := movieCredits(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, content)
	})

	// Get tv details (for tv page)
	content.Use(WhereaboutsRequired()).GET("/tv/:id", func(c *gin.Context) {
		if c.Param("id") == "" {
			c.Status(400)
			return
		}
		content, err := tvDetails(c.Param("id"), c.MustGet("userCountry").(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, content)
	})

	// Get tv cast
	content.GET("/tv/:id/credits", func(c *gin.Context) {
		if c.Param("id") == "" {
			c.Status(400)
			return
		}
		content, err := tvCredits(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, content)
	})

	// Get season details
	content.GET("/tv/:id/season/:num", func(c *gin.Context) {
		if c.Param("id") == "" || c.Param("num") == "" {
			c.Status(400)
			return
		}
		content, err := seasonDetails(c.Param("id"), c.Param("num"))
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, content)
	})

	// Get person details
	content.GET("/person/:id", func(c *gin.Context) {
		if c.Param("id") == "" {
			c.Status(400)
			return
		}
		content, err := personDetails(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, content)
	})

	// Get person credits
	content.GET("/person/:id/credits", func(c *gin.Context) {
		if c.Param("id") == "" {
			c.Status(400)
			return
		}
		content, err := personCredits(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, content)
	})

	// Discover movies
	content.GET("/discover/movies", func(c *gin.Context) {
		content, err := discoverMovies()
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, content)
	})

	// Discover shows
	content.GET("/discover/tv", func(c *gin.Context) {
		content, err := discoverTv()
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, content)
	})

	// Get all trending (movies, tv, people)
	content.GET("/trending", func(c *gin.Context) {
		content, err := allTrending()
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, content)
	})

	// Upcoming Movies
	content.GET("/upcoming/movies", func(c *gin.Context) {
		content, err := upcomingMovies()
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, content)
	})

	// Upcoming Tv
	content.GET("/upcoming/tv", func(c *gin.Context) {
		content, err := upcomingTv()
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, content)
	})
}

func (b *BaseRouter) addWatchedRoutes() {
	watched := b.rg.Group("/watched").Use(AuthRequired(nil))

	watched.GET("", func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		c.JSON(http.StatusOK, getWatched(b.db, userId))
	})

	watched.POST("", func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		var ar WatchedAddRequest
		err := c.ShouldBindJSON(&ar)
		if err == nil {
			response, err := addWatched(b.db, userId, ar)
			if err != nil {
				c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
				return
			}
			c.JSON(http.StatusOK, response)
			return
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	})

	watched.PUT(":id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.Status(400)
			return
		}
		userId := c.MustGet("userId").(uint)
		var ur WatchedUpdateRequest
		err = c.ShouldBindJSON(&ur)
		if err == nil {
			response, err := updateWatched(b.db, userId, uint(id), ur)
			if err != nil {
				c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
				return
			}
			c.JSON(http.StatusOK, response)
			return
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	})

	watched.DELETE(":id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.Status(400)
			return
		}
		userId := c.MustGet("userId").(uint)
		if err == nil {
			response, err := removeWatched(b.db, userId, uint(id))
			if err != nil {
				c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
				return
			}
			c.JSON(http.StatusOK, response)
			return
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	})
}

func (b *BaseRouter) addActivityRoutes() {
	activity := b.rg.Group("/activity").Use(AuthRequired(nil))

	activity.GET(":watchedId", func(c *gin.Context) {
		watchedId, err := strconv.ParseUint(c.Param("watchedId"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "check watched id route param"})
			return
		}
		userId := c.MustGet("userId").(uint)
		activity, err := getActivity(b.db, userId, uint(watchedId))
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, activity)
	})

	activity.POST("", func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		var ar ActivityAddRequest
		err := c.ShouldBindJSON(&ar)
		if err == nil {
			response, err := addActivity(b.db, userId, ar)
			if err != nil {
				c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
				return
			}
			c.JSON(http.StatusOK, response)
			return
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	})
}

func (b *BaseRouter) addAuthRoutes() {
	auth := b.rg.Group("/auth")

	// Login
	auth.POST("/", func(c *gin.Context) {
		var user User
		if c.ShouldBindJSON(&user) == nil {
			response, err := login(&user, b.db)
			if err != nil {
				c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
				return
			}
			c.JSON(http.StatusOK, response)
			return
		}
		c.Status(400)
	})

	// Jellyfin login
	auth.POST("/jellyfin", func(c *gin.Context) {
		var user User
		if c.ShouldBindJSON(&user) == nil {
			response, err := loginJellyfin(&user, b.db)
			if err != nil {
				c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
				return
			}
			c.JSON(http.StatusOK, response)
			return
		}
		c.Status(400)
	})

	// Register
	auth.POST("/register", func(c *gin.Context) {
		var user User
		if c.ShouldBindJSON(&user) == nil {
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

	// Get available auth providers
	auth.GET("/available", func(c *gin.Context) {
		signupEnabled := true
		if os.Getenv("SIGNUP_ENABLED") == "false" {
			signupEnabled = false
		}
		c.JSON(http.StatusOK, &AvailableAuthProvidersResponse{
			AvailableAuthProviders: AvailableAuthProviders,
			SignupEnabled:          signupEnabled,
		})
	})
}

func (b *BaseRouter) addProfileRoutes() {
	profile := b.rg.Group("/profile").Use(AuthRequired(nil))

	// Get user profile details
	profile.GET("", func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		response, err := getProfile(b.db, userId)
		if err != nil {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	})
}

func (b *BaseRouter) addJellyfinRoutes() {
	jf := b.rg.Group("/jellyfin").Use(AuthRequired(b.db))

	// Check if jf has item
	jf.GET("/:type/:name/:tmdbId", func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		userType := c.MustGet("userType").(UserType)
		username := c.MustGet("username").(string)
		userThirdPartyId := c.MustGet("userThirdPartyId").(string)
		userThirdPartyAuth := c.MustGet("userThirdPartyAuth").(string)
		response, err := jellyfinContentFind(userId, userType, username, userThirdPartyId, userThirdPartyAuth, c.Param("type"), c.Param("name"), c.Param("tmdbId"))
		if err != nil {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	})
}
