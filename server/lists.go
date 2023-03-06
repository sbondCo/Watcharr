package main

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

type Content struct {
	Name string `json:"name"`
}

type List struct {
	bun.BaseModel `bun:"table:lists"`

	ID     int `bun:"id,pk,autoincrement" json:"id"`
	ImdbID int `bun:"imdb_id" json:"imdbId"`
}

func getContent(db *bun.DB) List {
	ctx := context.TODO()
	list := new(List)
	err := db.NewSelect().Model(list).Where("id = ?", 1).Scan(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println(list.ID)
	fmt.Println(list.ImdbID)
	return *list
}
