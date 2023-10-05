package main

import (
	"crypto/rand"
	b64 "encoding/base64"
)

// Generate a random string
func generateString(len int) (string, error) {
	key := make([]byte, len)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	return b64.StdEncoding.EncodeToString([]byte(key)), nil
}
