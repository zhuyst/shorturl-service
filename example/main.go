package main

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/zhuyst/shorturl-service"
	"github.com/zhuyst/shorturl-service/logger"
	"regexp"
	"time"
)

func main() {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})

	if err := redisClient.Ping().Err(); err != nil {
		logger.Fatal("redisClient ping FAIL: %s", err.Error())
		return
	}

	r := gin.Default()
	if err := shorturl_service.InitRouter(r, redisClient, &shorturl_service.Option{
		Domain:           "d.zhuyst.cc",
		ServiceUri:       "/",
		LongUrlRegexp:    regexp.MustCompile("https://.*"),
		Long2ShortExpire: time.Hour * 24 * 7,
	}); err != nil {
		logger.Fatal("shorturl_service init FAIL: %s", err.Error())
	}

	if err := r.Run(":8080"); err != nil {
		logger.Fatal("gin init FAIL: %s", err.Error())
	}
}
