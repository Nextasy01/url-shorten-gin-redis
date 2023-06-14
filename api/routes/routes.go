package routes

import (
	"github.com/Nextasy01/url-shorten-gin-redis/api/handlers"
	"github.com/Nextasy01/url-shorten-gin-redis/api/repository"
	"github.com/gin-gonic/gin"
)

var (
	DB                = repository.NewDatabase()
	urlShortenHandler = handlers.NewURLShortenHandler(&DB)
	urlResolveHandler = handlers.NewURLResolveHandler(&DB)
)

func PublicRoutes(g *gin.RouterGroup) {
	g.GET("/:url", urlResolveHandler.ResolveURL)
	g.POST("/api/v1", urlShortenHandler.ShortenURL)

}
