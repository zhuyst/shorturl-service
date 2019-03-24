package url_storage

import (
	"shorturl_service/helper"
	"strings"
	"testing"
)

func TestNewUrlStorage(t *testing.T) {
	_, err := newUrlStorage()
	if err != nil {
		t.Errorf("NewUrlStorage ERROR: %s", err.Error())
		return
	}

	t.Logf("NewUrlStorage PASS")
}

func TestUrlStorage_GenerateShortUrl(t *testing.T) {
	urlStorage, err := newUrlStorage()
	if err != nil {
		t.Errorf("NewUrlStorage ERROR: %s", err.Error())
		return
	}

	longUrl := "https://github.com/zhuyst"
	shortUrl, err := urlStorage.GenerateShortUrl(longUrl)
	if err != nil {
		t.Errorf("UrlStorage_GenerateShortUrl ERROR: %s", err.Error())
		return
	}

	keySplit := strings.Split(shortUrl, urlStorage.shortUrlPrefix)
	if len(keySplit) != 2 {
		t.Errorf("UrlStorage_GenerateShortUrl ERROR, "+
			"expected len(keySplit) == 2, got %d", len(keySplit))
		return
	}

	key := keySplit[1]
	if key == "" {
		t.Errorf("UrlStorage_GenerateShortUrl ERROR, expected not empty key, got %s", key)
		return
	}

	t.Logf("UrlStorage_GenerateShortUrl PASS, url: %s", shortUrl)
}

func TestUrlStorage_GetLongUrlByKey(t *testing.T) {
	urlStorage, err := newUrlStorage()
	if err != nil {
		t.Errorf("NewUrlStorage ERROR: %s", err.Error())
		return
	}

	longUrl := "https://github.com/zhuyst"
	shortUrl, err := urlStorage.GenerateShortUrl(longUrl)
	if err != nil {
		t.Errorf("UrlStorage_GetLongUrlByKey ERROR: %s", err.Error())
		return
	}

	key := strings.Split(shortUrl, urlStorage.shortUrlPrefix)[1]
	longUrlFromStorage, err := urlStorage.GetLongUrlByKey(key)
	if err != nil {
		t.Errorf("UrlStorage_GetLongUrlByKey ERROR: %s", err.Error())
		return
	}

	if longUrl != longUrlFromStorage {
		t.Errorf("UrlStorage_GetLongUrlByKey ERROR, expected longUrl == longUrlFromStorage, got false")
		return
	}

	t.Logf("UrlStorage_GetLongUrlByKey PASS, shortUrl: %s, longUrl: %s", shortUrl, longUrl)
}

func newUrlStorage() (*UrlStorage, error) {
	redisClient := helper.NewTestRedisClient()
	return New(redisClient, "https://d.zhuyst.cc/")
}
