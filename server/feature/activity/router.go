package activity

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/activity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/feature/auth/authmiddleware"
	"github.com/sbondCo/Watcharr/router"
)

type Router struct {
	br *router.BaseRouter
}

func NewRouter(br *router.BaseRouter) *Router {
	return &Router{
		br,
	}
}

func (r *Router) AddRoutes() {
	activity := r.br.Router.
		Group("/activity").
		Use(authmiddleware.AuthRequired(r.br.Cfg))

	activity.GET(":watchedId", r.GetActivity)
	activity.PUT(":id", r.UpdateActivity)
	activity.DELETE(":id", r.DeleteActivity)
}

func (r *Router) GetActivity(c *gin.Context) {
	watchedId, err := strconv.ParseUint(c.Param("watchedId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "check watched id route param"})
		return
	}
	userId := c.MustGet("userId").(uint)
	activity, err := activity.Get(r.br.DB, userId, uint(watchedId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, activity)
}

func (r *Router) UpdateActivity(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.Status(400)
		return
	}
	var activityUpdateRequest domain.ActivityUpdateRequest
	err = c.ShouldBindJSON(&activityUpdateRequest)
	if err == nil {
		err = activity.Update(r.br.DB, userId, uint(id), activityUpdateRequest)
		if err != nil {
			c.JSON(http.StatusForbidden, router.ErrorResponse{Error: err.Error()})
			return
		}
		c.Status(http.StatusOK)
		return
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
}

func (r *Router) DeleteActivity(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.Status(400)
		slog.Error("DeleteActivity: Couldn't parse id.",
			"error", err.Error(), "id", c.Param("id"))
		return
	}
	err = activity.Delete(r.br.DB, userId, uint(id))
	if err != nil {
		c.JSON(http.StatusForbidden, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.Status(http.StatusOK)
}
