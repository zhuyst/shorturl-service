package url_storage

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/zhuyst/shorturl-service/key-generator"
	"github.com/zhuyst/shorturl-service/logger"
	"time"
)

const (
	shortUrlKey         = "SHORTURL_SERVICE:SHORT_URL"
	long2ShortKeyPrefix = "SHORTURL_SERVICE:LONG2SHORT:%s"
)

type UrlStorage struct {
	shortUrlPrefix   string
	long2ShortExpire time.Duration

	redisClient  *redis.Client
	keyGenerator *key_generator.KeyGenerator
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

	// 检查是否存在相同的longUrl，有则直接返回
	shortUrl, err := storage.long2ShortUrl(longUrl)
	if err == nil {
		return shortUrl, nil
	} else if err != redis.Nil {
		return "", err
	}

	key := storage.keyGenerator.Generate()
	if err := storage.redisClient.HSet(shortUrlKey, key, longUrl).Err(); err != nil {
		return "", err
	}

	go func() {
		long2ShortKey := fmt.Sprintf(long2ShortKeyPrefix, longUrl)
		if err := storage.redisClient.Set(long2ShortKey, key, storage.long2ShortExpire).Err(); err != nil {
			logger.Error("SET long2ShortUrl FAIL, longUrl: %s, key: %s", longUrl, key)
		}
	}()

	return storage.getShortUrl(key), nil
}

func (storage *UrlStorage) GetLongUrlByKey(key string) (string, error) {
	return storage.redisClient.HGet(shortUrlKey, key).Result()
}

func (storage *UrlStorage) long2ShortUrl(longUrl string) (string, error) {
	long2ShortKey := fmt.Sprintf(long2ShortKeyPrefix, longUrl)
	key, err := storage.redisClient.Get(long2ShortKey).Result()
	if err != nil {
		return "", err
	}

	return storage.getShortUrl(key), nil
}

func (storage *UrlStorage) getShortUrl(key string) string {
	return fmt.Sprintf(storage.shortUrlPrefix + key)
}
