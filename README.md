### 极简分享

[Docker镜像](https://hub.docker.com/repository/docker/kingwrcy/hotnews)

| 环境变量          | 解释                  | 示例                                                                                                             |
|---------------|---------------------|----------------------------------------------------------------------------------------------------------------|
| PORT          | 监听端口                | 选填,默认32919                                                                                                     |
| COOKIE_SECRET | cookie密钥            | 必填,如:UbnpjqcvDJ8mDCB                                                                                           |
| STATIC_CDN_PREFIX | 静态资源CDN前缀           | 选填,默认取使用本地静态文件                                                                                                 |
| DB            | 数据库链接,目前只支持Postgres | 必填,'host=localhost user=username password=password dbname=hn port=5432 sslmode=disable TimeZone=Asia/Shanghai' |

