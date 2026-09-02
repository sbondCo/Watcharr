package user

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/feature/auth/authmiddleware"
	"github.com/sbondCo/Watcharr/router"
)

type Router struct {
	br            *router.BaseRouter
	service       *Service
	manageService *ManageService
}

func NewRouter(br *router.BaseRouter, service *Service, manageService *ManageService) *Router {
	return &Router{
		br,
		service,
		manageService,
	}
}

func (r *Router) AddRoutes() {
	u := r.br.Router.Group("/user").Use(authmiddleware.AuthRequired(r.br.Cfg))

	// Get current user info
	u.GET("", r.GetUserInfo)
	// Update current user settings
	u.POST("/update", r.UpdateSettings)
	// Get current user setting
	u.GET("/settings", r.GetSettings)
	// Search users
	u.GET("/search", r.GetSearchUsers)
	// Get user public info
	u.GET("/public/:pubUserId/:pubUsername", r.GetUserPublicInfo)
	// Update bio
	u.POST("/bio", r.UpdateBio)
	// Upload avatar
	u.POST("/avatar", r.UpdateAvatar)
}

// Get current user info
func (r *Router) GetUserInfo(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	response, err := r.service.GetUserInfo(userId)
	if err != nil {
		c.JSON(http.StatusForbidden, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

// Update current user settings
func (r *Router) UpdateSettings(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	var ur entity.UserSettings
	err := c.ShouldBindJSON(&ur)
	if err == nil {
		response, err := r.service.UserUpdate(userId, ur)
		if err != nil {
			c.JSON(http.StatusForbidden, router.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
		return
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
}

// Get current user setting
func (r *Router) GetSettings(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	response, err := r.service.UserGetSettings(userId)
	if err != nil {
		c.JSON(http.StatusForbidden, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

// Search users
func (r *Router) GetSearchUsers(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "a query was not provided"})
		return
	}
	response, err := r.service.UserSearch(userId, query)
	if err != nil {
		c.JSON(http.StatusForbidden, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

// Get user public info
func (r *Router) GetUserPublicInfo(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("pubUserId"))
	if err != nil {
		c.Status(400)
		return
	}
	response, err := r.service.GetUserPublicInfo(uint(id), c.Param("pubUsername"))
	if err != nil {
		c.JSON(http.StatusForbidden, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

// Update bio
func (r *Router) UpdateBio(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	var br domain.UserBioUpdateRequest
	err := c.ShouldBindJSON(&br)
	if err == nil {
		err := r.service.UserUpdateBio(userId, br.NewBio)
		if err != nil {
			c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: err.Error()})
			return
		}
		c.Status(http.StatusOK)
		return
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
}

// Upload avatar
func (r *Router) UpdateAvatar(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	response, err := r.service.UploadUserAvatar(c, userId)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}
