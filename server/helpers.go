package main

import (
	"crypto/rand"
	b64 "encoding/base64"
)

// Generate a random string, changed to safe64 encoding to be used in urls, this shouldn't affect anything program wide.
// if anything it limits the pool of characters to choose from for the JWT signing secret, however looking over where this is used elsewhere, it still gives
// you extremely high entropy secrets that are extremely safe to use.
func generateString(len int) (string, error) {
	key := make([]byte, len)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	return b64.RawURLEncoding.EncodeToString(key), nil
}
