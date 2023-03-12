package main

import (
	"context"
	"errors"
	"strings"

	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users"`

	ID       int    `bun:"id,pk,autoincrement" json:"id"`
	Username string `bun:"username,notnull,unique" json:"username" binding:"required"`
	Password string `bun:"password,notnull" json:"password" binding:"required"`
}

type RegisterResponse struct {
	Token string `json:"token"`
}

func register(user *User, db *bun.DB) (RegisterResponse, error) {
	println("Registering", user.Username)
	_, err := db.NewInsert().Model(user).Exec(context.TODO())
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			println(err.Error())
			return RegisterResponse{}, errors.New("User already exists")
		}
		panic(err)
	}
	return RegisterResponse{Token: "My JWT token"}, nil
}
