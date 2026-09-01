package imprt

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/feature/auth/authmiddleware"
	"github.com/sbondCo/Watcharr/router"
)

type Router struct {
	br           *router.BaseRouter
	service      *Service
	traktService *TraktService
}

func NewRouter(br *router.BaseRouter, service *Service, traktService *TraktService) *Router {
	return &Router{
		br,
		service,
		traktService,
	}
}

func (r *Router) AddRoutes() {
	imprt := r.br.Router.Group("/import").Use(authmiddleware.AuthRequired(nil, r.br.Cfg))

	imprt.POST("", r.ImportContent)
	imprt.POST("/trakt", r.ImportTrakt)

	// Saved choices of what a name from an import file refers to.
	imprt.GET("/mappings", r.GetImportMappings)
	imprt.PUT("/mappings/:id", r.UpdateImportMapping)
	imprt.DELETE("/mappings", r.DeleteAllImportMappings)
	imprt.DELETE("/mappings/:id", r.DeleteImportMapping)
}

// Get all of our saved import mappings.
func (r *Router) GetImportMappings(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	mappings, err := r.service.GetImportMappings(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, mappings)
}

// Point a saved import mapping at different content.
func (r *Router) UpdateImportMapping(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	userId := c.MustGet("userId").(uint)
	var ur domain.ImportMappingUpdateRequest
	if err := c.ShouldBindJSON(&ur); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	mapping, err := r.service.UpdateImportMapping(userId, uint(id), ur)
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, mapping)
}

// Forget all of our saved import mappings.
func (r *Router) DeleteAllImportMappings(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	numDeleted, err := r.service.DeleteAllImportMappings(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"numDeleted": numDeleted})
}

// Forget a saved import mapping.
func (r *Router) DeleteImportMapping(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	userId := c.MustGet("userId").(uint)
	if err := r.service.DeleteImportMapping(userId, uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

// Import content (the client handle processing data and sends it to us in a uniform way).
func (r *Router) ImportContent(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	var ar domain.ImportRequest
	err := c.ShouldBindJSON(&ar)
	if err == nil {
		response, err := r.service.ImportContent(userId, ar)
		if err != nil {
			c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
		return
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
}

// Import Trakt.
func (r *Router) ImportTrakt(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	var ar TraktImportRequest
	err := c.ShouldBindJSON(&ar)
	if err == nil {
		response, err := r.traktService.TraktImportWatched(userId, ar)
		if err != nil {
			c.JSON(http.StatusForbidden, router.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
		return
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
}
