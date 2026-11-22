package plex

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/feature/auth/authmiddleware"
	"github.com/sbondCo/Watcharr/feature/plex/plexmiddleware"
	"github.com/sbondCo/Watcharr/router"
)

type Router struct {
	br *router.BaseRouter
	ss *SyncService
}

func NewRouter(br *router.BaseRouter, ss *SyncService) *Router {
	return &Router{
		br,
		ss,
	}
}

func (r *Router) AddRoutes() {
	plex := r.br.Router.Group("/plex").
		Use(
			authmiddleware.AuthRequired(r.br.Cfg),
			plexmiddleware.PlexAccessRequired(r.br.DB, r.br.Cfg),
		)

	// Sync users plex watched items to watchlist
	plex.GET("/sync", r.GetSync)
}

// Sync users plex watched items to watchlist
func (r *Router) GetSync(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	userPlexLocalAuth := c.MustGet("plexLocalAuthToken").(string)
	response, err := r.ss.PlexSyncWatched(userId, userPlexLocalAuth)
	if err != nil {
		c.JSON(http.StatusForbidden, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}
