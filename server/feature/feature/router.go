package feature

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/feature/auth/authmiddleware"
	"github.com/sbondCo/Watcharr/feature/user/usermiddleware"
	"github.com/sbondCo/Watcharr/router"
)

type Router struct {
	*router.BaseRouter
	service *Service
}

func NewRouter(br *router.BaseRouter, service *Service) *Router {
	return &Router{
		br,
		service,
	}
}

func (r *Router) AddRoutes() {
	feature := r.Router.Group("/features").
		Use(authmiddleware.AuthRequired(r.Cfg))

	// Get enabled features (aka functionality)
	feature.GET("", usermiddleware.WithUser(r.DB), func(c *gin.Context) {
		user := usermiddleware.UserFromContext(c)
		c.JSON(http.StatusOK, r.service.GetEnabledFeatures(user.Permissions))
	})
}
