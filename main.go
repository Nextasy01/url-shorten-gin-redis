package main

import (
	"os"

	"github.com/Nextasy01/url-shorten-gin-redis/api/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	server := gin.Default()

	rg := server.Group("/")
	routes.PublicRoutes(rg)

	// Enabling enviromental variables
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	server.Run(":" + os.Getenv("PORT"))
}
