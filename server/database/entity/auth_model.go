package entity

import (
	"github.com/golang-jwt/jwt/v5"
)

type ArgonParams struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

func GetPassArgonParams() *ArgonParams {
	return &ArgonParams{
		Memory:      64 * 1024,
		Iterations:  3,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}
}

type TokenClaims struct {
	UserID   uint     `json:"userId"`
	Username string   `json:"username"`
	Type     UserType `json:"type"`
	jwt.RegisteredClaims
}
