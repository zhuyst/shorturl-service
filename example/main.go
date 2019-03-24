package main

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"log"
	"regexp"
	"shorturl_service"
)

func main() {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})

	if err := redisClient.Ping().Err(); err != nil {
		log.Fatalf("redisClient ping FAIL: %s", err.Error())
		return
	}

	r := gin.Default()
	if err := shorturl_service.InitRouter(r, redisClient, &shorturl_service.Option{
		Domain:        "d.zhuyst.cc",
		ServiceUri:    "/",
		LongUrlRegexp: regexp.MustCompile("https://.*"),
	}); err != nil {
		log.Fatalf("shorturl_service init FAIL: %s", err.Error())
	}

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("gin init FAIL: %s", err.Error())
	}
}
