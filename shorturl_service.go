package shorturl_service

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/zhuyst/shorturl-service/url-storage"
	"regexp"
)

var (
	defaultLongUrlRegexp = regexp.MustCompile("https://.*")
	defaultServiceUri    = "/"
)

type Option struct {
	LongUrlRegexp *regexp.Regexp
	Domain        string
	ServiceUri    string

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

	if option.Domain == "" {
		return errors.New("need option.domain")
	}

	shortUrlPrefix := fmt.Sprintf("https://%s%s", option.Domain, option.ServiceUri)
	urlStorage, err := url_storage.New(redisClient, shortUrlPrefix)
	if err != nil {
		return err
	}
	option.urlStorage = urlStorage

	return nil
}
