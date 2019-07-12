package helper

import (
	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis"
)

// NewTestRedisClient 创建一个Mock Redis
func NewTestRedisClient() *redis.Client {
	ms, err := miniredis.Run()
	if err != nil {
		panic(err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: ms.Addr(),
	})
	if err = redisClient.Ping().Err(); err != nil {
		panic(err)
	}

	return redisClient
}
