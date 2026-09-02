package arr

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/arr"
	"github.com/sbondCo/Watcharr/config/cfgmodel"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/feature/auth/authmiddleware"
	"github.com/sbondCo/Watcharr/feature/user/usermiddleware"
	"github.com/sbondCo/Watcharr/router"
)

// ContentProvider - Temporary, this ARR code at some point will
// be turned into services to conform with new code format, for now
// passing contentprovider through here.
type ContentProvider interface {
	GetOrCacheContent(contentType entity.ContentType, tmdbId int) (entity.Content, error)
}

type Router struct {
	br              *router.BaseRouter
	contentProvider ContentProvider
}

func NewRouter(br *router.BaseRouter, contentProvider ContentProvider) *Router {
	return &Router{
		br,
		contentProvider,
	}
}

func (r *Router) AddRoutes() {
	// **NOTE:** Routes are manually given authmiddleware.AdminRequired or authmiddleware.PermRequired middleware.

	// SONARR
	{
		s := r.br.Router.Group("/arr/son").
			Use(
				authmiddleware.AuthRequired(r.br.Cfg),
				usermiddleware.WithUser(r.br.DB),
			)

		// Routes are manually given authmiddleware.AdminRequired or authmiddleware.PermRequired middleware.

		// Test configuration
		s.POST("/test", authmiddleware.AdminRequired(), r.TestSonarr)
		// Used to get config for specific server (quality profile, root folder, etc)
		s.GET("/config/:name", authmiddleware.PermRequired(entity.PERM_REQUEST_CONTENT), r.GetSonarrServer)
		// Add sonarr server into config
		s.POST("/add", authmiddleware.AdminRequired(), r.AddSonarr)
		// Edit sonarr servers config
		s.POST("/edit", authmiddleware.AdminRequired(), r.UpdateSonarrServer)
		// Remove sonarr server
		s.POST("/rm/:name", authmiddleware.AdminRequired(), r.UpdateRemoveSonarrServer)
		// Get safe config for all sonarr servers
		s.GET("", authmiddleware.PermRequired(entity.PERM_REQUEST_CONTENT), r.GetSonarrsSafe)
		// Request a show
		s.POST("/request", authmiddleware.PermRequired(entity.PERM_REQUEST_CONTENT), r.CreateSonarrRequest)
		s.GET("/request/:tmdbId", authmiddleware.PermRequired(entity.PERM_REQUEST_CONTENT), r.GetSonarrRequestByTmdbId)
		s.POST("/request/approve/:id", authmiddleware.PermRequired(entity.PERM_ADMIN), r.UpdateApproveSonarrRequest)
		s.GET("/status/:serverName/:arrId", authmiddleware.PermRequired(entity.PERM_REQUEST_CONTENT), r.GetSonarrQueueDetails)
		s.GET("/info/:requestId", authmiddleware.PermRequired(entity.PERM_REQUEST_CONTENT), r.GetSonarrRequestInfo)
	}

	// RADARR
	{
		s := r.br.Router.Group("/arr/rad").
			Use(
				authmiddleware.AuthRequired(r.br.Cfg),
				usermiddleware.WithUser(r.br.DB),
			)

		// Routes are manually given authmiddleware.AdminRequired or authmiddleware.PermRequired middleware.

		// Test configuration
		s.POST("/test", authmiddleware.AdminRequired(), r.TestRadarr)
		// Get config for specific server
		s.GET("/config/:name", authmiddleware.PermRequired(entity.PERM_REQUEST_CONTENT), r.GetRadarrServer)
		s.POST("/add", authmiddleware.AdminRequired(), r.AddRadarr)
		s.POST("/edit", authmiddleware.AdminRequired(), r.UpdateRadarrServer)
		s.POST("/rm/:name", authmiddleware.AdminRequired(), r.UpdateRemoveRadarrServer)
		s.GET("", authmiddleware.PermRequired(entity.PERM_REQUEST_CONTENT), r.GetRadarrsSafe)
		s.POST("/request", authmiddleware.PermRequired(entity.PERM_REQUEST_CONTENT), r.CreateRadarrRequest)
		s.GET("/request/:tmdbId", authmiddleware.PermRequired(entity.PERM_REQUEST_CONTENT), r.GetRadarrRequestByTmdbId)
		s.POST("/request/approve/:id", authmiddleware.PermRequired(entity.PERM_ADMIN), r.UpdateApproveRadarrRequest)
		s.GET("/status/:serverName/:arrId", authmiddleware.PermRequired(entity.PERM_REQUEST_CONTENT), r.GetRadarrQueueDetails)
		s.GET("/info/:requestId", authmiddleware.PermRequired(entity.PERM_REQUEST_CONTENT), r.GetRadarrRequestInfo)
	}

	// Request Management
	{
		s := r.br.Router.Group("/arr/request").
			Use(
				authmiddleware.AuthRequired(r.br.Cfg),
				usermiddleware.WithUser(r.br.DB),
			)

		// Get all requests (for manage_requests view), only for admins.
		s.GET("/", authmiddleware.AdminRequired(), r.GetAllRequests)
		// Deny a request (for manage_requests view), only for admins.
		s.POST("/deny/:id", authmiddleware.AdminRequired(), r.UpdateDenyRequest)
	}
}

// Test configuration
func (r *Router) TestSonarr(c *gin.Context) {
	var ur ArrTestParams
	err := c.ShouldBindJSON(&ur)
	if err == nil {
		resp, err := testSonarr(ur)
		if err != nil {
			c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
		return
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
}

// Used to get config for specific server (quality profile, root folder, etc)
func (r *Router) GetSonarrServer(c *gin.Context) {
	server, err := getSonarr(r.br.Cfg, c.Param("name"))
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	resp, err := testSonarr(ArrTestParams{Host: server.Host, Key: server.Key})
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// Add sonarr server into config
func (r *Router) AddSonarr(c *gin.Context) {
	var ur cfgmodel.SonarrSettings
	err := c.ShouldBindJSON(&ur)
	if err == nil {
		err := addSonarr(r.br.Cfg, ur)
		if err != nil {
			c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
			return
		}
		c.Status(http.StatusOK)
		return
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
}

// Edit sonarr servers config
func (r *Router) UpdateSonarrServer(c *gin.Context) {
	var ur cfgmodel.SonarrSettings
	err := c.ShouldBindJSON(&ur)
	if err == nil {
		err := editSonarr(r.br.Cfg, ur)
		if err != nil {
			c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
			return
		}
		c.Status(http.StatusOK)
		return
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
}

// Remove sonarr server
func (r *Router) UpdateRemoveSonarrServer(c *gin.Context) {
	err := rmSonarr(r.br.Cfg, c.Param("name"))
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

// Get safe config for all sonarr servers
func (r *Router) GetSonarrsSafe(c *gin.Context) {
	response := getSonarrsSafe(r.br.Cfg)
	c.JSON(http.StatusOK, response)
}

// Request a show
func (r *Router) CreateSonarrRequest(c *gin.Context) {
	var ur arr.SonarrRequest
	err := c.ShouldBindJSON(&ur)
	if err == nil {
		userId := c.MustGet("userId").(uint)
		user := usermiddleware.UserFromContext(c)
		response, err := createSonarrRequest(
			r.br.Cfg, r.br.DB, r.contentProvider, userId, user.Permissions, ur)
		if err != nil {
			c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
		return
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
}

func (r *Router) GetSonarrRequestByTmdbId(c *gin.Context) {
	tmdbId, err := strconv.Atoi(c.Param("tmdbId"))
	if err != nil {
		slog.Error("Couldn't parse tmdbId", "tmdbId", tmdbId)
		c.Status(400)
		return
	}
	response, err := getArrRequestByTmdbId(r.br.DB, entity.SHOW, tmdbId)
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

func (r *Router) UpdateApproveSonarrRequest(c *gin.Context) {
	var ur arr.SonarrRequest
	err := c.ShouldBindJSON(&ur)
	if err == nil {
		requestId, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			slog.Error("Couldn't parse request id", "request_id", requestId)
			c.Status(400)
			return
		}
		response, err := approveSonarrRequest(r.br.Cfg, r.br.DB, uint(requestId), ur)
		if err != nil {
			c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
		return
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
}

func (r *Router) GetSonarrQueueDetails(c *gin.Context) {
	response, err := getSonarrQueueDetails(r.br.Cfg, c.Param("serverName"), c.Param("arrId"))
	if err != nil {
		if err.Error() == "no details found" {
			c.Status(http.StatusNoContent) // Item not found in queue.. missing
			return
		}
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

func (r *Router) GetSonarrRequestInfo(c *gin.Context) {
	requestId, err := strconv.ParseUint(c.Param("requestId"), 10, 64)
	if err != nil {
		slog.Error("/info/:requestId - requestId could not be parsed", "requestId", requestId)
		c.Status(http.StatusBadRequest)
		return
	}
	response, err := getSonarrRequestInfo(r.br.Cfg, r.br.DB, uint(requestId))
	if err != nil {
		if err.Error() == "request deleted" {
			c.JSON(http.StatusNotFound, router.ErrorResponse{Error: "request deleted"})
			return
		}
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

// Test configuration
func (r *Router) TestRadarr(c *gin.Context) {
	var ur ArrTestParams
	err := c.ShouldBindJSON(&ur)
	if err == nil {
		resp, err := testRadarr(ur)
		if err != nil {
			c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
		return
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
}

// Get config for specific server
func (r *Router) GetRadarrServer(c *gin.Context) {
	server, err := getRadarr(r.br.Cfg, c.Param("name"))
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	resp, err := testRadarr(ArrTestParams{Host: server.Host, Key: server.Key})
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (r *Router) AddRadarr(c *gin.Context) {
	var ur cfgmodel.RadarrSettings
	err := c.ShouldBindJSON(&ur)
	if err == nil {
		err := addRadarr(r.br.Cfg, ur)
		if err != nil {
			c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
			return
		}
		c.Status(http.StatusOK)
		return
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
}

func (r *Router) UpdateRadarrServer(c *gin.Context) {
	var ur cfgmodel.RadarrSettings
	err := c.ShouldBindJSON(&ur)
	if err == nil {
		err := editRadarr(r.br.Cfg, ur)
		if err != nil {
			c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
			return
		}
		c.Status(http.StatusOK)
		return
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
}

func (r *Router) UpdateRemoveRadarrServer(c *gin.Context) {
	err := rmRadarr(r.br.Cfg, c.Param("name"))
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (r *Router) GetRadarrsSafe(c *gin.Context) {
	response := getRadarrsSafe(r.br.Cfg)
	c.JSON(http.StatusOK, response)
}

func (r *Router) CreateRadarrRequest(c *gin.Context) {
	var ur arr.RadarrRequest
	err := c.ShouldBindJSON(&ur)
	if err == nil {
		userId := c.MustGet("userId").(uint)
		user := usermiddleware.UserFromContext(c)
		response, err := createRadarrRequest(
			r.br.Cfg, r.br.DB, r.contentProvider, userId, user.Permissions, ur)
		if err != nil {
			c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
		return
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
}

func (r *Router) GetRadarrRequestByTmdbId(c *gin.Context) {
	tmdbId, err := strconv.Atoi(c.Param("tmdbId"))
	if err != nil {
		slog.Error("Couldn't parse tmdbId", "tmdbId", tmdbId)
		c.Status(400)
		return
	}
	response, err := getArrRequestByTmdbId(r.br.DB, entity.MOVIE, tmdbId)
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

func (r *Router) UpdateApproveRadarrRequest(c *gin.Context) {
	var ur arr.RadarrRequest
	err := c.ShouldBindJSON(&ur)
	if err == nil {
		requestId, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			slog.Error("Couldn't parse request id", "request_id", requestId)
			c.Status(400)
			return
		}
		response, err := approveRadarrRequest(r.br.Cfg, r.br.DB, uint(requestId), ur)
		if err != nil {
			c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
		return
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
}

func (r *Router) GetRadarrQueueDetails(c *gin.Context) {
	response, err := getRadarrQueueDetails(r.br.Cfg, c.Param("serverName"), c.Param("arrId"))
	if err != nil {
		if err.Error() == "no details found" {
			c.Status(http.StatusNoContent) // Item not found in queue.. missing
			return
		}
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

func (r *Router) GetRadarrRequestInfo(c *gin.Context) {
	requestId, err := strconv.ParseUint(c.Param("requestId"), 10, 64)
	if err != nil {
		slog.Error("/info/:requestId - requestId could not be parsed", "requestId", requestId)
		c.Status(http.StatusBadRequest)
		return
	}
	response, err := getRadarrRequestInfo(r.br.Cfg, r.br.DB, uint(requestId))
	if err != nil {
		if err.Error() == "request deleted" {
			c.JSON(http.StatusNotFound, router.ErrorResponse{Error: "request deleted"})
			return
		}
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

// Get all requests (for manage_requests view), only for admins.
func (r *Router) GetAllRequests(c *gin.Context) {
	response, err := getArrRequests(r.br.DB)
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

// Deny a request (for manage_requests view), only for admins.
func (r *Router) UpdateDenyRequest(c *gin.Context) {
	requestId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		slog.Error("Couldn't parse request id", "request_id", requestId)
		c.Status(400)
		return
	}
	err = denyArrRequest(r.br.DB, uint(requestId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.Status(http.StatusOK)
}
