package main

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

type Content struct {
	bun.BaseModel `bun:"table:content"`

	ID   int    `bun:"id,pk,autoincrement" json:"id"`
	Type string `json:"-"`
	Name string `json:"name"`
}

type List struct {
	bun.BaseModel `bun:"table:lists"`

	ID        int  `bun:"id,pk,autoincrement" json:"id"`
	Watched   bool `bun:"watched" json:"watched"`
	UserID    int  `bun:"user_id" json:"-"`
	ContentID int
	Content   *Content `bun:"rel:belongs-to,join:content_id=id" json:"content"`
}

func getContent(db *bun.DB) List {
	ctx := context.TODO()
	list := new(List)
	err := db.NewSelect().Model(list).Relation("Content").Where("user_id = ?", 8).Scan(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println(list.ID)
	return *list
}
