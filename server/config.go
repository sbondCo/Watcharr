package main

import (
	"crypto/rand"
	b64 "encoding/base64"
	"encoding/json"
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

var Config ServerConfig

func readConfig() {

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
	barej, err := json.MarshalIndent(ServerConfig{JWT_SECRET: encKey}, "", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile("./data/watcharr.json", barej, 0755)
}
