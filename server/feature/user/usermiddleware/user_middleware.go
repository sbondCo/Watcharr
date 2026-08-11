// Middleware relating to user data.
package usermiddleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/database/entity"
	"gorm.io/gorm"
)

const (
	reqContextKeyUser        = "user"
	reqContextKeyUserService = "userService"
)

// Fields we want for the User object when stored in request context.
// We don't need all the fields, especially not any sensitive ones.
// If a handler needs sensitive fields, it should get them itself, but also
// consider why a handler would even need anything sensitive.
type UserForMiddleware struct {
	Username    string
	Type        entity.UserType
	Permissions int
}

// Attach user to context.
// Must be ran after AuthRequired middleware.
func WithUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userId").(uint)
		user := new(UserForMiddleware)
		err := db.
			Model(&entity.User{}).
			Where("id = ?", userID).
			Take(&user).
			Error
		if err != nil {
			slog.Error("WithUser: Query failed!", "error", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.Set(reqContextKeyUser, user)
		c.Next()
	}
}

// Get user from request context.
func UserFromContext(c *gin.Context) *UserForMiddleware {
	return c.MustGet(reqContextKeyUser).(*UserForMiddleware)
}

// Attach user_service (SINGULAR!) of `name` to context.
// Context value key is `reqContextKeyUserService+name`.
// Only for use with services where only one by `name` is allowed/supported..
// if you need to get a slice of user_services, use another middleware!
func WithUserServiceByName(
	db *gorm.DB,
	name entity.UserServiceName,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userId").(uint)
		userService := new(entity.UserServices)
		err := db.
			Model(&entity.UserServices{}).
			Where("user_id = ? AND name = ?", userID, name).
			Take(&userService).
			Error
		if err != nil {
			slog.Error("WithUserServiceByName: Query failed!", "error", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.Set(reqContextKeyUserService+name, userService)
		c.Next()
	}
}

// Get user service (SINGULAR!) by name from request context.
func UserServiceFromContext(
	c *gin.Context,
	name entity.UserServiceName,
) *entity.UserServices {
	return c.MustGet(reqContextKeyUserService + name).(*entity.UserServices)
}
