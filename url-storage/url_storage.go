package url_storage

import (
	"fmt"
	"github.com/go-redis/redis"
	"shorturl_service/key-generator"
)

const (
	shortUrlKey = "SHORTURL_SERVICE:SHORT_URL"
)

type UrlStorage struct {
	domain       string
	redisClient  *redis.Client
	keyGenerator *key_generator.KeyGenerator
}

func New(redisClient *redis.Client, domain string) (*UrlStorage, error) {
	if err := redisClient.Ping().Err(); err != nil {
		return nil, err
	}

	keyGenerator, err := key_generator.New(redisClient)
	if err != nil {
		return nil, err
	}

	return &UrlStorage{
		domain:       domain,
		redisClient:  redisClient,
		keyGenerator: keyGenerator,
	}, nil
}

func (storage *UrlStorage) GenerateShortUrl(longUrl string) (string, error) {
	key := storage.keyGenerator.Generate()
	if err := storage.redisClient.HSet(shortUrlKey, key, longUrl).Err(); err != nil {
		return "", err
	}

	return fmt.Sprintf(storage.domain + key), nil
}

func (storage *UrlStorage) GetLongUrlByKey(key string) (string, error) {
	return storage.redisClient.HGet(shortUrlKey, key).Result()
}
