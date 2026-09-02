package task

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/feature/auth/authmiddleware"
	"github.com/sbondCo/Watcharr/feature/user/usermiddleware"
	"github.com/sbondCo/Watcharr/router"
	"github.com/sbondCo/Watcharr/task"
)

type Router struct {
	br *router.BaseRouter
}

func NewRouter(br *router.BaseRouter) *Router {
	return &Router{br: br}
}

func (r *Router) AddRoutes() {
	task := r.br.Router.Group("/task").
		Use(
			authmiddleware.AuthRequired(r.br.Cfg),
			usermiddleware.WithUser(r.br.DB),
			authmiddleware.AdminRequired(),
		)

	task.GET("/", r.GetAllTasks)
	task.PUT(":name", r.UpdateTaskSchedule)
}

func (r *Router) GetAllTasks(c *gin.Context) {
	response := task.GetAllTasks(r.br.Cfg)
	c.JSON(http.StatusOK, response)
}

func (r *Router) UpdateTaskSchedule(c *gin.Context) {
	if c.Param("name") == "" {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "no task name provided"})
		return
	}
	var rr task.TaskRescheduleRequest
	err := c.ShouldBindJSON(&rr)
	if err == nil {
		err := task.RescheduleTask(r.br.Cfg, c.Param("name"), rr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: err.Error()})
			return
		}
		c.Status(http.StatusOK)
		return
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
}
