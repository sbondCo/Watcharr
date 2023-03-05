package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("Watcharr Starting")
	router := gin.Default()
	addContentRoutes(router)

	router.Run("localhost:3080")
}
