package profile

import (
	"net/http"

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
	profile := r.br.Router.Group("/profile").Use(authmiddleware.AuthRequired(r.br.Cfg))

	// Get user profile details
	profile.GET("", r.GetProfile)
}

// Get user profile details
func (r *Router) GetProfile(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	response, err := r.service.getProfile(userId)
	if err != nil {
		c.JSON(http.StatusForbidden, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}
