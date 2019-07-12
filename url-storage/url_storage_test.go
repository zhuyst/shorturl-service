package url_storage

import (
	"github.com/zhuyst/shorturl-service/helper"
	"strings"
	"testing"
	"time"
)

func TestNewUrlStorage(t *testing.T) {
	newUrlStorage(t)
	t.Logf("NewUrlStorage PASS")
}

func TestUrlStorage_GenerateShortUrl(t *testing.T) {
	urlStorage := newUrlStorage(t)

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

func TestUrlStorage_GenerateShortUrlAgain(t *testing.T) {
	urlStorage := newUrlStorage(t)

	longUrl := "https://github.com/zhuyst"
	getNewShortUrl := func() string {
		shortUrl, err := urlStorage.GenerateShortUrl(longUrl)
		if err != nil {
			t.Fatalf("UrlStorage_GenerateShortUrl ERROR: %s", err.Error())
		}
		return shortUrl
	}
	firstShortUrl := getNewShortUrl()

	for i := 0; i < 3; i++ {
		time.Sleep(time.Second)
		newShortUrl := getNewShortUrl()
		if firstShortUrl != newShortUrl {
			t.Fatalf("TestUrlStorage_GenerateShortUrlAgain ERROR, "+
				"expected firstShortUrl == newShortUrl, "+
				"got first: %s, new: %s", firstShortUrl, newShortUrl)
		}
	}

	t.Logf("TestUrlStorage_GenerateShortUrlAgain PASS, url: %s", firstShortUrl)
}

func TestUrlStorage_GetLongUrlByKey(t *testing.T) {
	urlStorage := newUrlStorage(t)

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

func newUrlStorage(t *testing.T) *UrlStorage {
	redisClient := helper.NewTestRedisClient()

	urlStorage, err := New(redisClient, "https://d.zhuyst.cc/")
	if err != nil {
		t.Fatalf("NewUrlStorage ERROR: %s", err.Error())
		return nil
	}

	return urlStorage
}
