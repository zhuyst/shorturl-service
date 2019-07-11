package shorturl_service

import (
	"github.com/gin-gonic/gin"
	"github.com/zhuyst/shorturl-service/logger"
	"net/http"
)

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
		logger.Error("generateShortUrl FAIL, longUrl: %s, Error: %s", longUrl, err.Error())

		c.JSON(http.StatusInternalServerError, &result{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	logger.Info("generateShortUrl SUCCESS, %s - %s", shortUrl, longUrl)
	c.JSON(http.StatusOK, &result{
		Code:    http.StatusOK,
		Message: "OK",
		Url:     shortUrl,
	})
}
