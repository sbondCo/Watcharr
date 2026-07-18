package authmiddleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/feature/auth/permission"
	"gorm.io/gorm"
)

// Auth middleware
// If db is passed, extra user info from the database will be fetched.
//
// **NOTE:** Instead of providing the `db` parameter, it is probably better to
// fetch what you need in the handler directly! We might follow that pattern
// from now on and potentially remove `db` from this func in the future.
func AuthRequired(db *gorm.DB, cfg *config.ServerConfig) gin.HandlerFunc {
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
			// If db passed, get extra user info and set as variables in req context
			if db != nil {
				slog.Debug("AuthRequired: db passed.. getting extra user info")
				dbUser := new(entity.User)
				res := db.Where("id = ?", claims.UserID).Take(&dbUser)
				if res.Error != nil {
					slog.Error("AuthRequired: Failed to select user from database", "error", res.Error)
					c.AbortWithStatus(401)
					return
				}
				slog.Debug("AuthRequired: fetched extra user info. Setting vars.", "userThirdPartyId", dbUser.ThirdPartyID, "userThirdPartyAuth", "lol this is censored dude")
				c.Set("userThirdPartyId", dbUser.ThirdPartyID)
				c.Set("userThirdPartyAuth", dbUser.ThirdPartyAuth)
				c.Set("username", dbUser.Username)
				c.Set("userPermissions", dbUser.Permissions)
				if dbUser.Country != nil {
					c.Set("userCountry", *dbUser.Country)
				}
			}
			c.Next()
		} else {
			slog.Error("Token is **not** valid")
			c.AbortWithStatus(401)
			return
		}
	}
}

// Admin only middleware (use after AuthRequired with extra info!)
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.GetUint("userId")
		perms := c.GetInt("userPermissions")
		if permission.Has(perms, entity.PERM_ADMIN) {
			slog.Debug("AdminRequired: User has permission to access admin only route", "user_id", userId)
			c.Next()
			return
		}
		slog.Info("AdminRequired: User denied permission to access admin only route", "user_id", userId)
		c.AbortWithStatus(401)
	}
}

// Specific perm only middleware (use after AuthRequired with extra info!)
func PermRequired(perm int) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.GetUint("userId")
		perms := c.GetInt("userPermissions")
		if permission.Has(perms, perm) {
			slog.Debug("PermRequired: User has permission to access perm only route", "user_id", userId, "required_perm", perm)
			c.Next()
			return
		}
		slog.Info("PermRequired: User denied permission to access perm only route", "user_id", userId, "required_perm", perm)
		c.AbortWithStatus(401)
	}
}
