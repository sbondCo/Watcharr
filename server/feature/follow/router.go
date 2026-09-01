package follow

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/feature/auth/authmiddleware"
	"github.com/sbondCo/Watcharr/router"
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
	f := r.br.Router.Group("/follow").Use(authmiddleware.AuthRequired(r.br.DB, r.br.Cfg))

	// Get users follows
	f.GET("", r.GetFollows)
	// Follow a user
	f.POST("/:toFollowId", r.AddFollowUser)
	// Unfollow a user
	f.DELETE("/:toUnfollowId", r.DeleteFollow)
	// Get follows thoughts on content
	// TODO Rename `tmdbId` to `mediaId` to match what it is actually used as (since it works for games).
	f.GET("/thoughts/:type/:tmdbId", r.GetFollowsThoughts)
}

// Get users follows // TODO extend to support optionally passing user id as route param, default to current user
func (r *Router) GetFollows(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	response, err := r.service.GetFollows(userId)
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

// Follow a user
func (r *Router) AddFollowUser(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	toFollowId, err := strconv.ParseUint(c.Param("toFollowId"), 10, 64)
	if err != nil {
		slog.Error("failed to convert toFollowId param to uint", "toFollowId", toFollowId)
		c.Status(400)
		return
	}
	response, err := r.service.FollowUser(userId, uint(toFollowId))
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

// Unfollow a user
func (r *Router) DeleteFollow(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	toUnfollowId, err := strconv.ParseUint(c.Param("toUnfollowId"), 10, 64)
	if err != nil {
		slog.Error("failed to convert toUnfollowId param to uint", "toUnfollowId", toUnfollowId)
		c.Status(400)
		return
	}
	response, err := r.service.UnfollowUser(userId, uint(toUnfollowId))
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

// Get follows thoughts on content
func (r *Router) GetFollowsThoughts(c *gin.Context) {
	t := c.Param("type")
	if t != "movie" && t != "tv" && t != "game" && t != "book" {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "only movie, tv, game or book types are supported"})
		return
	}
	userId := c.MustGet("userId").(uint)
	response, err := r.service.GetFollowsThoughts(userId, t, c.Param("tmdbId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}
