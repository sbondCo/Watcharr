package authmiddleware

import (
	"errors"
	"log/slog"
	"net/http"
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

		if err := Authenticate(c, db, cfg); err != nil {
			slog.Warn("Authentication failed", "error", err)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Next()
	}
}

// used for a few routes which can be accessed by guests if ALLOW_GUESTS server setting is enabled
func AuthOptional(db *gorm.DB, cfg *config.ServerConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := Authenticate(c, db, cfg)

		slog.Info("headers",
			"authorization", c.GetHeader("Authorization"),
		)

		if err != nil {
			if cfg.ALLOW_GUESTS {
				c.Next()
				return
			}

			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Next()
	}
}


func Authenticate(c *gin.Context, db *gorm.DB, cfg *config.ServerConfig) error {
	atoken := c.GetHeader("Authorization")
	if atoken == "" {
		return errors.New("authorization header not provided")
	}

	token, err := jwt.ParseWithClaims(
		atoken,
		&entity.TokenClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWT_SECRET), nil
		},
	)
	if err != nil {
		return err
	}

	claims, ok := token.Claims.(*entity.TokenClaims)
	if !ok || !token.Valid {
		return errors.New("invalid token")
	}

	timeOfNewLoginRequired, _ :=
		time.Parse(time.RFC822, "18 Aug 23 20:30 UTC")

	if claims.IssuedAt.Before(timeOfNewLoginRequired) {
		return errors.New("token is too old")
	}

	c.Set("userId", claims.UserID)
	c.Set("userType", claims.Type)

	if db != nil {
		dbUser := new(entity.User)
		res := db.Where("id = ?", claims.UserID).Take(dbUser)
		if res.Error != nil {
			return res.Error
		}

		c.Set("userThirdPartyId", dbUser.ThirdPartyID)
		c.Set("userThirdPartyAuth", dbUser.ThirdPartyAuth)
		c.Set("username", dbUser.Username)
		c.Set("userPermissions", dbUser.Permissions)

		if dbUser.Country != nil {
			c.Set("userCountry", *dbUser.Country)
		}
	}

	return nil
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
