package url_storage

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/zhuyst/shorturl-service/key-generator"
)

const (
	shortUrlKey = "SHORTURL_SERVICE:SHORT_URL"
)

type UrlStorage struct {
	shortUrlPrefix string
	redisClient    *redis.Client
	keyGenerator   *key_generator.KeyGenerator
}

func New(redisClient *redis.Client, shortUrlPrefix string) (*UrlStorage, error) {
	if err := redisClient.Ping().Err(); err != nil {
		return nil, err
	}

	keyGenerator, err := key_generator.New(redisClient)
	if err != nil {
		return nil, err
	}

	return &UrlStorage{
		shortUrlPrefix: shortUrlPrefix,
		redisClient:    redisClient,
		keyGenerator:   keyGenerator,
	}, nil
}

func (storage *UrlStorage) GenerateShortUrl(longUrl string) (string, error) {
	key := storage.keyGenerator.Generate()
	if err := storage.redisClient.HSet(shortUrlKey, key, longUrl).Err(); err != nil {
		return "", err
	}

	return fmt.Sprintf(storage.shortUrlPrefix + key), nil
}

func (storage *UrlStorage) GetLongUrlByKey(key string) (string, error) {
	return storage.redisClient.HGet(shortUrlKey, key).Result()
}
