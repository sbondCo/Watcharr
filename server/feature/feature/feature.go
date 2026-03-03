package feature

import (
	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/feature/auth/permission"
)

type ServerFeatures struct {
	Sonarr bool `json:"sonarr"`
	Radarr bool `json:"radarr"`
	Games  bool `json:"games"`
}

type Service struct {
	cfg *config.ServerConfig
}

func NewService(cfg *config.ServerConfig) *Service {
	return &Service{
		cfg,
	}
}

// Get enabled server functionality from Config.
// Mainly so the frontend can store this once and know
// which btns should be shown, etc.
func (s *Service) GetEnabledFeatures(userPerms int) ServerFeatures {
	var f ServerFeatures
	if s.cfg.TwitchEnabled() {
		f.Games = true
	}
	if permission.Has(userPerms, entity.PERM_REQUEST_CONTENT) {
		if len(s.cfg.SONARR) > 0 {
			f.Sonarr = true
		}
		if len(s.cfg.RADARR) > 0 {
			f.Radarr = true
		}
	}
	return f
}
