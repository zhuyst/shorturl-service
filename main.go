package main

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"log"
	"net/http"
	"regexp"
	"shorturl_service/url-storage"
)

var (
	longUrlRegexp = regexp.MustCompile("https://.*")
	urlStorage    *url_storage.UrlStorage
)

type result struct {
	code    int
	message string
	url     string
}

func init() {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})
	var err error
	urlStorage, err = url_storage.New(redisClient)
	if err != nil {
		log.Fatalf("redis connect fail, err: %s", err.Error())
		return
	}
}

func main() {

	r := gin.Default()
	r.GET("/:key", redirectLongUrl)
	r.POST("/new", generateShortUrl)
}

func redirectLongUrl(c *gin.Context) {
	key := c.Param("key")
	longUrl, err := urlStorage.GetLongUrlByKey(key)
	if err != nil {
		c.String(http.StatusNotFound, "%s not found", key)
		return
	}

	c.JSON(http.StatusMovedPermanently, longUrl)
}

func generateShortUrl(c *gin.Context) {
	longUrl, exists := c.GetPostForm("url")
	if !exists {
		c.JSON(http.StatusBadRequest, result{
			code:    http.StatusBadRequest,
			message: "required url",
		})
		return
	}

	if !longUrlRegexp.MatchString(longUrl) {
		c.JSON(http.StatusBadRequest, result{
			code:    http.StatusBadRequest,
			message: "need prefix with https://",
		})
		return
	}

	shortUrl, err := urlStorage.GenerateShortUrl(longUrl)
	if err != nil {
		c.JSON(http.StatusInternalServerError, result{
			code:    http.StatusInternalServerError,
			message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result{
		code:    http.StatusOK,
		message: "OK",
		url:     shortUrl,
	})
}
