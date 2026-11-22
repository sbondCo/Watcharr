// auth_proxy contains the logic for Trusted Header Authentication/SSO.
//
// This code is inherently dangerous since we are implicitly trusting
// a header for auth, so this should only be configured if you are
// certain your watcharr instance is only available behind your proxy.

package auth

import (
	"errors"
	"log/slog"

	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/database/entity"
	"gorm.io/gorm"
)

type TrustedHeaderAuthLogoutDetailsResponse struct {
	LogoutUrl string `json:"logoutUrl,omitempty"`
}

type TrustedHeaderService struct {
	db          *gorm.DB
	cfg         *config.ServerConfig
	authService *Service
}

func NewTrustedHeaderService(db *gorm.DB, cfg *config.ServerConfig, authService *Service) *TrustedHeaderService {
	return &TrustedHeaderService{
		db,
		cfg,
		authService,
	}
}

// Is trusted header auth configured on this server?
func (s *TrustedHeaderService) TrustedHeaderAuthIsEnabled() bool {
	return s.cfg.HEADER_AUTH.Enabled && s.cfg.HEADER_AUTH.HeaderName != ""
}

func (s *TrustedHeaderService) SetTrustedHeaderAuthSetting(has config.TrustedHeaderAuthSetting) error {
	slog.Debug("setTrustedHeaderAuthSetting: Attempting to update to new provided value", "new_value", has)
	s.cfg.HEADER_AUTH = has
	err := s.cfg.Write()
	if err != nil {
		slog.Error("setTrustedHeaderAuthSetting: Failed to write updated config!", "error", err)
		return errors.New("failed to write config")
	}
	return nil
}

// Gets proxy logout details.
// Details are accessible to any user for the logout flow.
// If proxy configured should be checked before using this.
func (s *TrustedHeaderService) GetTrustedHeaderAuthLogoutDetails() *TrustedHeaderAuthLogoutDetailsResponse {
	return &TrustedHeaderAuthLogoutDetailsResponse{
		LogoutUrl: s.cfg.HEADER_AUTH.LogoutUrl,
	}
}

// Login via header sso
func (s *TrustedHeaderService) LoginTrustedHeaderAuth(user *entity.User) (AuthResponse, error) {
	slog.Debug("loginTrustedHeaderAuth: A user is logging in", "username_from_header", user.Username)
	dbUser := new(entity.User)
	res := s.db.Where("username = ? AND type = ?", user.Username, entity.PROXY_USER).Take(&dbUser)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			slog.Info("loginTrustedHeaderAuth: Creating new User from authentication header", "username_from_header", user.Username)
			// Record not found, so we should create the user (if configured to do so)
			// dbUser will be empty, so we can just reuse it for this purpose.
			dbUser.Username = user.Username
			dbUser.Type = entity.PROXY_USER
			dbUser.Country = &s.cfg.DEFAULT_COUNTRY

			res = s.db.Create(&dbUser)
			if res.Error != nil {
				slog.Error("loginTrustedHeaderAuth: Failed to create new user in db", "error", res.Error)
				return AuthResponse{}, errors.New("failed to create new user")
			}
		} else {
			slog.Error("loginTrustedHeaderAuth: An error occurred when looking up user in db", "error", res.Error)
			return AuthResponse{}, errors.New("error locating user in db")
		}
	}
	token, err := s.authService.signJWT(dbUser.ID, dbUser.Username, entity.PROXY_USER)
	if err != nil {
		slog.Error("loginTrustedHeaderAuth: Failed to sign new jwt", "error", err)
		return AuthResponse{}, errors.New("failed to get auth token")
	}
	return AuthResponse{Token: token}, nil
}
