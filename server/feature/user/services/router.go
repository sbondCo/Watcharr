package services

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/domain"
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
	us := r.br.Router.Group("/user/services").
		Use(authmiddleware.AuthRequired(r.br.Cfg))

	// Get all current users connected services.
	us.GET("", r.GetAll)
	us.PATCH(":id/settings", r.PatchSettings)
}

// Get all current users connected services.
func (r *Router) GetAll(c *gin.Context) {
	userID := c.MustGet("userId").(uint)

	services, err := r.service.GetAll(userID)
	if err != nil {
		c.JSON(http.StatusForbidden, router.ErrorResponse{Error: err.Error()})
		return
	}

	response := domain.NewAllUserServicesResponse(services)
	c.JSON(http.StatusOK, response)
}

// Update service (by id) settings.
func (r *Router) PatchSettings(c *gin.Context) {
	userID := c.MustGet("userId").(uint)

	serviceID, err := strconv.ParseUint(c.Param("id"), 10, 0)
	if err != nil || serviceID == 0 {
		slog.Error("PatchSettings: Invalid serviceID.",
			"id", serviceID, "error", err)
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "id not provided"})
		return
	}

	var req domain.UserServicePatchSettingsDto
	err = c.ShouldBindJSON(&req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}

	resp, err := r.service.Update(uint(serviceID), userID, req)
	if err != nil {
		c.JSON(http.StatusForbidden, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
