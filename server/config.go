package main

import (
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"os"

	"github.com/sbondCo/Watcharr/arr"
)

type ServerConfig struct {
	// Used to sign JWT tokens. Make sure to make
	// it strong, just like a very long, complicated password.
	JWT_SECRET string `json:",omitempty"`

	// Optional: Point to your Jellyfin install
	// to enable it as an auth provider.
	JELLYFIN_HOST string `json:",omitempty"`

	// Enable/disable signup functionality.
	// Set to `false` to disable registering an account.
	SIGNUP_ENABLED bool `json:",omitempty"`

	// Optional: Provide your own TMDB API Key.
	// If unprovided, the default Watcharr API key will be used.
	TMDB_KEY string `json:",omitempty"`

	// Optional: Points to Sonarr install.
	SONARR_HOST string `json:",omitempty"`

	// Optional: Sonarr API Key.
	SONARR_KEY string `json:",omitempty"`

	SONARR_QUALITY_PROFILE int `json:",omitempty"`

	// Optional: Points to Radarr install.
	RADARR_HOST string `json:",omitempty"`

	// Optional: Radarr API Key.
	RADARR_KEY string `json:",omitempty"`

	// Enable/disable debug logging. Useful for when trying
	// to figure out exactly what the server is doing at a point
	// of failure.
	// Set to `true` to enable.
	DEBUG bool `json:",omitempty"`
}

// ServerConfig, but with JWT_SECRET removed from json.
// Used for returning to user from get config api request.
//
// Technically only admins will have access to that api route,
// but I feel more comfortable removing it anyways (+ this is
// not editable on frontend, so not needed).
func (c *ServerConfig) GetSafe() ServerConfig {
	return ServerConfig{
		SIGNUP_ENABLED:         c.SIGNUP_ENABLED,
		JELLYFIN_HOST:          c.JELLYFIN_HOST,
		TMDB_KEY:               c.TMDB_KEY,
		DEBUG:                  c.DEBUG,
		SONARR_HOST:            c.SONARR_HOST,
		SONARR_KEY:             c.SONARR_KEY,
		SONARR_QUALITY_PROFILE: c.SONARR_QUALITY_PROFILE,
		RADARR_HOST:            c.RADARR_HOST,
		RADARR_KEY:             c.RADARR_KEY,
	}
}

var (
	// Our server config.. set defaults here, then `readConfig`
	// will overwrite if provided in watcharr.json cfg file.
	Config = ServerConfig{
		SIGNUP_ENABLED: true,
	}
	sonarr = arr.New(arr.SONARR, &Config.SONARR_HOST, &Config.SONARR_KEY)
	radarr = arr.New(arr.RADARR, &Config.RADARR_HOST, &Config.RADARR_KEY)
)

// Read config file
// Calls generateConfig if file doesn't exist
func readConfig() error {
	cfg, err := os.Open("./data/watcharr.json")
	if err != nil {
		if os.IsNotExist(err) {
			slog.Info("Config file doesn't exist... generating.")
			if err = generateConfig(); err == nil {
				return nil
			}
		}
		return err
	}
	defer cfg.Close()
	jsonParser := json.NewDecoder(cfg)
	if err = jsonParser.Decode(&Config); err != nil {
		return err
	}
	initFromConfig()
	return nil
}

// Ensure required config is provided
func initFromConfig() error {
	if Config.JWT_SECRET == "" {
		log.Fatal("JWT_SECRET missing from config!")
	}
	return nil
}

// Generate new barebones watcharr.json config file.
// Currently only JWT_SECRET is required, so this method
// generates a secret.
func generateConfig() error {
	key, err := generateString(64)
	if err != nil {
		return err
	}
	cfg := ServerConfig{JWT_SECRET: key}
	barej, err := json.MarshalIndent(cfg, "", "\t")
	if err != nil {
		return err
	}
	Config.JWT_SECRET = cfg.JWT_SECRET
	return os.WriteFile("./data/watcharr.json", barej, 0755)
}

// Update server config property
func updateConfig(k string, v any) error {
	slog.Debug("updateConfig", "k", k, "v", v)
	if v == nil {
		return errors.New("invalid value")
	}
	if k == "JELLYFIN_HOST" {
		Config.JELLYFIN_HOST = v.(string)
	} else if k == "SIGNUP_ENABLED" {
		Config.SIGNUP_ENABLED = v.(bool)
	} else if k == "TMDB_KEY" {
		Config.TMDB_KEY = v.(string)
	} else if k == "SONARR_HOST" {
		Config.SONARR_HOST = v.(string)
	} else if k == "SONARR_KEY" {
		Config.SONARR_KEY = v.(string)
	} else if k == "SONARR_QUALITY_PROFILE" {
		// Not sure why v insists its a float64 but just going with it..
		Config.SONARR_QUALITY_PROFILE = int(v.(float64))
	} else if k == "RADARR_HOST" {
		Config.RADARR_HOST = v.(string)
	} else if k == "RADARR_KEY" {
		Config.RADARR_KEY = v.(string)
	} else if k == "DEBUG" {
		Config.DEBUG = v.(bool)
		setLoggingLevel()
	} else {
		return errors.New("invalid setting")
	}
	return writeConfig()
}

// Write current Config to file
func writeConfig() error {
	barej, err := json.MarshalIndent(Config, "", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile("./data/watcharr.json", barej, 0755)
}
