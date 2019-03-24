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

	shortUrlPrefix := fmt.Sprintf("https://%s%s", option.Domain, option.ServiceUri)
	urlStorage, err := url_storage.New(redisClient, shortUrlPrefix)
	if err != nil {
		return err
	}
	option.urlStorage = urlStorage

	router.GET(fmt.Sprintf("%s:key", option.ServiceUri), option.redirectLongUrl)
	router.POST(fmt.Sprintf("%snew", option.ServiceUri), option.generateShortUrl)

	return nil
}

type result struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Url     string `json:"url"`
}

func (option *Option) redirectLongUrl(c *gin.Context) {
	key := c.Param("key")
	longUrl, err := option.urlStorage.GetLongUrlByKey(key)
	if err != nil {
		c.String(http.StatusNotFound, "%s not found", key)
		return
	}

	c.Redirect(http.StatusMovedPermanently, longUrl)
}

func (option *Option) generateShortUrl(c *gin.Context) {
	longUrl, exists := c.GetPostForm("url")
	if !exists {
		c.JSON(http.StatusBadRequest, &result{
			Code:    http.StatusBadRequest,
			Message: "required url",
		})
		return
	}

	if !option.LongUrlRegexp.MatchString(longUrl) {
		c.JSON(http.StatusBadRequest, &result{
			Code:    http.StatusBadRequest,
			Message: "need prefix with https://",
		})
		return
	}

	shortUrl, err := option.urlStorage.GenerateShortUrl(longUrl)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &result{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &result{
		Code:    http.StatusOK,
		Message: "OK",
		Url:     shortUrl,
	})
}
