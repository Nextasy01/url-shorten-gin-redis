package handlers

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Nextasy01/url-shorten-gin-redis/api/repository"
	"github.com/Nextasy01/url-shorten-gin-redis/api/services"
	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type URLShortenHandler struct {
	db *repository.Database
}

type request struct {
	URL         string `json:"url"`
	CustomShort string `json:"short"`
}

type response struct {
	URL             string    `json:"url"`
	Short           string    `json:"short"`
	Expiry          time.Time `json:"expiry"`
	XRateRemaining  int       `json:"rate_limit"`
	XRateLimitReset time.Time `json:"rate_limit_reset"`
}

func NewURLShortenHandler(db *repository.Database) URLShortenHandler {
	return URLShortenHandler{db}
}

func (u *URLShortenHandler) ShortenURL(ctx *gin.Context) {

	// This is my timezone, you can change that line for whichever you prefer
	// Complete list of timezones you can find here:
	// https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
	tz, _ := time.LoadLocation("Asia/Samarkand")

	body := new(request)

	resp := new(response)

	// Getting and binding body parameters
	if err := ctx.BindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	r := u.db.CreateConnection(1)
	defer r.Close()

	// Checking if such client exists already
	// if not then, adding to db
	// if yes(else cond) then checking if rate limit exceeded or not
	val, err := r.Get(u.db.Conn, ctx.ClientIP()).Result()
	if err == redis.Nil {
		_ = r.Set(u.db.Conn, ctx.ClientIP(), os.Getenv("API_QUOTA"), 30*60*time.Second).Err()
	} else {
		valInt, _ := strconv.Atoi(val)
		log.Println("API_QUOTA: " + os.Getenv("API_QUOTA"))
		log.Println(err.Error())
		if valInt <= 0 {
			limit, _ := r.TTL(u.db.Conn, ctx.ClientIP()).Result()
			ctx.JSON(http.StatusServiceUnavailable, gin.H{
				"error":            "Your rate limit is exceeded",
				"rate_limit_reset": limit / time.Nanosecond / time.Minute,
			})
			return
		}
	}
	// Validating URL
	if !govalidator.IsURL(body.URL) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "this url is not valid!",
		})
		return
	}

	// Checking if user abuses domain by using it as shorting which may lead to infinite loop
	if !services.RemoveDomainError(body.URL) {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "You think you so smart? I don't think so.",
		})
		return
	}

	// convert URL to http:// format
	body.URL = services.EnforceHTTP(body.URL)

	var id string
	// Check if user provided own custom shorting
	// If not then generate new one
	if body.CustomShort == "" {
		id = uuid.New().String()[:6]
	} else {
		id = body.CustomShort
	}

	r2 := u.db.CreateConnection(0)
	defer r2.Close()
	// Check if provided user short url is already in use
	val, _ = r2.Get(u.db.Conn, id).Result()
	if val != "" {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "Provided short URL is already in use",
		})
		return
	}

	if err = r2.Set(u.db.Conn, id, body.URL, 24*3600*time.Second).Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unable to connect to server",
		})
		return
	}

	r.Decr(u.db.Conn, ctx.ClientIP())

	resp.URL = body.URL
	resp.Short = os.Getenv("DOMAIN") + "/" + id
	resp.Expiry = time.Now().In(tz).Add(24 * 3600 * time.Second)

	val, _ = r.Get(u.db.Conn, ctx.ClientIP()).Result()
	resp.XRateRemaining, _ = strconv.Atoi(val)

	ttl, _ := r.TTL(u.db.Conn, ctx.ClientIP()).Result()
	resp.XRateLimitReset = time.Now().In(tz).Add(ttl)

	ctx.JSON(http.StatusOK, gin.H{
		"data": resp,
	})
}
