package handlers

import (
	"log"
	"net/http"

	"github.com/Nextasy01/url-shorten-gin-redis/api/repository"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

type URLResolveHandler struct {
	db *repository.Database
}

func NewURLResolveHandler(db *repository.Database) URLResolveHandler {
	return URLResolveHandler{db}
}

func (u *URLResolveHandler) ResolveURL(ctx *gin.Context) {
	url, _ := ctx.Params.Get("url")
	log.Println(url)
	r := u.db.CreateConnection(0)
	defer r.Close()

	val, err := r.Get(u.db.Conn, url).Result()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	} else if err == redis.Nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "this short url was not found in database",
		})
		return
	}

	rInr := u.db.CreateConnection(1)
	defer rInr.Close()

	_ = rInr.Incr(u.db.Conn, "counter")

	// delete url after user redirects to it
	_, err = rInr.Del(ctx, url).Result()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.Redirect(301, val)
}
