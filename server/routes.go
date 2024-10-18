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
	"github.com/sbondCo/Watcharr/game"
	"gorm.io/gorm"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type ValueRequest struct {
	Value any `json:"value"`
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
	content.GET("/search/multi/:query", cache.CachePage(b.ms, exp, func(c *gin.Context) {
		if c.Param("query") == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "a query was not provided"})
			return
		}
		pageQ := c.Query("page")
		pageNum := 1
		if pageQ != "" {
			num, err := strconv.Atoi(pageQ)
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "query parameter 'page' is not a number"})
				return
			}
			pageNum = num
		}
		content, err := searchContent(c.Param("query"), pageNum)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, content)
	}))

	// Search for movies
	content.GET("/search/movie/:query", cache.CachePage(b.ms, exp, func(c *gin.Context) {
		if c.Param("query") == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "a query was not provided"})
			return
		}
		pageQ := c.Query("page")
		pageNum := 1
		if pageQ != "" {
			num, err := strconv.Atoi(pageQ)
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "query parameter 'page' is not a number"})
				return
			}
			pageNum = num
		}
		content, err := searchMovies(c.Param("query"), pageNum)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, content)
	}))

	// Search for shows
	content.GET("/search/tv/:query", cache.CachePage(b.ms, exp, func(c *gin.Context) {
		if c.Param("query") == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "a query was not provided"})
			return
		}
		pageQ := c.Query("page")
		pageNum := 1
		if pageQ != "" {
			num, err := strconv.Atoi(pageQ)
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "query parameter 'page' is not a number"})
				return
			}
			pageNum = num
		}
		content, err := searchTv(c.Param("query"), pageNum)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, content)
	}))

	// Search for people
	content.GET("/search/person/:query", cache.CachePage(b.ms, exp, func(c *gin.Context) {
		if c.Param("query") == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "a query was not provided"})
			return
		}
		pageQ := c.Query("page")
		pageNum := 1
		if pageQ != "" {
			num, err := strconv.Atoi(pageQ)
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "query parameter 'page' is not a number"})
				return
			}
			pageNum = num
		}
		content, err := searchPeople(c.Param("query"), pageNum)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, content)
	}))

	// Get movie details (for movie page)
	content.GET("/movie/:id", WhereaboutsRequired(), cache.CachePage(b.ms, exp, func(c *gin.Context) {
		if c.Param("id") == "" {
			c.Status(400)
			return
		}
		content, err := movieDetails(b.db, c.Param("id"), c.MustGet("userCountry").(string), map[string]string{"append_to_response": "videos,watch/providers,similar"})
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
	content.GET("/tv/:id", WhereaboutsRequired(), cache.CachePage(b.ms, exp, func(c *gin.Context) {
		if c.Param("id") == "" {
			c.Status(400)
			return
		}
		content, err := tvDetails(b.db, c.Param("id"), c.MustGet("userCountry").(string), map[string]string{"append_to_response": "videos,watch/providers,similar,external_ids,keywords"})
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
	// Supports `watchedId` query parameter for saving the requested season as `LastViewedSeason`.
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
		// If a `watchedId` is passed, we should update it with this season
		// number, so the LastViewedSeason field is up to date (this seemed
		// better than making a new request for just saving this).
		// We will attach a `watcharr-lastviewedseason-saved` header if
		// this part succeeds so the client can decide on showing an error.
		if watchedIdQ := c.Query("watchedId"); watchedIdQ != "" {
			userId := c.MustGet("userId").(uint)
			watchedId, err := strconv.ParseUint(watchedIdQ, 10, 64)
			if err != nil {
				slog.Error("get season details route: Processing watchedId param failed", "error", err.Error(), "id", watchedIdQ)
			} else {
				if seasonNum, err := strconv.ParseInt(c.Param("num"), 10, 64); err == nil {
					if err = updateWatchedLastViewedSeason(b.db, userId, uint(watchedId), int(seasonNum)); err == nil {
						c.Header("watcharr-lastviewedseason-saved", "1")
					}
				} else {
					slog.Error("get season details route: Parsing season number as int failed", "error", err.Error(), "season_num", c.Param("num"))
				}
			}
		} else {
			slog.Debug("get season details route: No watchedId parameter found.. not doing anything.")
		}
		c.JSON(http.StatusOK, content)
	})

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

	// Available regions for watch providers
	content.GET("/regions", func(c *gin.Context) {
		r, err := regions()
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, r)
	})
}

func (b *BaseRouter) addGameRoutes() {
	gamer := b.rg.Group("/game").Use(AuthRequired(nil))
	exp := time.Hour * 24

	igdb := &Config.TWITCH
	igdb.OnTokenRefreshed(func() {
		// Save new token to config when we refresh it.
		slog.Debug("GameRoutes: token refreshed.. saving to config.")
		if err := writeConfig(); err != nil {
			slog.Error("GameRoutes: failed to save refreshed token to config.", "error", err)
		}
	})
	err := igdb.Init()
	// Save cfg if init succeeded, this will save our access token
	if err == nil {
		slog.Error("GameRoutes: Twitch init failed!", "error", err)
	}

	// Search for games
	gamer.GET("/search/:query", cache.CachePage(b.ms, exp, func(c *gin.Context) {
		if c.Param("query") == "" {
			c.Status(400)
			return
		}
		games, err := igdb.Search(c.Param("query"))
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, games)
	}))

	// Game details for game page
	gamer.GET("/:id", cache.CachePage(b.ms, exp, func(c *gin.Context) {
		if c.Param("id") == "" {
			c.Status(400)
			return
		}
		content, err := igdb.GameDetails(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, content)
	}))

	// Add game to played(watched) list
	gamer.POST("/played", func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		var ar PlayedAddRequest
		err := c.ShouldBindJSON(&ar)
		if err == nil {
			response, err := addPlayed(b.db, igdb, userId, ar, ADDED_WATCHED)
			if err != nil {
				c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
				return
			}
			c.JSON(http.StatusOK, response)
			return
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	})

	// IMPORTANT: Routes below only for admins!
	gamer.Use(AuthRequired(b.db), AdminRequired())
	{
		gamer.POST("/config", func(c *gin.Context) {
			var ar game.IGDB
			err := c.ShouldBindJSON(&ar)
			if err == nil {
				err := saveTwitchConfig(ar)
				if err != nil {
					c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
					return
				}
				igdb = &Config.TWITCH
				c.Status(http.StatusOK)
				return
			}
			c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		})
	}
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
		if err == nil {
			userId := c.MustGet("userId").(uint)
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
			response, err := addWatchedSeason(b.db, userId, ar)
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

	watched.POST("/episode", func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		var ar WatchedEpisodeAddRequest
		err := c.ShouldBindJSON(&ar)
		if err == nil {
			response, err := addWatchedEpisodes(b.db, userId, ar)
			if err != nil {
				c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
				return
			}
			c.JSON(http.StatusOK, response)
			return
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	})

	watched.DELETE("/episode/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.Status(400)
			return
		}
		userId := c.MustGet("userId").(uint)
		response, err := rmWatchedEpisode(b.db, userId, uint(id))
		if err != nil {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	})

	watched.POST(":id/tag/:tagId", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			slog.Error("tag watched route failed to convert id param to int", "error", err)
			c.Status(http.StatusBadRequest)
			return
		}
		tagId, err := strconv.Atoi(c.Param("tagId"))
		if err != nil {
			slog.Error("tag watched route failed to convert tagId param to int", "error", err)
			c.Status(http.StatusBadRequest)
			return
		}
		userId := c.MustGet("userId").(uint)
		err = addWatchedToTag(b.db, userId, uint(tagId), uint(id))
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.Status(http.StatusOK)
	})

	watched.DELETE(":id/tag/:tagId", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			slog.Error("tag watched route failed to convert id param to int", "error", err)
			c.Status(http.StatusBadRequest)
			return
		}
		tagId, err := strconv.Atoi(c.Param("tagId"))
		if err != nil {
			slog.Error("tag watched route failed to convert tagId param to int", "error", err)
			c.Status(http.StatusBadRequest)
			return
		}
		userId := c.MustGet("userId").(uint)
		err = rmWatchedFromTag(b.db, userId, uint(tagId), uint(id))
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.Status(http.StatusOK)
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

	activity.PUT(":id", func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.Status(400)
			return
		}
		var activityUpdateRequest ActivityUpdateRequest
		err = c.ShouldBindJSON(&activityUpdateRequest)
		if err == nil {
			err = updateActivity(b.db, userId, uint(id), activityUpdateRequest)
			if err != nil {
				c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
				return
			}
			c.Status(http.StatusOK)
			return
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	})

	activity.DELETE(":id", func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.Status(400)
			slog.Error("Could not process activity id when attempting a deletion", "error", err.Error(), "id", c.Param("id"))
			return
		}
		err = deleteActivity(b.db, userId, uint(id))
		if err != nil {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
			return
		}
		c.Status(http.StatusOK)
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

	// Proxy Login
	auth.POST("/proxy", func(c *gin.Context) {
		var user User

		if Config.PROXY_AUTH_HEADER == "" {
			slog.Error("Request made to login via Proxy, but PROXY_AUTH_HEADER has not been configured.")
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "Proxy authentication disabled"})
			return
		}

		user.Username = c.GetHeader(Config.PROXY_AUTH_HEADER)
                if user.Username == "" {
			slog.Error("Request made to login via Proxy, but authentication header was not provided")
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "Authentication header missing"})
			return
        }

		response, err := loginProxy(&user, b.db)
		if err != nil {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
			return
		}

		c.JSON(http.StatusOK, response)
		return
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

	// Plex login
	auth.POST("/plex", func(c *gin.Context) {
		var plexRequest PlexLoginRequest
		if c.ShouldBindJSON(&plexRequest) == nil {
			response, err := loginPlex(&plexRequest, b.db)
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
		if Config.PLEX_HOST != "" && Config.PLEX_MACHINE_ID != "" {
			availableAuthProviders = append(availableAuthProviders, "plex")
		}
		if Config.PROXY_AUTH_HEADER != "" {
			availableAuthProviders = append(availableAuthProviders, "proxy")
		}

		c.JSON(http.StatusOK, &AvailableAuthProvidersResponse{
			AvailableAuthProviders: availableAuthProviders,
			SignupEnabled:          Config.SIGNUP_ENABLED,
			IsInSetup:              ServerInSetup,
			UseEmby:                Config.USE_EMBY,
		})
	})

	// IMPORTANT: Routes below here must be authenticated.
	auth.Use(AuthRequired(nil))
	{
		// Request admin token
		auth.GET("/admin_token", func(c *gin.Context) {
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
		auth.POST("/admin_token", func(c *gin.Context) {
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

		// Change password
		auth.POST("/change_password", func(c *gin.Context) {
			userId := c.MustGet("userId").(uint)
			var pwds UserPasswordUpdateRequest
			err := c.ShouldBindJSON(&pwds)
			if err == nil {
				err := userChangePassword(b.db, pwds, userId)
				if err != nil {
					c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
					return
				}
				c.Status(http.StatusOK)
				return
			}
			c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		})
	}
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
	jf := b.rg.Group("/jellyfin").Use(AuthRequired(b.db), JellyfinAccessRequired())

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

	// Sync users jellyfin watched items to watchlist
	jf.GET("/sync", func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		userType := c.MustGet("userType").(UserType)
		username := c.MustGet("username").(string)
		userThirdPartyId := c.MustGet("userThirdPartyId").(string)
		userThirdPartyAuth := c.MustGet("userThirdPartyAuth").(string)
		response, err := jellyfinSyncWatched(b.db, userId, userType, username, userThirdPartyId, userThirdPartyAuth)
		if err != nil {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	})
}

func (b *BaseRouter) addPlexRoutes() {
	plex := b.rg.Group("/plex").Use(AuthRequired(b.db), PlexAccessRequired(b.db))

	// Sync users plex watched items to watchlist
	plex.GET("/sync", func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		userPlexLocalAuth := c.MustGet("plexLocalAuthToken").(string)
		response, err := plexSyncWatched(b.db, userId, userPlexLocalAuth)
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

	// Get follows thoughts on content
	f.GET("/thoughts/:type/:tmdbId", func(c *gin.Context) {
		t := c.Param("type")
		if t != "movie" && t != "tv" && t != "game" {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "only movie, tv or game types are supported"})
			return
		}
		userId := c.MustGet("userId").(uint)
		response, err := getFollowsThoughts(b.db, userId, t, c.Param("tmdbId"))
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

	imprt.POST("/trakt", func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		var ar TraktImportRequest
		err := c.ShouldBindJSON(&ar)
		if err == nil {
			response, err := traktImportWatched(b.db, userId, ar.Username)
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

	// Update plex host config
	server.POST("/config/plex_host", func(c *gin.Context) {
		var ur ValueRequest
		err := c.ShouldBindJSON(&ur)
		if err == nil {
			resp, err := updateConfigPlexHost(ur.Value.(string))
			if err != nil {
				c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
				return
			}
			c.JSON(http.StatusOK, resp)
			return
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	})

	// Get server stats
	server.GET("/stats", cache.CachePage(b.ms, time.Minute*5, func(c *gin.Context) {
		c.JSON(http.StatusOK, getServerStats(b.db))
	}))

	// Get all server users (for manage users page)
	server.GET("/users", func(c *gin.Context) {
		resp, err := getAllUsers(b.db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	// Edit a user (for manage users page)
	server.POST("/users/:id", func(c *gin.Context) {
		userId, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			slog.Error("/users/:id failed to parse id as a uint", "error", err)
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to parse id"})
			return
		}
		var ur UpdateUserRequest
		err = c.ShouldBindJSON(&ur)
		if err == nil {
			err := manageUser(b.db, uint(userId), ur)
			if err != nil {
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
				return
			}
			c.Status(http.StatusOK)
			return
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	})
}

func (b *BaseRouter) addFeatureRoutes() {
	feature := b.rg.Group("/features").Use(AuthRequired(b.db))

	// Get enabled features (aka functionality)
	feature.GET("", func(c *gin.Context) {
		c.JSON(http.StatusOK, getEnabledFeatures(c.GetInt("userPermissions")))
	})
}

func (b *BaseRouter) addSonarrRoutes() {
	s := b.rg.Group("/arr/son").Use(AuthRequired(b.db))

	// Routes are manually given AdminRequired or PermRequired middleware.

	// Test configuration
	s.POST("/test", AdminRequired(), func(c *gin.Context) {
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
	s.GET("/config/:name", PermRequired(PERM_REQUEST_CONTENT), func(c *gin.Context) {
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
	s.POST("/add", AdminRequired(), func(c *gin.Context) {
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
	s.POST("/edit", AdminRequired(), func(c *gin.Context) {
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
	s.POST("/rm/:name", AdminRequired(), func(c *gin.Context) {
		err := rmSonarr(c.Param("name"))
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.Status(http.StatusOK)
	})

	// Get safe config for all sonarr servers
	s.GET("", PermRequired(PERM_REQUEST_CONTENT), func(c *gin.Context) {
		response := getSonarrsSafe()
		c.JSON(http.StatusOK, response)
	})

	// Request a show
	s.POST("/request", PermRequired(PERM_REQUEST_CONTENT), func(c *gin.Context) {
		var ur arr.SonarrRequest
		err := c.ShouldBindJSON(&ur)
		if err == nil {
			userId := c.MustGet("userId").(uint)
			perms := c.GetInt("userPermissions")
			response, err := createSonarrRequest(b.db, userId, perms, ur)
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
				return
			}
			c.JSON(http.StatusOK, response)
			return
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	})

	s.GET("/request/:tmdbId", PermRequired(PERM_REQUEST_CONTENT), func(c *gin.Context) {
		tmdbId, err := strconv.Atoi(c.Param("tmdbId"))
		if err != nil {
			slog.Error("Couldn't parse tmdbId", "tmdbId", tmdbId)
			c.Status(400)
			return
		}
		response, err := getArrRequestByTmdbId(b.db, SHOW, tmdbId)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	})

	s.POST("/request/approve/:id", PermRequired(PERM_ADMIN), func(c *gin.Context) {
		var ur arr.SonarrRequest
		err := c.ShouldBindJSON(&ur)
		if err == nil {
			requestId, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				slog.Error("Couldn't parse request id", "request_id", requestId)
				c.Status(400)
				return
			}
			response, err := approveSonarrRequest(b.db, uint(requestId), ur)
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
				return
			}
			c.JSON(http.StatusOK, response)
			return
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	})

	s.GET("/status/:serverName/:arrId", PermRequired(PERM_REQUEST_CONTENT), func(c *gin.Context) {
		response, err := getSonarrQueueDetails(c.Param("serverName"), c.Param("arrId"))
		if err != nil {
			if err.Error() == "no details found" {
				c.Status(http.StatusNoContent) // Item not found in queue.. missing
				return
			}
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	})

	s.GET("/info/:requestId", PermRequired(PERM_REQUEST_CONTENT), func(c *gin.Context) {
		requestId, err := strconv.ParseUint(c.Param("requestId"), 10, 64)
		if err != nil {
			slog.Error("/info/:requestId - requestId could not be parsed", "requestId", requestId)
			c.Status(http.StatusBadRequest)
			return
		}
		response, err := getSonarrRequestInfo(b.db, uint(requestId))
		if err != nil {
			if err.Error() == "request deleted" {
				c.JSON(http.StatusNotFound, ErrorResponse{Error: "request deleted"})
				return
			}
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	})
}

func (b *BaseRouter) addRadarrRoutes() {
	s := b.rg.Group("/arr/rad").Use(AuthRequired(b.db))

	// Routes are manually given AdminRequired or PermRequired middleware.

	// Test configuration
	s.POST("/test", AdminRequired(), func(c *gin.Context) {
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
	s.GET("/config/:name", PermRequired(PERM_REQUEST_CONTENT), func(c *gin.Context) {
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

	s.POST("/add", AdminRequired(), func(c *gin.Context) {
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

	s.POST("/edit", AdminRequired(), func(c *gin.Context) {
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

	s.POST("/rm/:name", AdminRequired(), func(c *gin.Context) {
		err := rmRadarr(c.Param("name"))
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.Status(http.StatusOK)
	})

	s.GET("", PermRequired(PERM_REQUEST_CONTENT), func(c *gin.Context) {
		response := getRadarrsSafe()
		c.JSON(http.StatusOK, response)
	})

	s.POST("/request", PermRequired(PERM_REQUEST_CONTENT), func(c *gin.Context) {
		var ur arr.RadarrRequest
		err := c.ShouldBindJSON(&ur)
		if err == nil {
			userId := c.MustGet("userId").(uint)
			perms := c.GetInt("userPermissions")
			response, err := createRadarrRequest(b.db, userId, perms, ur)
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
				return
			}
			c.JSON(http.StatusOK, response)
			return
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	})

	s.GET("/request/:tmdbId", PermRequired(PERM_REQUEST_CONTENT), func(c *gin.Context) {
		tmdbId, err := strconv.Atoi(c.Param("tmdbId"))
		if err != nil {
			slog.Error("Couldn't parse tmdbId", "tmdbId", tmdbId)
			c.Status(400)
			return
		}
		response, err := getArrRequestByTmdbId(b.db, MOVIE, tmdbId)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	})

	s.POST("/request/approve/:id", PermRequired(PERM_ADMIN), func(c *gin.Context) {
		var ur arr.RadarrRequest
		err := c.ShouldBindJSON(&ur)
		if err == nil {
			requestId, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				slog.Error("Couldn't parse request id", "request_id", requestId)
				c.Status(400)
				return
			}
			response, err := approveRadarrRequest(b.db, uint(requestId), ur)
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
				return
			}
			c.JSON(http.StatusOK, response)
			return
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	})

	s.GET("/status/:serverName/:arrId", PermRequired(PERM_REQUEST_CONTENT), func(c *gin.Context) {
		response, err := getRadarrQueueDetails(c.Param("serverName"), c.Param("arrId"))
		if err != nil {
			if err.Error() == "no details found" {
				c.Status(http.StatusNoContent) // Item not found in queue.. missing
				return
			}
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	})

	s.GET("/info/:requestId", PermRequired(PERM_REQUEST_CONTENT), func(c *gin.Context) {
		requestId, err := strconv.ParseUint(c.Param("requestId"), 10, 64)
		if err != nil {
			slog.Error("/info/:requestId - requestId could not be parsed", "requestId", requestId)
			c.Status(http.StatusBadRequest)
			return
		}
		response, err := getRadarrRequestInfo(b.db, uint(requestId))
		if err != nil {
			if err.Error() == "request deleted" {
				c.JSON(http.StatusNotFound, ErrorResponse{Error: "request deleted"})
				return
			}
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	})
}

func (b *BaseRouter) addArrRequestRoutes() {
	s := b.rg.Group("/arr/request").Use(AuthRequired(b.db))

	// Get all requests (for manage_requests view), only for admins.
	s.GET("/", AdminRequired(), func(c *gin.Context) {
		response, err := getArrRequests(b.db)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	})

	// Deny a request (for manage_requests view), only for admins.
	s.POST("/deny/:id", AdminRequired(), func(c *gin.Context) {
		requestId, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			slog.Error("Couldn't parse request id", "request_id", requestId)
			c.Status(400)
			return
		}
		err = denyArrRequest(b.db, uint(requestId))
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}
		c.Status(http.StatusOK)
	})
}

func (b *BaseRouter) addJobRoutes() {
	job := b.rg.Group("/job").Use(AuthRequired(nil))

	// Uses wildcard so it still works in cases where the job id includes a /.
	// (yes i changed this instead of not allowing a / when we generate a job id becuz easier)
	job.GET("/*id", func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		// When we get id param, don't include first letter, which will be the beginning '/'.
		response, err := getJob(c.Param("id")[1:], userId)
		if err != nil {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, *response)
	})
}

func (b *BaseRouter) addTaskRoutes() {
	task := b.rg.Group("/task").Use(AuthRequired(b.db), AdminRequired())

	task.GET("/", func(c *gin.Context) {
		response := getAllTasks()
		c.JSON(http.StatusOK, response)
	})

	task.PUT(":name", func(c *gin.Context) {
		if c.Param("name") == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "no task name provided"})
			return
		}
		var rr TaskRescheduleRequest
		err := c.ShouldBindJSON(&rr)
		if err == nil {
			err := rescheduleTask(c.Param("name"), rr)
			if err != nil {
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
				return
			}
			c.Status(http.StatusOK)
			return
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	})
}

func (b *BaseRouter) addTagRoutes() {
	tag := b.rg.Group("/tag").Use(AuthRequired(nil))

	// Get list of all our tags.
	tag.GET("", func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		tags, err := getTags(b.db, userId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, tags)
	})

	// // Get all items within one of our tags.
	// tag.GET(":id", func(c *gin.Context) {
	// 	id, err := strconv.Atoi(c.Param("id"))
	// 	if err != nil {
	// 		slog.Error("getTag route failed to convert id param to int", "error", err)
	// 		c.Status(http.StatusBadRequest)
	// 		return
	// 	}
	// 	userId := c.MustGet("userId").(uint)
	// 	tags, err := getTag(b.db, userId, uint(id))
	// 	if err != nil {
	// 		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	// 		return
	// 	}
	// 	c.JSON(http.StatusOK, tags)
	// })

	// Create a tag.
	tag.POST("", func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		var tr TagAddRequest
		err := c.ShouldBindJSON(&tr)
		if err == nil {
			response, err := addTag(b.db, userId, tr)
			if err != nil {
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
				return
			}
			c.JSON(http.StatusOK, response)
			return
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	})

	// Update a tag.
	tag.PUT(":id", func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.Status(400)
			slog.Error("tag update rote: failed to process tag id.", "error", err.Error(), "id", c.Param("id"))
			return
		}
		var tr TagAddRequest
		err = c.ShouldBindJSON(&tr)
		if err == nil {
			err := updateTag(b.db, userId, uint(id), tr)
			if err != nil {
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
				return
			}
			c.Status(http.StatusOK)
			return
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	})

	tag.DELETE(":id", func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.Status(400)
			slog.Error("tag delete rote: failed to process tag id.", "error", err.Error(), "id", c.Param("id"))
			return
		}
		err = deleteTag(b.db, userId, uint(id))
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}
		c.Status(http.StatusOK)
	})
}
