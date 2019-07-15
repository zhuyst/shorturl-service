package url_storage

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/zhuyst/shorturl-service/key-generator"
	"github.com/zhuyst/shorturl-service/logger"
	"time"
)

const (
	// shortUrlKey 存储短-长ur映射关系的hash Key
	shortUrlKey = "SHORTURL_SERVICE:SHORT_URL"

	// long2ShortKeyPrefix 存储长-短url映射关系的key前缀
	// 为了能使用expire功能故不使用hash结构存储
	long2ShortKeyPrefix = "SHORTURL_SERVICE:LONG2SHORT:%s"
)

// UrlStorage url存储
type UrlStorage struct {
	shortUrlPrefix   string        // 短url前缀(一般为域名)
	long2ShortExpire time.Duration // 长-短url映射关系的过期时间

	redisClient  *redis.Client
	keyGenerator *key_generator.KeyGenerator
}

// New 实例化一个UrlStorage
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

// GenerateShortUrl 生成短url
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

	// 长-短url的临时缓存
	go func() {
		long2ShortKey := fmt.Sprintf(long2ShortKeyPrefix, longUrl)
		if err := storage.redisClient.Set(long2ShortKey, key, storage.long2ShortExpire).Err(); err != nil {
			logger.Error("SET long2ShortUrl FAIL, longUrl: %s, key: %s", longUrl, key)
		}
	}()

	return storage.getShortUrl(key), nil
}

// GetLongUrlByKey 通过Key值获取长url
func (storage *UrlStorage) GetLongUrlByKey(key string) (string, error) {
	return storage.redisClient.HGet(shortUrlKey, key).Result()
}

// long2ShortUrl 通过长url获取短url(临时缓存)
func (storage *UrlStorage) long2ShortUrl(longUrl string) (string, error) {
	long2ShortKey := fmt.Sprintf(long2ShortKeyPrefix, longUrl)
	key, err := storage.redisClient.Get(long2ShortKey).Result()
	if err != nil {
		return "", err
	}

	return storage.getShortUrl(key), nil
}

// getShortUrl 通过key值拼装出完整的短url
func (storage *UrlStorage) getShortUrl(key string) string {
	return fmt.Sprintf(storage.shortUrlPrefix + key)
}
