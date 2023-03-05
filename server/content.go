package main

type Content struct {
	Name string `json:"name"`
}

func getContent() Content {
	return Content{
		Name: "hello",
	}
}
