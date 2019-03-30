# 短URL生成服务

短URL生成服务，使用snowflake生成key，使用redis存储映射

## 使用

* 生成一个短URL：
```sh
curl -X POST \
  https://d.zhuyst.cc/new \
  -d 'url=https://github.com/zhuyst/shorturl-service'
  
{"code":200,"message":"OK","url":"https://d.zhuyst.cc/4dUaeq5"}
```

* 在浏览器中输入`d.zhuyst.cc/4dUaeq5`

## 快速搭建短URL服务

```sh
git clone https://github.com/zhuyst/shorturl-service.git
docker-compose up -d
```