package tag

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/feature/auth/authmiddleware"
	"github.com/sbondCo/Watcharr/router"
	"github.com/sbondCo/Watcharr/util"
)

type Router struct {
	br      *router.BaseRouter
	service *Service
}

func NewRouter(br *router.BaseRouter, service *Service) *Router {
	return &Router{
		br,
		service,
	}
}

func (r *Router) AddRoutes() {
	tag := r.br.Router.Group("/tag").Use(authmiddleware.AuthRequired(nil, r.br.Cfg))

	tag.GET("", r.GetTags)
	tag.GET(":id", r.GetTag)
	tag.GET(":id/watched", router.PaginatedRequest(true), r.GetTagWatched)
	tag.POST("", r.CreateTag)
	tag.PUT(":id", r.UpdateTag)
	tag.DELETE(":id", r.DeleteTag)
}

// Get all of our tags.
func (r *Router) GetTags(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	tags, err := r.service.GetTags(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, tags)
}

// Get all items within one of our tags.
func (r *Router) GetTag(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		slog.Error("getTag route failed to convert id param to int", "error", err)
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "invalid id"})
		return
	}
	tags, err := r.service.GetTag(userId, uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, tags)
}

// Get watched items in our tag with pagination.
func (r *Router) GetTagWatched(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		slog.Error("getTagWatched route failed to convert id param to int", "error", err)
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "invalid id"})
		return
	}
	pp := c.MustGet("paginationParams").(util.PaginationParams)
	wp := domain.WatchedGetPageRequest{
		// Defaults..
		Sort:    domain.WatchedSortDateAdded,
		SortDir: domain.WatchedSortDirAsc,
	}
	if err := c.ShouldBind(&wp); err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "failed to get request parameters"})
		return
	}
	if wp, err := r.service.GetTagPage(userId, uint(id), pp, wp); err == nil {
		dto := util.PaginationResponse[domain.Media, util.None]{
			PaginationParams: wp.PaginationParams,
			TotalPages:       wp.TotalPages,
			TotalResults:     wp.TotalResults,
			Results:          domain.NewWatchedGetPageResponse(wp.Results),
		}
		c.JSON(http.StatusOK, dto)
	} else {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "failed to get page"})
	}
}

// Create a tag.
func (r *Router) CreateTag(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	var tr domain.TagAddRequest
	err := c.ShouldBindJSON(&tr)
	if err == nil {
		response, err := r.service.AddTag(userId, tr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
		return
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
}

func (r *Router) UpdateTag(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.Status(400)
		slog.Error("tag update rote: failed to process tag id.", "error", err.Error(), "id", c.Param("id"))
		return
	}
	var tr domain.TagAddRequest
	err = c.ShouldBindJSON(&tr)
	if err == nil {
		err := r.service.UpdateTag(userId, uint(id), tr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: err.Error()})
			return
		}
		c.Status(http.StatusOK)
		return
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
}

func (r *Router) DeleteTag(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.Status(400)
		slog.Error("tag delete rote: failed to process tag id.", "error", err.Error(), "id", c.Param("id"))
		return
	}
	err = r.service.DeleteTag(userId, uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.Status(http.StatusOK)
}
