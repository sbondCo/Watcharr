package episode

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/feature/auth/authmiddleware"
	"github.com/sbondCo/Watcharr/router"
)

type Router struct {
	br *router.BaseRouter
	s  *Service
}

func NewRouter(
	br *router.BaseRouter,
	service *Service,
) *Router {
	return &Router{
		br: br,
		s:  service,
	}
}

func (r *Router) AddRoutes() {
	episode := r.br.Router.Group("/watched/episode").Use(authmiddleware.AuthRequired(nil, r.br.Cfg))

	episode.POST("", r.AddWatchedEpisode)
	episode.DELETE(":id", r.DeleteWatchedEpisode)
}

func (r *Router) AddWatchedEpisode(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	var ar domain.WatchedEpisodeAddRequest
	err := c.ShouldBindJSON(&ar)
	if err == nil {
		response, err := r.s.AddWatchedEpisodes(userId, ar)
		if err != nil {
			c.JSON(http.StatusForbidden, router.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
		return
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
}

func (r *Router) DeleteWatchedEpisode(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(400)
		return
	}
	userId := c.MustGet("userId").(uint)
	response, err := r.s.rmWatchedEpisode(userId, uint(id))
	if err != nil {
		c.JSON(http.StatusForbidden, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}
