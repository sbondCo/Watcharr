package main

import (
	"crypto/rand"
	b64 "encoding/base64"
	"encoding/json"
	"log"
	"log/slog"
	"os"
)

type ServerConfig struct {
	// Used to sign JWT tokens. Make sure to make
	// it strong, just like a very long, complicated password.
	JWT_SECRET string

	// Optional: Point to your Jellyfin install
	// to enable it as an auth provider.
	JELLYFIN_HOST string `json:",omitempty"`

	// Enable/disable signup functionality.
	// Set to `false` to disable registering an account.
	SIGNUP_ENABLED bool `json:",omitempty"`

	// Optional: Provide your own TMDB API Key.
	// If unprovided, the default Watcharr API key will be used.
	TMDB_KEY string `json:",omitempty"`

	// Enable/disable debug logging. Useful for when trying
	// to figure out exactly what the server is doing at a point
	// of failure.
	// Set to `true` to enable.
	DEBUG bool `json:",omitempty"`

	// Optional: When not set we assume production, should only
	// be set to DEV when developing the app.
	// MODE string
}

var (
	// Our server config.. set defaults here, then `readConfig`
	// will overwrite if provided in watcharr.json cfg file.
	Config = ServerConfig{
		SIGNUP_ENABLED: true,
	}
	AvailableAuthProviders = []string{}
	TMDBKey                = "d047fa61d926371f277e7a83c9c4ff2c"
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
// and initialize from the config if required (update vars)
func initFromConfig() error {
	if Config.JWT_SECRET == "" {
		log.Fatal("JWT_SECRET missing from config!")
	}

	if Config.JELLYFIN_HOST != "" {
		slog.Info("Adding Jellyfin as an auth provider.")
		AvailableAuthProviders = append(AvailableAuthProviders, "jellyfin")
	}

	if Config.TMDB_KEY != "" {
		slog.Info("Default TMDBKey being overriden by TMDB_KEY from config.")
		TMDBKey = Config.TMDB_KEY
	}
	return nil
}

// Generate new barebones watcharr.json config file.
// Currently only JWT_SECRET is required, so this method
// generates a secret.
func generateConfig() error {
	key := make([]byte, 64)
	_, err := rand.Read(key)
	if err != nil {
		return err
	}
	encKey := b64.StdEncoding.EncodeToString([]byte(key))
	cfg := ServerConfig{JWT_SECRET: encKey}
	barej, err := json.MarshalIndent(cfg, "", "\t")
	if err != nil {
		return err
	}
	Config = cfg
	return os.WriteFile("./data/watcharr.json", barej, 0755)
}
