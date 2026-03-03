package config

import (
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"os"
	"path"
	"time"

	"github.com/sbondCo/Watcharr/config/cfgmodel"
	"github.com/sbondCo/Watcharr/logging"
	"github.com/sbondCo/Watcharr/media/igdb"
	"github.com/sbondCo/Watcharr/util"
)

var DataPath = func() string {
	path := os.Getenv("WATCHARR_DATA")
	if path == "" {
		path = "./data"
	}
	return path
}()

type TrustedHeaderAuthSetting struct {
	// Required: Should header auth be enabled?
	// This bool exists so header auth can be toggled
	// easily without having to remove configuration.
	// To be actually enabled, HEADER_NAME must also
	// be set.
	Enabled bool `json:"enabled"`
	// Required: What is the name of the trusted header
	// that will contain the logged in users username?
	HeaderName string `json:"headerName"`
	// Should the frontend attempt auto login if
	// trusted header auth is enabled.
	AutoLogin bool `json:"autoLogin"`
	// Where can we redirect the user to logout
	// of the auth service?
	LogoutUrl string `json:"logoutUrl"`
}

type ServerConfig struct {
	// Used to sign JWT tokens. Make sure to make
	// it strong, just like a very long, complicated password.
	JWT_SECRET string `json:",omitempty"`

	// Default country for new users. This is used to set the default
	// region to get correct content streaming providers.
	// TODO Enforce iso_3166_1 validity (same as tmdb)
	DEFAULT_COUNTRY string `json:",omitempty"`

	// Optional: Point to your Jellyfin install
	// to enable it as an auth provider.
	JELLYFIN_HOST string `json:",omitempty"`

	// Optional: Use Emby instead of Jellyfin branding in the ui.
	USE_EMBY bool

	// Enable/disable signup functionality.
	// Set to `false` to disable registering an account.
	SIGNUP_ENABLED bool

	// Optional: Provide your own TMDB API Key.
	// If unprovided, the default Watcharr API key will be used.
	TMDB_KEY string `json:",omitempty"`

	// Optional: Point to Plex install to enable plex features.
	PLEX_HOST string `json:",omitempty"`

	// Optional: Machine identifier of your Plex server.
	// This is used to ensure only users of your Plex server
	// can use this Watcharr instance.
	// Will be fetched automatically when PLEX_HOST is provided via web ui.
	PLEX_MACHINE_ID string `json:",omitempty"`

	// Optional: Trusted header authentication configuration.
	// VERY DANGEROUS if access is not controlled correctly!
	HEADER_AUTH TrustedHeaderAuthSetting `json:",omitempty"`

	SONARR []cfgmodel.SonarrSettings `json:",omitempty"`
	RADARR []cfgmodel.RadarrSettings `json:",omitempty"`
	TWITCH igdb.IGDB                 `json:",omitempty"`

	// Optional: Schedule for tasks.
	TASK_SCHEDULE map[string]int `json:",omitempty"`

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
		SIGNUP_ENABLED:  c.SIGNUP_ENABLED,
		DEFAULT_COUNTRY: c.DEFAULT_COUNTRY,
		JELLYFIN_HOST:   c.JELLYFIN_HOST,
		USE_EMBY:        c.USE_EMBY,
		TMDB_KEY:        c.TMDB_KEY,
		PLEX_HOST:       c.PLEX_HOST,
		PLEX_MACHINE_ID: c.PLEX_MACHINE_ID,
		DEBUG:           c.DEBUG,
		SONARR:          c.SONARR, // Dont act safe, this contains sonarr api key, needed for config
		RADARR:          c.RADARR, // Dont act safe, this contains radarr api key, needed for config
		TWITCH: igdb.IGDB{
			ClientID:     c.TWITCH.ClientID,
			ClientSecret: c.TWITCH.ClientSecret,
		}, // Dont act safe, this contains twitch secrets, needed for config
	}
}

type ServerConfigGetByName struct {
	Value any `json:"value"`
}

// Get config item by name.
func (c *ServerConfig) Get(s string) (ServerConfigGetByName, error) {
	switch s {
	case "DEFAULT_COUNTRY":
		return ServerConfigGetByName{Value: c.DEFAULT_COUNTRY}, nil
	case "JELLYFIN_HOST":
		return ServerConfigGetByName{Value: c.JELLYFIN_HOST}, nil
	case "USE_EMBY":
		return ServerConfigGetByName{Value: c.USE_EMBY}, nil
	case "SIGNUP_ENABLED":
		return ServerConfigGetByName{Value: c.SIGNUP_ENABLED}, nil
	case "TMDB_KEY":
		return ServerConfigGetByName{Value: c.TMDB_KEY}, nil
	case "PLEX_HOST":
		return ServerConfigGetByName{Value: c.PLEX_HOST}, nil
	case "PLEX_MACHINE_ID":
		return ServerConfigGetByName{Value: c.PLEX_MACHINE_ID}, nil
	case "HEADER_AUTH":
		return ServerConfigGetByName{Value: c.HEADER_AUTH}, nil
	case "DEBUG":
		return ServerConfigGetByName{Value: c.DEBUG}, nil
	}
	return ServerConfigGetByName{}, errors.New("invalid setting")
}

// Update server config property
func (c *ServerConfig) UpdateConfig(k string, v any) error {
	slog.Debug("updateConfig", "k", k, "v", v)
	if v == nil {
		return errors.New("invalid value")
	}
	if k == "JELLYFIN_HOST" {
		c.JELLYFIN_HOST = v.(string)
	} else if k == "USE_EMBY" {
		c.USE_EMBY = v.(bool)
	} else if k == "SIGNUP_ENABLED" {
		c.SIGNUP_ENABLED = v.(bool)
	} else if k == "TMDB_KEY" {
		c.TMDB_KEY = v.(string)
	} else if k == "DEBUG" {
		c.DEBUG = v.(bool)
		logging.SetLevel(c.DEBUG)
	} else if k == "DEFAULT_COUNTRY" {
		c.DEFAULT_COUNTRY = v.(string)
	} else {
		return errors.New("invalid setting")
	}
	err := c.Write()
	if err != nil {
		slog.Error("updateConfig: Failed to write updated config!", "error", err)
		return errors.New("failed to write config")
	}
	return nil
}

// Write current Config to file
func (c *ServerConfig) Write() error {
	barej, err := json.MarshalIndent(*c, "", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile(path.Join(DataPath, "watcharr.json"), barej, 0755)
}

func (c *ServerConfig) SaveTwitchConfig(newt igdb.IGDB) error {
	// If existing client id and secret are same.. just return here
	if (c.TWITCH.ClientID != nil && newt.ClientID != nil && c.TWITCH.ClientSecret != nil && newt.ClientSecret != nil) &&
		*c.TWITCH.ClientID == *newt.ClientID && *c.TWITCH.ClientSecret == *newt.ClientSecret {
		slog.Info("SaveTwitchConfig: New ClientID and ClientSecret match old ClientID and ClientSecret.. ignoring request to update.")
		return nil
	}
	// Update our config
	c.TWITCH.ClientID = newt.ClientID
	c.TWITCH.ClientSecret = newt.ClientSecret
	c.TWITCH.AccessToken = ""
	c.TWITCH.AccessTokenExpires = time.Time{}
	// Try to init again
	err := c.TWITCH.Init()
	if err != nil {
		slog.Error("SaveTwitchConfig failed to initialize TWITCH", "error", err)
		return errors.New("initialization with credentials failed")
	}
	err = c.Write()
	if err != nil {
		slog.Error("SaveTwitchConfig failed to write config", "error", err)
		return errors.New("failed to save config")
	}
	return nil
}

func (c *ServerConfig) TwitchEnabled() bool {
	if c.TWITCH.ClientID != nil && c.TWITCH.ClientSecret != nil {
		return true
	}
	return false
}

// Read config file
// Calls generateConfig if file doesn't exist
func read() (*ServerConfig, error) {
	cfgFile, err := os.Open(path.Join(DataPath, "watcharr.json"))
	if err != nil {
		if os.IsNotExist(err) {
			slog.Info("Config file doesn't exist... generating.")
			if genCfg, err := generateConfig(); err == nil {
				return genCfg, nil
			}
		}
		return nil, err
	}
	defer cfgFile.Close()

	c := new(ServerConfig)
	dec := json.NewDecoder(cfgFile)
	if err = dec.Decode(c); err != nil {
		return nil, err
	}

	initFromConfig(c)

	return c, nil
}

// Ensure required config is provided
func initFromConfig(c *ServerConfig) {
	if c.JWT_SECRET == "" {
		log.Fatal("JWT_SECRET missing from config!")
	}
}

// Generate new barebones watcharr.json config file.
// Generates a JWT_SECRET and set default config.
func generateConfig() (*ServerConfig, error) {
	key, err := util.GenerateString(64)
	if err != nil {
		return nil, err
	}
	cfg := ServerConfig{
		JWT_SECRET: key,
		// Other defaults..
		DEFAULT_COUNTRY: "US",
		SIGNUP_ENABLED:  true,
	}
	barej, err := json.MarshalIndent(cfg, "", "\t")
	if err != nil {
		return nil, err
	}
	return &cfg, os.WriteFile(path.Join(DataPath, "watcharr.json"), barej, 0755)
}

// Get server config.
// Reads from config file.
func Get() (*ServerConfig, error) {
	cfg, err := read()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
