package main

import (
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"os"

	"gorm.io/gorm"
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

	SONARR []SonarrSettings `json:",omitempty"`
	RADARR []RadarrSettings `json:",omitempty"`

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
		SIGNUP_ENABLED: c.SIGNUP_ENABLED,
		JELLYFIN_HOST:  c.JELLYFIN_HOST,
		TMDB_KEY:       c.TMDB_KEY,
		DEBUG:          c.DEBUG,
		SONARR:         c.SONARR, // Dont act safe, this contains sonarr api key, needed for config
		RADARR:         c.RADARR, // Dont act safe, this contains radarr api key, needed for config
	}
}

var (
	// Our server config.. set defaults here, then `readConfig`
	// will overwrite if provided in watcharr.json cfg file.
	Config = ServerConfig{
		SIGNUP_ENABLED: true,
	}
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

type ServerFeatures struct {
	Sonarr bool `json:"sonarr"`
	Radarr bool `json:"radarr"`
}

// Get enabled server functionality from Config.
// Mainly so the frontend can store this once and know
// which btns should be shown, etc.
func getEnabledFeatures(userPerms int) ServerFeatures {
	var f ServerFeatures
	// https://github.com/sbondCo/Watcharr/issues/211
	// Currently requesting permissions have not been setup, only admins for now.
	if !hasPermission(userPerms, PERM_ADMIN) {
		return f
	}
	if len(Config.SONARR) > 0 {
		f.Sonarr = true
	}
	if len(Config.RADARR) > 0 {
		f.Radarr = true
	}
	return f
}

type ServerStats struct {
	Users            int64   `json:"users"`
	PrivateUsers     int64   `json:"privateUsers"`
	WatchedMovies    int64   `json:"watchedMovies"`
	WatchedShows     int64   `json:"watchedShows"`
	WatchedSeasons   int64   `json:"watchedSeasons"`
	MostWatchedMovie Content `json:"mostWatchedMovie"`
	MostWatchedShow  Content `json:"mostWatchedShow"`
	Activities       int64   `json:"activities"`
}

// Collect and return server stats
// I cant sql so this the best yall gettin
func getServerStats(db *gorm.DB) ServerStats {
	stats := ServerStats{}
	db.Model(&User{}).Count(&stats.Users).Where("private = 1").Count(&stats.PrivateUsers)
	db.Model(&WatchedSeason{}).Count(&stats.WatchedSeasons)
	db.Model(&Activity{}).Count(&stats.Activities)
	db.Joins("JOIN contents ON contents.id = watcheds.content_id AND contents.type = ?", "tv").Find(&Watched{}).Count(&stats.WatchedShows)
	db.Joins("JOIN contents ON contents.id = watcheds.content_id AND contents.type = ?", "movie").Find(&Watched{}).Count(&stats.WatchedMovies)
	var w Watched
	db.Model(&Watched{}).Select("content_id, COUNT(*) AS mag").Joins("JOIN contents ON contents.type = ? AND contents.id = watcheds.content_id", "tv").Group("content_id").Order("mag DESC").Preload("Content").First(&w)
	stats.MostWatchedShow = w.Content
	db.Model(&Watched{}).Select("content_id, COUNT(*) AS mag").Joins("JOIN contents ON contents.type = ? AND contents.id = watcheds.content_id", "movie").Group("content_id").Order("mag DESC").Preload("Content").First(&w)
	stats.MostWatchedMovie = w.Content
	return stats
}
