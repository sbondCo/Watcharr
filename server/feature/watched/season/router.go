package season

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
	season := r.br.Router.Group("/watched/season").Use(authmiddleware.AuthRequired(nil, r.br.Cfg))

	season.POST("", r.AddWatchedSeason)
	season.DELETE(":id", r.DeleteWatchedSeason)
}

func (r *Router) AddWatchedSeason(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	var ar domain.WatchedSeasonAddRequest
	err := c.ShouldBindJSON(&ar)
	if err == nil {
		response, err := r.s.AddWatchedSeason(userId, ar)
		if err != nil {
			c.JSON(http.StatusForbidden, router.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
		return
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
}

func (r *Router) DeleteWatchedSeason(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(400)
		return
	}
	userId := c.MustGet("userId").(uint)
	response, err := r.s.RmWatchedSeason(userId, uint(id))
	if err != nil {
		c.JSON(http.StatusForbidden, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}
