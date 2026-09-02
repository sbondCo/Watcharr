package job

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/feature/auth/authmiddleware"
	"github.com/sbondCo/Watcharr/job"
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
	job := r.br.Router.Group("/job").Use(authmiddleware.AuthRequired(r.br.Cfg))

	// Uses wildcard so it still works in cases where the job id includes a /.
	// (yes i changed this instead of not allowing a / when we generate a job id becuz easier)
	job.GET("/*id", r.GetJobById)
}

func (r *Router) GetJobById(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	// When we get id param, don't include first letter, which will be the beginning '/'.
	response, err := job.GetJob(c.Param("id")[1:], userId)
	if err != nil {
		c.JSON(http.StatusForbidden, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, *response)
}
