package watched

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/feature/auth/authmiddleware"
	"github.com/sbondCo/Watcharr/media/tmdb"
	"github.com/sbondCo/Watcharr/router"
	"github.com/sbondCo/Watcharr/util"
)

type Router struct {
	br *router.BaseRouter
	t  *tmdb.TMDB
	s  *Service
}

func NewRouter(
	br *router.BaseRouter,
	t *tmdb.TMDB,
	service *Service,
) *Router {
	return &Router{
		br: br,
		t:  t,
		s:  service,
	}
}

func (r *Router) AddRoutes() {
	watched := r.br.Router.Group("/watched").Use(authmiddleware.AuthRequired(nil, r.br.Cfg))

	watched.GET("", router.PaginatedRequest(false), r.GetWatchedList)
	watched.GET(":id/:username", router.PaginatedRequest(true), r.GetPublicWatchedList)
	watched.POST("", r.AddWatched)
	watched.PUT(":id", r.UpdateWatched)
	watched.DELETE(":id", r.DeleteWatched)
	// TODO Move add/delete watched from tag to the `tag` package (the service code is there so the route may as well be under there, also avoids a circular dep).
	watched.POST(":id/tag/:tagId", r.AddWatchedToTag)
	watched.DELETE(":id/tag/:tagId", r.DeleteWatchedFromTag)
}

// Get our (logged in user) watched list.
func (r *Router) GetWatchedList(c *gin.Context) {
	isPaginated := c.MustGet("paginationEnabled").(bool)
	userId := c.MustGet("userId").(uint)
	if isPaginated {
		pp := c.MustGet("paginationParams").(util.PaginationParams)
		wpr := domain.WatchedGetPageRequest{
			// Defaults..
			Sort:    domain.WatchedSortDateAdded,
			SortDir: domain.WatchedSortDirAsc,
		}
		if err := c.ShouldBind(&wpr); err != nil {
			c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "failed to get request parameters"})
			return
		}
		wp, err := r.s.GetWatchedPage(userId, pp, wpr, nil)
		if err != nil {
			c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "failed to get page"})
		}
		dto := util.PaginationResponse[domain.Media, util.None]{
			PaginationParams: wp.PaginationParams,
			TotalPages:       wp.TotalPages,
			TotalResults:     wp.TotalResults,
			Results:          domain.NewWatchedGetPageResponse(wp.Results),
		}
		c.JSON(http.StatusOK, dto)
		return
	}
	// Non paginated response (doesn't support sorting/filtering)
	// This just exists for backwards compatibility or for
	// downloading the entire list unmodified.
	if w, err := r.s.getWatched(userId); err == nil {
		c.JSON(http.StatusOK, w)
	} else {
		c.JSON(http.StatusForbidden, router.ErrorResponse{Error: "failed"})
	}
}

// Get another users watched list (if its public).
func (r *Router) GetPublicWatchedList(c *gin.Context) {
	pp := c.MustGet("paginationParams").(util.PaginationParams)
	wpr := domain.WatchedGetPageRequest{
		// Defaults..
		Sort:    domain.WatchedSortDateAdded,
		SortDir: domain.WatchedSortDirAsc,
	}
	if err := c.ShouldBind(&wpr); err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "failed to get request parameters"})
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		slog.Error("getPublicWatched route failed to convert id param to uint", "id", id)
		c.Status(400)
		return
	}
	wp, err := r.s.getPublicWatched(uint(id), c.Param("username"), pp, wpr)
	if err != nil {
		c.JSON(http.StatusForbidden, router.ErrorResponse{Error: err.Error()})
		return
	}
	dto := util.PaginationResponse[domain.Media, util.None]{
		PaginationParams: wp.PaginationParams,
		TotalPages:       wp.TotalPages,
		TotalResults:     wp.TotalResults,
		Results:          domain.NewWatchedPublicGetPageResponse(wp.Results),
	}
	c.JSON(http.StatusOK, dto)
}

func (r *Router) AddWatched(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	var ar domain.WatchedAddRequest
	err := c.ShouldBindJSON(&ar)
	if err == nil {
		response, err := r.s.AddWatched(userId, ar, entity.ADDED_WATCHED)
		if err != nil {
			c.JSON(http.StatusForbidden, router.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
		return
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
}

func (r *Router) UpdateWatched(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(400)
		return
	}
	userId := c.MustGet("userId").(uint)
	var ur domain.WatchedUpdateRequest
	err = c.ShouldBindJSON(&ur)
	if err == nil {
		response, err := r.s.updateWatched(userId, uint(id), ur)
		if err != nil {
			c.JSON(http.StatusForbidden, router.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
		return
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
}

func (r *Router) DeleteWatched(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err == nil {
		userId := c.MustGet("userId").(uint)
		response, err := r.s.removeWatched(userId, uint(id))
		if err != nil {
			c.JSON(http.StatusForbidden, router.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
		return
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
}

func (r *Router) AddWatchedToTag(c *gin.Context) {
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
	err = AddWatchedToTag(r.br.DB, userId, uint(tagId), uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (r *Router) DeleteWatchedFromTag(c *gin.Context) {
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
	err = RmWatchedFromTag(r.br.DB, userId, uint(tagId), uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.Status(http.StatusOK)
}
