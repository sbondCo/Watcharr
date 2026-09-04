package jellyfin

import (
	"log/slog"
	"net/http"
	"uuid"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/feature/auth/authmiddleware"
	"github.com/sbondCo/Watcharr/feature/user/usermiddleware"
	"github.com/sbondCo/Watcharr/router"
)

type Router struct {
	br             *router.BaseRouter
	s              *Service
	syncService    *SyncService
	webhookService *WebhookService
}

func NewRouter(
	br *router.BaseRouter,
	s *Service,
	syncService *SyncService,
	webhookService *WebhookService,
) *Router {
	return &Router{
		br:             br,
		s:              s,
		syncService:    syncService,
		webhookService: webhookService,
	}
}

func (r *Router) AddRoutes() {
	jf := r.br.Router.Group("/jellyfin")

	// Unauthenticated.
	{
		// Jellyfin webhook data ingest endpoint.
		jf.POST("/webhook/:uuid", r.PostWebhook)
	}

	// Authenticated and jellyfin access required.
	jf.Use(
		authmiddleware.AuthRequired(r.br.Cfg),
		usermiddleware.WithUser(r.br.DB),
		usermiddleware.WithUserServiceByName(r.br.DB, entity.UserServiceNameJellyfin),
		authmiddleware.JellyfinAccessRequired(r.br.Cfg),
	)
	{
		// Check if jf has item
		jf.GET("/:type/:name/:tmdbId", r.GetFindContent)
		// Sync users jellyfin watched items to watchlist
		jf.GET("/sync", r.GetSync)
	}
}

// Check if jf has item
func (r *Router) GetFindContent(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	user := usermiddleware.UserFromContext(c)
	userJellyfinService := usermiddleware.UserServiceFromContext(
		c, entity.UserServiceNameJellyfin)
	response, err := r.s.JellyfinContentFind(
		userId,
		user.Username,
		userJellyfinService.ClientID,
		userJellyfinService.AuthToken,
		c.Param("type"),
		c.Param("name"),
		c.Param("tmdbId"),
	)
	if err != nil {
		c.JSON(http.StatusForbidden, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

// Sync users jellyfin watched items to watchlist
func (r *Router) GetSync(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	user := usermiddleware.UserFromContext(c)
	userJellyfinService := usermiddleware.UserServiceFromContext(
		c, entity.UserServiceNameJellyfin)
	response, err := r.syncService.jellyfinSyncWatched(
		userId,
		user.Username,
		userJellyfinService.ClientID,
		userJellyfinService.AuthToken,
	)
	if err != nil {
		c.JSON(http.StatusForbidden, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

// (unauthenticated) Webhook endpoint.
// Ingest service will use the `uuid` param as a "secret".
func (r *Router) PostWebhook(c *gin.Context) {
	if !r.br.Cfg.JELLYFIN_WEBHOOK_ENABLED {
		slog.Warn("PostWebhook: Webhook support is disabled.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	uuidParam := c.Param("uuid")
	if uuidParam == "" {
		slog.Warn("PostWebhook: No key given. Ensure the address is fully and correctly entered into Jellyfin!")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	uuidKey, err := uuid.Parse(uuidParam)
	if err != nil {
		slog.Warn("PostWebhook: Invalid key given.", "error", err)
		// A user trying to debug why this is failing may appreciate this log.
		// Only in Debug log, since it is a credential.
		slog.Debug("PostWebhook: Invalid key given.", "key_given", uuidParam)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if uuidKey != r.br.Cfg.JELLYFIN_WEBHOOK_KEY {
		slog.Warn("PostWebhook: Wrong key given.")
		// A user trying to debug why this is failing may appreciate this log.
		// Only in Debug log, since it is a credential.
		slog.Debug("PostWebhook: Wrong key given.", "key_given", uuidParam)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var data WebhookData
	err = c.ShouldBindJSON(&data)
	if err != nil {
		slog.Warn("PostWebhook: Bad request recieved.", "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest,
			router.ErrorResponse{Error: err.Error()})
		return
	}
	err = r.webhookService.Ingest(data)
	if err != nil {
		// Since the error data might be sensitive (regarding auth/if a user
		// exists), we just return a generic "failed" message.
		// The real error should be obtained from watcharrs logs, since this
		// endpoint is "unauthenticated" unless you have the secret uuid,
		// we don't want to give randoms any more info if they are prying.
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			router.ErrorResponse{Error: "ingest failed"})
		return
	}
	c.Status(http.StatusOK)
}
