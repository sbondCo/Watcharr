package auth

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/feature/auth/authmiddleware"
	"github.com/sbondCo/Watcharr/feature/plex"
	"github.com/sbondCo/Watcharr/feature/setup/setupglob"
	"github.com/sbondCo/Watcharr/router"
	"github.com/sbondCo/Watcharr/token"
)

type Router struct {
	br                   *router.BaseRouter
	service              *Service
	trustedHeaderService *TrustedHeaderService
}

func NewRouter(br *router.BaseRouter, service *Service, trustedHeaderService *TrustedHeaderService) *Router {
	return &Router{
		br,
		service,
		trustedHeaderService,
	}
}

func (r *Router) AddRoutes() {
	auth := r.br.Router.Group("/auth")

	// Login
	auth.POST("/", r.Login)
	// Jellyfin login
	auth.POST("/jellyfin", r.LoginJellyfin)
	// Plex login
	auth.POST("/plex", r.LoginPlex)
	// Proxy Login
	auth.POST("/proxy", r.LoginProxy)
	// Register
	auth.POST("/register", r.Register)
	// Get available auth providers
	auth.GET("/available", r.GetAvailableAuthProviders)

	// IMPORTANT: Routes below here must be authenticated.
	auth.Use(authmiddleware.AuthRequired(r.br.Cfg))
	{
		// Request details for logout process for proxy users.
		// Any proxy user can request this for logout.
		auth.GET("/proxy_logout_details", r.GetProxyLogoutDetails)
		// Request admin token
		auth.GET("/admin_token", r.GetAdminToken)
		// Use admin token
		auth.POST("/admin_token", r.UseAdminToken)
		// Change password
		auth.POST("/change_password", r.UpdateUserPassword)
	}
}

// Login
func (r *Router) Login(c *gin.Context) {
	var user entity.User
	if c.ShouldBindJSON(&user) == nil {
		response, err := r.service.Login(&user)
		if err != nil {
			c.JSON(http.StatusForbidden, router.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
		return
	}
	c.Status(400)
}

// Jellyfin login
func (r *Router) LoginJellyfin(c *gin.Context) {
	var user entity.User
	if c.ShouldBindJSON(&user) == nil {
		response, err := r.service.LoginJellyfin(&user)
		if err != nil {
			c.JSON(http.StatusForbidden, router.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
		return
	}
	c.Status(400)
}

// Plex login
func (r *Router) LoginPlex(c *gin.Context) {
	var plexRequest plex.PlexLoginRequest
	if c.ShouldBindJSON(&plexRequest) == nil {
		response, err := r.service.LoginPlex(&plexRequest)
		if err != nil {
			c.JSON(http.StatusForbidden, router.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
		return
	}
	c.Status(400)
}

// Proxy Login
func (r *Router) LoginProxy(c *gin.Context) {
	var user entity.User
	if !r.trustedHeaderService.TrustedHeaderAuthIsEnabled() {
		slog.Error("ProxyLogin: SSO has not been configured.")
		c.JSON(http.StatusForbidden, router.ErrorResponse{Error: "proxy authentication is disabled"})
		return
	}
	user.Username = c.GetHeader(r.br.Cfg.HEADER_AUTH.HeaderName)
	if user.Username == "" {
		slog.Error("ProxyLogin: Authentication header is missing.")
		c.JSON(http.StatusForbidden, router.ErrorResponse{Error: "authentication header missing"})
		return
	}
	response, err := r.trustedHeaderService.LoginTrustedHeaderAuth(&user)
	if err != nil {
		c.JSON(http.StatusForbidden, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

// Register
func (r *Router) Register(c *gin.Context) {
	var user UserRegisterRequest
	if c.ShouldBindJSON(&user) == nil {
		response, err := r.service.Register(&user, entity.PERM_NONE)
		if err != nil {
			c.JSON(http.StatusForbidden, router.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
		return
	}
	c.Status(400)
}

// Get available auth providers
func (r *Router) GetAvailableAuthProviders(c *gin.Context) {
	resp := &AvailableAuthProvidersResponse{
		AvailableAuthProviders: []string{},
		SignupEnabled:          r.br.Cfg.SIGNUP_ENABLED,
		IsInSetup:              setupglob.ServerInSetup,
		UseEmby:                r.br.Cfg.USE_EMBY,
	}
	if r.br.Cfg.JELLYFIN_HOST != "" {
		resp.AvailableAuthProviders = append(resp.AvailableAuthProviders, "jellyfin")
	}
	if r.br.Cfg.PLEX_HOST != "" && r.br.Cfg.PLEX_MACHINE_ID != "" {
		resp.AvailableAuthProviders = append(resp.AvailableAuthProviders, "plex")
	}
	if r.trustedHeaderService.TrustedHeaderAuthIsEnabled() {
		resp.AvailableAuthProviders = append(resp.AvailableAuthProviders, "header")
		resp.HeaderAuthAutoLogin = r.br.Cfg.HEADER_AUTH.AutoLogin
	}
	c.JSON(http.StatusOK, resp)
}

// Request details for logout process for proxy users.
// Any proxy user can request this for logout.
func (r *Router) GetProxyLogoutDetails(c *gin.Context) {
	if !r.trustedHeaderService.TrustedHeaderAuthIsEnabled() {
		slog.Error("GetProxy: SSO has not been configured.")
		c.JSON(http.StatusForbidden, router.ErrorResponse{Error: "proxy authentication is disabled"})
		return
	}
	userType := c.MustGet("userType").(entity.UserType)
	if userType != entity.PROXY_USER {
		slog.Error("GetProxy: Non proxy user attempted to fetch proxy logout details.")
		c.JSON(http.StatusForbidden, router.ErrorResponse{Error: "you are not a proxy user"})
		return
	}
	c.JSON(http.StatusOK, r.trustedHeaderService.GetTrustedHeaderAuthLogoutDetails())
}

// Request admin token
func (r *Router) GetAdminToken(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	token, err := token.CreateOneUseToken(r.br.DB, entity.TOKENTYPE_ADMIN, userId)
	if err != nil {
		c.JSON(http.StatusForbidden, router.ErrorResponse{Error: err.Error()})
		return
	}
	slog.Info("Admin token generated. Type this token into the web ui to gain admin access on your account.", "token", token, "generated_for", userId)
	c.Status(http.StatusNoContent)
}

// Use admin token
func (r *Router) UseAdminToken(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	var atr UseAdminTokenRequest
	if c.ShouldBindJSON(&atr) == nil {
		err := r.service.UseAdminToken(&atr, userId)
		if err != nil {
			c.JSON(http.StatusForbidden, router.ErrorResponse{Error: err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
		return
	}
	c.Status(400)
}

// Change password
func (r *Router) UpdateUserPassword(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	var pwds UserPasswordUpdateRequest
	err := c.ShouldBindJSON(&pwds)
	if err == nil {
		err := r.service.UserChangePassword(pwds, userId)
		if err != nil {
			c.JSON(http.StatusForbidden, router.ErrorResponse{Error: err.Error()})
			return
		}
		c.Status(http.StatusOK)
		return
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
}
