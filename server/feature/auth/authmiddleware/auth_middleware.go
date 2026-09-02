package authmiddleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/feature/auth/permission"
	"github.com/sbondCo/Watcharr/feature/user/usermiddleware"
)

// Auth middleware
func AuthRequired(cfg *config.ServerConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		slog.Debug("AuthRequired middleware hit")
		atoken := c.GetHeader("Authorization")
		// Make sure auth header isn't empty
		if atoken == "" {
			slog.Warn("Returning 401, Authorization header not provided")
			c.AbortWithStatus(401)
			return
		}
		// Parse token
		token, err := jwt.ParseWithClaims(atoken, &entity.TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWT_SECRET), nil
		})
		if err != nil {
			slog.Error("AuthRequired failed to parse token", "error", err)
			c.AbortWithStatus(401)
			return
		}
		// If token is valid, go to next handler
		if claims, ok := token.Claims.(*entity.TokenClaims); ok && token.Valid {
			// Check if token issuedAt is from before `timeOfNewLoginRequired`.
			// Basically just so we can logout old tokens and force relogin...
			// since new changes require the user login again.
			timeOfNewLoginRequired, _ := time.Parse(time.RFC822, "18 Aug 23 20:30 UTC")
			if claims.IssuedAt.Before(timeOfNewLoginRequired) {
				slog.Info("Token is from before timeOfNewLoginRequired.. returning 401", "token_issued_at", claims.IssuedAt, "time_of_new_login_required", timeOfNewLoginRequired)
				c.AbortWithStatus(401)
				return
			}
			slog.Debug("Token is valid", "claims", claims)
			c.Set("userId", claims.UserID)
			c.Set("userType", claims.Type)
			c.Next()
		} else {
			slog.Error("Token is **not** valid")
			c.AbortWithStatus(401)
			return
		}
	}
}

// Admin only middleware (only use after AuthRequired & usermiddleware.WithUser!)
func AdminRequired() gin.HandlerFunc {
	return PermRequired(entity.PERM_ADMIN)
}

// Specific perm only middleware (only use after AuthRequired & usermiddleware.WithUser!)
func PermRequired(perm int) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.GetUint("userId")
		user := usermiddleware.UserFromContext(c)
		if permission.Has(user.Permissions, perm) {
			slog.Debug("PermRequired: User has permission!",
				"user_id", userId, "required_perm", perm)
			c.Next()
			return
		}
		slog.Info("PermRequired: User denied permission!",
			"user_id", userId, "required_perm", perm)
		c.AbortWithStatus(401)
	}
}
