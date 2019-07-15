package shorturl_service

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/zhuyst/shorturl-service/logger"
	"github.com/zhuyst/shorturl-service/url-storage"
	"regexp"
	"time"
)

var (
	// defaultLongUrlRegexp 默认的长url验证表达式
	defaultLongUrlRegexp = regexp.MustCompile("https://.*")

	// defaultServiceUri 服务的默认uri前缀
	defaultServiceUri = "/"
)

// Option shorturl-service的配置项
type Option struct {
	LongUrlRegexp    *regexp.Regexp // 长url验证表达式，不设置默认为defaultLongUrlRegexp
	Domain           string         // 短url服务的域名
	ServiceUri       string         // 短url服务的uri前缀
	Long2ShortExpire time.Duration  // 长-短url映射缓存的过期时间，默认为不过期

	Logger logger.ILogger

	urlStorage *url_storage.UrlStorage
}

func InitRouter(router *gin.Engine, redisClient *redis.Client, option *Option) error {
	if err := option.initConfig(redisClient); err != nil {
		return err
	}

	router.GET(fmt.Sprintf("%s:key", option.ServiceUri), option.redirectLongUrl)
	router.POST(fmt.Sprintf("%snew", option.ServiceUri), option.generateShortUrl)

	return nil
}

func (option *Option) initConfig(redisClient *redis.Client) error {
	if option.LongUrlRegexp == nil {
		option.LongUrlRegexp = defaultLongUrlRegexp
	}

	if option.ServiceUri == "" {
		option.ServiceUri = defaultServiceUri
	}

	if option.Logger != nil {
		logger.Logger = option.Logger
	}

	if option.Domain == "" {
		return errors.New("need option.domain")
	}

	// Domain + ServiceUri作为短url前缀
	shortUrlPrefix := fmt.Sprintf("https://%s%s", option.Domain, option.ServiceUri)
	urlStorage, err := url_storage.New(redisClient, shortUrlPrefix)
	if err != nil {
		return err
	}
	option.urlStorage = urlStorage

	return nil
}
