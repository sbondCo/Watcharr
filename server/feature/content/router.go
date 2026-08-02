package content

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/cache"
	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/feature/auth/authmiddleware"
	"github.com/sbondCo/Watcharr/feature/watched/addedtocontent"
	"github.com/sbondCo/Watcharr/media/tmdb"
	"github.com/sbondCo/Watcharr/router"
	"github.com/sbondCo/Watcharr/util"
)

type WatchedProvider interface {
	UpdateWatchedLastViewedSeason(userId uint, id uint, seasonNum int) error
	GetWatchedItemBySupportedMediaId(userId uint, id uint, t util.SupportedMedia) (entity.Watched, error)
	GetWatchedItemsBySupportedMediaIds(userId uint, c []addedtocontent.IdToTypePair) ([]entity.Watched, error)
}

type Router struct {
	br   *router.BaseRouter
	cs   *Service
	wp   WatchedProvider
	tmdb *tmdb.TMDB
}

func NewRouter(
	br *router.BaseRouter,
	cs *Service,
	wp WatchedProvider,
	tmdb *tmdb.TMDB,
) *Router {
	return &Router{
		br:   br,
		cs:   cs,
		wp:   wp,
		tmdb: tmdb,
	}
}

func (r *Router) AddRoutes() {
	content := r.br.Router.Group("/content").Use(authmiddleware.AuthRequired(nil, r.br.Cfg))
	exp := time.Hour * 24

	// NOTE: Some routes use `cache.CachePage`, but others that contain user watched data
	// don't and rather have their caching on the TMDB methods directly.

	// Get movie details (for movie page)
	content.GET("/movie/:id", router.WhereaboutsRequired(r.br.Cfg), r.GetMovieDetails)
	// Get movie cast
	content.GET("/movie/:id/credits", cache.CachePage(r.br.MemStore, exp, r.GetMovieCredits))
	// Get tv details (for tv page)
	content.GET("/tv/:id", router.WhereaboutsRequired(r.br.Cfg), r.GetTvDetails)
	// Get tv cast
	content.GET("/tv/:id/credits", cache.CachePage(r.br.MemStore, exp, r.GetTvCredits))
	// Get season details
	// Supports `watchedId` query parameter for saving the requested season as `LastViewedSeason`.
	content.GET("/tv/:id/season/:num", r.GetSeasonDetails)
	// Get person details
	content.GET("/person/:id", cache.CachePage(r.br.MemStore, exp, r.GetPerson))
	// Get person credits
	content.GET("/person/:id/credits", r.GetPersonCredits)
	// Available regions for watch providers
	content.GET("/regions", r.GetRegions)
}

func (r *Router) GetMovieDetails(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	if c.Param("id") == "" {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "an id was not provided"})
		return
	}
	content, err := r.tmdb.MovieDetails(tmdb.MovieDetailsOptions{
		ID:      c.Param("id"),
		Country: c.MustGet("userCountry").(string),
		Params: map[string]string{
			"append_to_response": "videos,watch/providers,similar",
		},
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	contentAsMedia := content.AsMedia()
	if err := addedtocontent.AddSingularAndList(
		r.wp,
		userId,
		contentAsMedia,
		func(w *entity.Watched) {
			contentAsMedia.Watched = domain.NewWatchedDtoForContentPage(w)
		},
		[]*addedtocontent.AddListCall[domain.Media]{
			addedtocontent.NewAddListCall(
				contentAsMedia.Similar,
				func(i int, w *entity.Watched) {
					contentAsMedia.Similar[i].Watched = domain.NewWatchedDtoForLists(w)
				},
			),
		},
	); err != nil {
		slog.Error("GetMovieDetails: Failed to add watched to content!", "error", err)
		c.JSON(
			http.StatusInternalServerError,
			router.ErrorResponse{Error: "failed to add watched data to response"},
		)
		return
	}
	c.JSON(http.StatusOK, contentAsMedia)
}

func (r *Router) GetMovieCredits(c *gin.Context) {
	if c.Param("id") == "" {
		c.Status(400)
		return
	}
	content, err := r.tmdb.MovieCredits(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, content)
}

func (r *Router) GetTvDetails(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	if c.Param("id") == "" {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "an id was not provided"})
		return
	}
	// 1. Get details
	content, err := r.tmdb.ShowDetails(tmdb.ShowDetailsOptions{
		ID:      c.Param("id"),
		Country: c.MustGet("userCountry").(string),
		Params: map[string]string{
			"append_to_response": "videos,watch/providers,similar,external_ids,keywords",
		},
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	contentAsMedia := content.AsMedia()
	if err := addedtocontent.AddSingularAndList(
		r.wp,
		userId,
		contentAsMedia,
		func(w *entity.Watched) {
			contentAsMedia.Watched = domain.NewWatchedDtoForContentPage(w)
		},
		[]*addedtocontent.AddListCall[domain.Media]{
			addedtocontent.NewAddListCall(
				contentAsMedia.Similar,
				func(i int, w *entity.Watched) {
					contentAsMedia.Similar[i].Watched = domain.NewWatchedDtoForLists(w)
				},
			),
		},
	); err != nil {
		slog.Error("GetTvDetails: Failed to add watched to content!", "error", err)
		c.JSON(
			http.StatusInternalServerError,
			router.ErrorResponse{Error: "failed to add watched data to response"},
		)
		return
	}
	c.JSON(http.StatusOK, contentAsMedia)
}

func (r *Router) GetTvCredits(c *gin.Context) {
	if c.Param("id") == "" {
		c.Status(400)
		return
	}
	content, err := r.tmdb.ShowCredits(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, content)
}

// Get season details
// Supports `watchedId` query parameter for saving the requested season as `LastViewedSeason`.
func (r *Router) GetSeasonDetails(c *gin.Context) {
	if c.Param("id") == "" || c.Param("num") == "" {
		c.Status(400)
		return
	}
	content, err := r.tmdb.SeasonDetails(c.Param("id"), c.Param("num"))
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
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
				if err = r.wp.UpdateWatchedLastViewedSeason(
					userId, uint(watchedId), int(seasonNum),
				); err == nil {
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
}

func (r *Router) GetPerson(c *gin.Context) {
	if c.Param("id") == "" {
		c.Status(400)
		return
	}
	content, err := r.tmdb.PersonDetails(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, content.AsPersonDetailsResponse())
}

func (r *Router) GetPersonCredits(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	if c.Param("id") == "" {
		c.Status(400)
		return
	}
	content, err := r.tmdb.PersonCredits(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	// Add content into response struct then add watched entries
	resp := domain.PersonCreditsResponse{}
	for i := range content.Cast {
		resp.Credits = append(resp.Credits, content.Cast[i].AsMedia())
	}
	if err := addedtocontent.AddList(
		r.wp,
		userId,
		resp.Credits,
		func(i int, w *entity.Watched) {
			resp.Credits[i].Watched = domain.NewWatchedDtoForLists(w)
		},
	); err != nil {
		slog.Error("GetPersonCredits: Failed to add watched to content!", "error", err)
		c.JSON(
			http.StatusInternalServerError,
			router.ErrorResponse{Error: "failed to add watched data to response"},
		)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (r *Router) GetRegions(c *gin.Context) {
	re, err := r.tmdb.Regions()
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, re)
}
