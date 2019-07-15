# 短URL生成服务

[![Build Status](https://travis-ci.org/zhuyst/shorturl-service.svg?branch=master)](https://travis-ci.org/zhuyst/shorturl-service)
[![Go Report Card](https://goreportcard.com/badge/github.com/zhuyst/shorturl-service)](https://goreportcard.com/report/github.com/zhuyst/shorturl-service)
[![codecov](https://codecov.io/gh/zhuyst/shorturl-service/branch/master/graph/badge.svg)](https://codecov.io/gh/zhuyst/shorturl-service)

短URL生成服务，使用snowflake生成key，使用redis存储映射。

## 项目主要依赖

* Web服务: [Gin](https://github.com/gin-gonic/gin)
* 存储服务: [go-redis](https://github.com/go-redis/redis)
* Key值生成: [snowflake](https://github.com/bwmarrin/snowflake)

## 直接使用

1. 生成一个短URL：
```bash
curl -X POST \
  https://d.zhuyst.cc/new \
  -d 'url=https://github.com/zhuyst/shorturl-service'
  
{"code":200,"message":"OK","url":"https://d.zhuyst.cc/4dUaeq5"}
```

2. 在浏览器中输入<a href="https://d.zhuyst.cc/4dUaeq5" target="_blank">d.zhuyst.cc/4dUaeq5</a>

## 在原有服务添加短URL服务

1. 安装服务
```sh
go get -u github.com/zhuyst/shorturl-service
```

2. 引入服务
```go
import "github.com/zhuyst/shorturl-service"
```

3. 使用[go-redis](https://github.com/go-redis/redis)与[Gin](https://github.com/gin-gonic/gin)启动服务
```go
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/zhuyst/shorturl-service"
	"log"
	"regexp"
)

func main() {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})

	if err := redisClient.Ping().Err(); err != nil {
		log.Fatalf("redisClient ping FAIL: %s", err.Error())
		return
	}

	r := gin.Default()
	if err := shorturl_service.InitRouter(r, redisClient, &shorturl_service.Option{
		Domain:        "d.zhuyst.cc",
		ServiceUri:    "/",
		LongUrlRegexp: regexp.MustCompile("https://.*"),
	}); err != nil {
		log.Fatalf("shorturl_service init FAIL: %s", err.Error())
	}

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("gin init FAIL: %s", err.Error())
	}
}
```

## 从零搭建一个短URL服务

1. clone项目
```sh
git clone https://github.com/zhuyst/shorturl-service.git
```

2. 修改`example\main.go`中`Option`相关配置
```go
if err := shorturl_service.InitRouter(r, redisClient, &shorturl_service.Option{
		Domain:        "d.zhuyst.cc",
		ServiceUri:    "/",
		LongUrlRegexp: regexp.MustCompile("https://.*"),
	}); err != nil {
	log.Fatalf("shorturl_service init FAIL: %s", err.Error())
}
```

3. 使用`docker-compose`启动项目
```sh
docker-compose up -d
```
