package shorturl_service

import (
	"github.com/gin-gonic/gin"
	"github.com/zhuyst/shorturl-service/helper"
	"testing"
)

func TestInitRouter(t *testing.T) {
	initTestRouter(t)
	t.Logf("InitRouter PASS")
}

func initTestRouter(t *testing.T) *gin.Engine {
	r := gin.Default()
	redisClient := helper.NewTestRedisClient()
	err := InitRouter(r, redisClient, &Option{
		Domain: "d.zhuyst.cc",
	})
	if err != nil {
		t.Fatalf("initRouter ERROR: %s", err.Error())
		return nil
	}

	return r
}
