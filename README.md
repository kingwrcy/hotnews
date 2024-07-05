### [极简分享](https://hotnews.pw)

[Docker镜像](https://hub.docker.com/repository/docker/kingwrcy/hotnews)

| 环境变量          | 解释                  | 示例                                                                                                             |
|---------------|---------------------|----------------------------------------------------------------------------------------------------------------|
| PORT          | 监听端口                | 选填,默认32919                                                                                                     |
| COOKIE_SECRET | cookie密钥            | 必填,如:UbnpjqcvDJ8mDCB                                                                                           |
| STATIC_CDN_PREFIX | 静态资源CDN前缀           | 选填,默认取使用本地静态文件                                                                                                 |
| DB            | 数据库链接,目前只支持Postgres | 必填,'host=localhost user=username password=password dbname=hn port=5432 sslmode=disable TimeZone=Asia/Shanghai' |

默认第一个注册的用户是管理员,自行注册即可.

目前可管理的功能很少,唯一能做的就是添加父标签/子标签,设置标签颜色等.

后台带了个用户列表和ip统计等.

需要的朋友自行部署吧.

得意于强大的go的内嵌静态资源的功能,镜像包只有**6.29mb**,启动之后占用内存只有**28mb**.

极度适合小内存的机器.当然数据库另说.

![alt](https://openai-75050.gzc.vod.tencent-cloud.com/openaiassets_5ba4ebcbd2030fee5ac43c38e41a0f41_2579861720144999302.png 'title')


### docker启动

1. 随便着个目录,在这个目录底下新建`.env`文件,内容如下,每个字段含义上面有写
```dotenv
PORT=32919
DB='host=localhost user=postgres password=a123456 dbname=hn port=5432 sslmode=disable TimeZone=Asia/Shanghai'
COOKIE_SECRET=UbnpjqcvDJ8mDCB
```

2. 使用如下命令启动
```shell
docker run --name hotnews -d --env-file .env -p 32912:32912 kingwrcy/hotnews:latest
```

3. 打开浏览器访问`本地ip:32912`即可.