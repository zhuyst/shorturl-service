package shorturl_service

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

const longUrl = `https://github.com/zhuyst/shorturl-service`

func TestGenerateShortUrl(t *testing.T) {
	r := initTestRouter(t)

	w := getGenerateShortUrlRecorder(r, longUrl)
	res := w.Result()
	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)
	t.Logf("GenerateShortUrl body: %s", string(body))

	if res.StatusCode != http.StatusOK {
		t.Errorf("GenerateShortUrl ERROR, expected 200, got %d", res.StatusCode)
		return
	}

	var result result
	if err := json.Unmarshal(body, &result); err != nil {
		t.Errorf("GenerateShortUrl jsonParseError: %s", err.Error())
		return
	}

	if result.Code != http.StatusOK {
		t.Errorf("GenerateShortUrl ERROR, expected 200, got %d", result.Code)
		return
	}

	prefix := "https://d.zhuyst.cc/"
	if !strings.HasPrefix(result.Url, prefix) {
		t.Errorf("GenerateShortUrl ERROR, expected has prefix %s, got %s", prefix, result.Url)
		return
	}

	keySplit := strings.Split(result.Url, prefix)
	if len(keySplit) != 2 {
		t.Errorf("GenerateShortUrl ERROR, expected 2, got %d", len(keySplit))
		return
	}

	testRedirectLongUrl(t, r, keySplit[1])
}

func TestGenerateShortUrlError(t *testing.T) {
	r := initTestRouter(t)

	t.Run("Required url", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/new", nil)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("GenerateShortUrlError Required url ERROR, expected %d, got %d",
				http.StatusBadRequest, res.StatusCode)
			return
		}
	})
	t.Run("LongUrlRegexp", func(t *testing.T) {
		w := getGenerateShortUrlRecorder(r, "http://github.com/zhuyst/shorturl-service")
		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("GenerateShortUrlError LongUrlRegexp ERROR, expected %d, got %d",
				http.StatusBadRequest, res.StatusCode)
			return
		}
	})
}

func TestRedirectLongUrlError(t *testing.T) {
	r := initTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/zhuyst", nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusNotFound {
		t.Errorf("RedirectLongUrlError ERROR, expected %d, got %d",
			http.StatusNotFound, res.StatusCode)
		return
	}

	t.Logf("RedirectLongUrlError PASS")
}

func testRedirectLongUrl(t *testing.T, r *gin.Engine, key string) {
	if key == "" {
		t.Error("RedirectLongUrl ERROR, expected not empty key, got empty key")
		return
	}

	req := httptest.NewRequest(http.MethodGet, "/"+key, nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusMovedPermanently {
		t.Errorf("RedirectLongUrl ERROR, expected %d, got %d",
			http.StatusMovedPermanently, res.StatusCode)
		return
	}

	location := res.Header.Get("Location")
	if location != longUrl {
		t.Errorf("RedirectLongUrl ERROR, expected %s, got %s", longUrl, location)
		return
	}

	t.Logf("RedirectLongUrl PASS")
}

func getGenerateShortUrlRecorder(r *gin.Engine, longUrl string) *httptest.ResponseRecorder {
	form := url.Values{}
	form.Add("url", longUrl)
	req := httptest.NewRequest(http.MethodPost, "/new", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
