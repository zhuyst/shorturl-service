package shorturl_service

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"net/http"
	"regexp"
	"shorturl_service/url-storage"
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

func (option *Option) initConfig() error {
	if option.LongUrlRegexp == nil {
		option.LongUrlRegexp = defaultLongUrlRegexp
	}

	if option.ServiceUri == "" {
		option.ServiceUri = defaultServiceUri
	}

	if option.Domain == "" {
		return errors.New("need option.domain")
	}

	return nil
}

func InitRouter(router *gin.Engine, redisClient *redis.Client, option *Option) error {
	if err := option.initConfig(); err != nil {
		return err
	}

	urlStorage, err := url_storage.New(redisClient, option.Domain)
	if err != nil {
		return err
	}
	option.urlStorage = urlStorage

	router.GET(fmt.Sprintf("%s:key", option.ServiceUri), option.redirectLongUrl)
	router.POST(fmt.Sprintf("%snew", option.ServiceUri), option.generateShortUrl)

	return nil
}

type result struct {
	code    int
	message string
	url     string
}

func (option *Option) redirectLongUrl(c *gin.Context) {
	key := c.Param("key")
	longUrl, err := option.urlStorage.GetLongUrlByKey(key)
	if err != nil {
		c.String(http.StatusNotFound, "%s not found", key)
		return
	}

	c.JSON(http.StatusMovedPermanently, longUrl)
}

func (option *Option) generateShortUrl(c *gin.Context) {
	longUrl, exists := c.GetPostForm("url")
	if !exists {
		c.JSON(http.StatusBadRequest, result{
			code:    http.StatusBadRequest,
			message: "required url",
		})
		return
	}

	if !option.LongUrlRegexp.MatchString(longUrl) {
		c.JSON(http.StatusBadRequest, result{
			code:    http.StatusBadRequest,
			message: "need prefix with https://",
		})
		return
	}

	shortUrl, err := option.urlStorage.GenerateShortUrl(longUrl)
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
