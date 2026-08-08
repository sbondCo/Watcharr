package authmiddleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/feature/user/usermiddleware"
)

// Jellyfin access middleware, ensures user is a jellyfin user.
// To be ran after AuthRequired & WithUserServiceByName middleware.
func JellyfinAccessRequired(cfg *config.ServerConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.MustGet("userId").(uint)
		userJellyfinService := usermiddleware.UserServiceFromContext(
			c, entity.UserServiceNameJellyfin)
		slog.Debug("JellyfinAccessRequired middleware hit", "user_id", userId)
		userType := c.MustGet("userType").(entity.UserType)
		if cfg.JELLYFIN_HOST == "" {
			slog.Error("JellyfinAccessRequired: JELLYFIN_HOST has not been configured.")
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		if userType != entity.JELLYFIN_USER || userJellyfinService.ClientID == "" {
			slog.Error("JellyfinAccessRequired: User is not a jellyfin user..",
				"user_type", userType,
				"user_third_party_id", userJellyfinService.ClientID)
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		if userJellyfinService.AuthToken == "" {
			slog.Error("JellyfinAccessRequired: User has no thirdPartyAuth token..")
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
	}
}
