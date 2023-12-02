package main

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/cache"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/arr"
	"gorm.io/gorm"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type KeyValueRequest struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

type BaseRouter struct {
	db *gorm.DB
	rg *gin.RouterGroup
	ms *persistence.InMemoryStore
}

func newBaseRouter(db *gorm.DB, rg *gin.RouterGroup) *BaseRouter {
	return &BaseRouter{
		db: db,
		rg: rg,
		ms: persistence.NewInMemoryStore(time.Hour * 24),
	}
}

// Since we cannot remove these setup routes after they are registered,
// each route/service should ensure we are still in setup before continuing.
// After server restart, these routes shouldn't exist if setup finished
// (currently it is finished if a user is created).
//
// Each controller can check ServerInSetup var first, then each service
// can double check what it needs to (eg create_admin service, registerFirstUser,
// will check that no users exist).
func (b *BaseRouter) addSetupRoutes() {
	setup := b.rg.Group("/setup")

	setup.POST("/create_admin", func(c *gin.Context) {
		if !ServerInSetup {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "not in setup"})
			return
		}
		var user UserRegisterRequest
		if c.ShouldBindJSON(&user) == nil {
			response, err := registerFirstUser(&user, b.db)
			if err != nil {
				c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
				return
			} else {
				// Set in setup to false after first user registered successfully
				ServerInSetup = false
			}
			c.JSON(http.StatusOK, response)
			return
		}
		c.Status(400)
	})
}

func (b *BaseRouter) addContentRoutes() {
	content := b.rg.Group("/content").Use(AuthRequired(nil))
	exp := time.Hour * 24

	// Search for content
	content.GET("/:query", cache.CachePage(b.ms, exp, func(c *gin.Context) {
		// println(c.Param("query"))
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
	}))

	// Get movie details (for movie page)
	content.Use(WhereaboutsRequired()).GET("/movie/:id", cache.CachePage(b.ms, exp, func(c *gin.Context) {
		if c.Param("id") == "" {
			c.Status(400)
			return
		}
		content, err := movieDetails(c.Param("id"), c.MustGet("userCountry").(string), map[string]string{"append_to_response": "videos,watch/providers,similar"})
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, content)
	}))

	// Get movie cast
	content.GET("/movie/:id/credits", cache.CachePage(b.ms, exp, func(c *gin.Context) {
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
	}))

	// Get tv details (for tv page)
	content.Use(WhereaboutsRequired()).GET("/tv/:id", cache.CachePage(b.ms, exp, func(c *gin.Context) {
		if c.Param("id") == "" {
			c.Status(400)
			return
		}
		content, err := tvDetails(c.Param("id"), c.MustGet("userCountry").(string), map[string]string{"append_to_response": "videos,watch/providers,similar,external_ids,keywords"})
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, content)
	}))

	// Get tv cast
	content.GET("/tv/:id/credits", cache.CachePage(b.ms, exp, func(c *gin.Context) {
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
	}))

	// Get season details
	content.GET("/tv/:id/season/:num", cache.CachePage(b.ms, exp, func(c *gin.Context) {
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
	}))

	// Get person details
	content.GET("/person/:id", cache.CachePage(b.ms, exp, func(c *gin.Context) {
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
	}))

	// Get person credits
	content.GET("/person/:id/credits", cache.CachePage(b.ms, exp, func(c *gin.Context) {
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
	}))

	// Discover movies
	content.GET("/discover/movies", cache.CachePage(b.ms, exp, func(c *gin.Context) {
		content, err := discoverMovies()
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, content)
	}))

	// Discover shows
	content.GET("/discover/tv", cache.CachePage(b.ms, exp, func(c *gin.Context) {
		content, err := discoverTv()
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, content)
	}))

	// Get all trending (movies, tv, people)
	content.GET("/trending", cache.CachePage(b.ms, exp, func(c *gin.Context) {
		content, err := allTrending()
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, content)
	}))

	// Upcoming Movies
	content.GET("/upcoming/movies", cache.CachePage(b.ms, exp, func(c *gin.Context) {
		content, err := upcomingMovies()
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, content)
	}))

	// Upcoming Tv
	content.GET("/upcoming/tv", cache.CachePage(b.ms, exp, func(c *gin.Context) {
		content, err := upcomingTv()
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, content)
	}))
}

func (b *BaseRouter) addWatchedRoutes() {
	watched := b.rg.Group("/watched").Use(AuthRequired(nil))

	watched.GET("", func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		c.JSON(http.StatusOK, getWatched(b.db, userId))
	})

	watched.GET(":id/:username", func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			slog.Error("getPublicWatched route failed to convert id param to uint", "id", id)
			c.Status(400)
			return
		}
		response, err := getPublicWatched(b.db, uint(id), c.Param("username"))
		if err != nil {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	})

	watched.POST("", func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		var ar WatchedAddRequest
		err := c.ShouldBindJSON(&ar)
		if err == nil {
			response, err := addWatched(b.db, userId, ar, ADDED_WATCHED)
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

	watched.POST("/season", func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		var ar WatchedSeasonAddRequest
		err := c.ShouldBindJSON(&ar)
		if err == nil {
			response, err := addWatchedSeason(b.db, userId, ar, ADDED_WATCHED)
			if err != nil {
				c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
				return
			}
			c.JSON(http.StatusOK, response)
			return
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	})

	watched.DELETE("/season/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.Status(400)
			return
		}
		userId := c.MustGet("userId").(uint)
		response, err := rmWatchedSeason(b.db, userId, uint(id))
		if err != nil {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
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
		var user UserRegisterRequest
		if c.ShouldBindJSON(&user) == nil {
			response, err := register(&user, PERM_NONE, b.db)
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
		availableAuthProviders := []string{}
		if Config.JELLYFIN_HOST != "" {
			availableAuthProviders = append(availableAuthProviders, "jellyfin")
		}
		c.JSON(http.StatusOK, &AvailableAuthProvidersResponse{
			AvailableAuthProviders: availableAuthProviders,
			SignupEnabled:          Config.SIGNUP_ENABLED,
			IsInSetup:              ServerInSetup,
		})
	})

	// Request admin token
	auth.Use(AuthRequired(nil)).GET("/admin_token", func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		token, err := createOneUseToken(b.db, TOKENTYPE_ADMIN, userId)
		if err != nil {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
			return
		}
		slog.Info("Admin token generated. Type this token into the web ui to gain admin access on your account.", "token", token, "generated_for", userId)
		c.Status(http.StatusNoContent)
	})

	// Use admin token
	auth.Use(AuthRequired(nil)).POST("/admin_token", func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		var atr UseAdminTokenRequest
		if c.ShouldBindJSON(&atr) == nil {
			err := useAdminToken(&atr, b.db, userId)
			if err != nil {
				c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
				return
			}
			c.Status(http.StatusNoContent)
			return
		}
		c.Status(400)
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

func (b *BaseRouter) addUserRoutes() {
	u := b.rg.Group("/user").Use(AuthRequired(b.db))

	// Get current user info
	u.GET("", func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		response, err := getUserInfo(b.db, userId)
		if err != nil {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	})

	// Update current user settings
	u.POST("/update", func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		var ur UserSettings
		err := c.ShouldBindJSON(&ur)
		if err == nil {
			response, err := userUpdate(b.db, userId, ur)
			if err != nil {
				c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
				return
			}
			c.JSON(http.StatusOK, response)
			return
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	})

	// Get current user setting
	u.GET("/settings", func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		response, err := userGetSettings(b.db, userId)
		if err != nil {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	})

	// Search users
	u.GET("/search/:query", func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		response, err := userSearch(b.db, userId, c.Param("query"))
		if err != nil {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	})

	// Search users
	u.GET("/public/:pubUserId/:pubUsername", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("pubUserId"))
		if err != nil {
			c.Status(400)
			return
		}
		response, err := getUserPublicInfo(b.db, uint(id), c.Param("pubUsername"))
		if err != nil {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	})

	// Update bio
	u.POST("/bio", func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		var br UserBioUpdateRequest
		err := c.ShouldBindJSON(&br)
		if err == nil {
			err := userUpdateBio(b.db, userId, br.NewBio)
			if err != nil {
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
				return
			}
			c.Status(http.StatusOK)
			return
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	})

	// Upload avatar
	u.POST("/avatar", func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		response, err := uploadUserAvatar(c, b.db, userId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	})
}

func (b *BaseRouter) addFollowRoutes() {
	f := b.rg.Group("/follow").Use(AuthRequired(b.db))

	// Get users follows // TODO extend to support optionally passing user id as route param, default to current user
	f.GET("", func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		response, err := getFollows(b.db, userId)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	})

	// Follow a user
	f.POST("/:toFollowId", func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		toFollowId, err := strconv.ParseUint(c.Param("toFollowId"), 10, 64)
		if err != nil {
			slog.Error("failed to convert toFollowId param to uint", "toFollowId", toFollowId)
			c.Status(400)
			return
		}
		response, err := followUser(b.db, userId, uint(toFollowId))
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	})

	// Unfollow a user
	f.DELETE("/:toUnfollowId", func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		toUnfollowId, err := strconv.ParseUint(c.Param("toUnfollowId"), 10, 64)
		if err != nil {
			slog.Error("failed to convert toUnfollowId param to uint", "toUnfollowId", toUnfollowId)
			c.Status(400)
			return
		}
		response, err := unfollowUser(b.db, userId, uint(toUnfollowId))
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	})
}

func (b *BaseRouter) addImportRoutes() {
	imprt := b.rg.Group("/import").Use(AuthRequired(nil))

	imprt.POST("", func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		var ar ImportRequest
		err := c.ShouldBindJSON(&ar)
		if err == nil {
			response, err := importContent(b.db, userId, ar)
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

func (b *BaseRouter) addServerRoutes() {
	server := b.rg.Group("/server").Use(AuthRequired(b.db), AdminRequired())

	// Get server config (minus very sensitive fields, like JWT_SECRET)
	server.GET("/config", func(c *gin.Context) {
		// Return new ServerConfig with only the fields we want to show in settings ui
		c.JSON(http.StatusOK, Config.GetSafe())
	})

	// Update config
	server.POST("/config", func(c *gin.Context) {
		var ur KeyValueRequest
		err := c.ShouldBindJSON(&ur)
		if err == nil {
			err := updateConfig(ur.Key, ur.Value)
			if err != nil {
				c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
				return
			}
			c.Status(http.StatusOK)
			return
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	})

	// Get server stats
	server.GET("/stats", cache.CachePage(b.ms, time.Minute*5, func(c *gin.Context) {
		c.JSON(http.StatusOK, getServerStats(b.db))
	}))
}

func (b *BaseRouter) addFeatureRoutes() {
	feature := b.rg.Group("/features").Use(AuthRequired(b.db))

	// Get enabled features (aka functionality)
	feature.GET("", func(c *gin.Context) {
		c.JSON(http.StatusOK, getEnabledFeatures(c.GetInt("userPermissions")))
	})
}

func (b *BaseRouter) addSonarrRoutes() {
	s := b.rg.Group("/arr/son").Use(AuthRequired(b.db), AdminRequired())

	// Test configuration
	s.POST("/test", func(c *gin.Context) {
		var ur ArrTestParams
		err := c.ShouldBindJSON(&ur)
		if err == nil {
			resp, err := testSonarr(ur)
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
				return
			}
			c.JSON(http.StatusOK, resp)
			return
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	})

	// Used to get config for specific server (quality profile, root folder, etc)
	s.GET("/config/:name", func(c *gin.Context) {
		server, err := getSonarr(c.Param("name"))
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		resp, err := testSonarr(ArrTestParams{Host: server.Host, Key: server.Key})
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	// Add sonarr server into config
	s.POST("/add", func(c *gin.Context) {
		var ur SonarrSettings
		err := c.ShouldBindJSON(&ur)
		if err == nil {
			err := addSonarr(ur)
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
				return
			}
			c.Status(http.StatusOK)
			return
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	})

	// Edit sonarr servers config
	s.POST("/edit", func(c *gin.Context) {
		var ur SonarrSettings
		err := c.ShouldBindJSON(&ur)
		if err == nil {
			err := editSonarr(ur)
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
				return
			}
			c.Status(http.StatusOK)
			return
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	})

	// Remove sonarr server
	s.POST("/rm/:name", func(c *gin.Context) {
		err := rmSonarr(c.Param("name"))
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.Status(http.StatusOK)
	})

	// Get safe config for all sonarr servers
	s.GET("", func(c *gin.Context) {
		response := getSonarrsSafe()
		c.JSON(http.StatusOK, response)
	})

	// Request a show
	s.POST("/request", func(c *gin.Context) {
		var ur arr.SonarrRequest
		err := c.ShouldBindJSON(&ur)
		if err == nil {
			server, err := getSonarr(ur.ServerName)
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
				return
			}
			ur.AutomaticSearch = server.AutomaticSearch
			sonarr := arr.New(arr.SONARR, &server.Host, &server.Key)
			err = sonarr.AddContent(sonarr.BuildAddShowBody(ur))
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
				return
			}
			c.Status(http.StatusOK)
			return
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	})
}

func (b *BaseRouter) addRadarrRoutes() {
	s := b.rg.Group("/arr/rad").Use(AuthRequired(b.db), AdminRequired())

	// Test configuration
	s.POST("/test", func(c *gin.Context) {
		var ur ArrTestParams
		err := c.ShouldBindJSON(&ur)
		if err == nil {
			resp, err := testRadarr(ur)
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
				return
			}
			c.JSON(http.StatusOK, resp)
			return
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	})

	// Get config for specific server
	s.GET("/config/:name", func(c *gin.Context) {
		server, err := getRadarr(c.Param("name"))
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		resp, err := testRadarr(ArrTestParams{Host: server.Host, Key: server.Key})
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	s.POST("/add", func(c *gin.Context) {
		var ur RadarrSettings
		err := c.ShouldBindJSON(&ur)
		if err == nil {
			err := addRadarr(ur)
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
				return
			}
			c.Status(http.StatusOK)
			return
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	})

	s.POST("/edit", func(c *gin.Context) {
		var ur RadarrSettings
		err := c.ShouldBindJSON(&ur)
		if err == nil {
			err := editRadarr(ur)
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
				return
			}
			c.Status(http.StatusOK)
			return
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	})

	s.POST("/rm/:name", func(c *gin.Context) {
		err := rmRadarr(c.Param("name"))
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.Status(http.StatusOK)
	})

	s.GET("", func(c *gin.Context) {
		response := getRadarrsSafe()
		c.JSON(http.StatusOK, response)
	})

	s.POST("/request", func(c *gin.Context) {
		var ur arr.RadarrRequest
		err := c.ShouldBindJSON(&ur)
		if err == nil {
			server, err := getRadarr(ur.ServerName)
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
				return
			}
			ur.AutomaticSearch = server.AutomaticSearch
			radarr := arr.New(arr.RADARR, &server.Host, &server.Key)
			err = radarr.AddContent(radarr.BuildAddMovieBody(ur))
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
				return
			}
			c.Status(http.StatusOK)
			return
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	})
}
