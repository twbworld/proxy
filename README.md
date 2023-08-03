
**Proxy**
===========
[![](https://github.com/twbworld/proxy/workflows/ci/badge.svg?branch=main)](https://github.com/twbworld/proxy/actions)
[![](https://img.shields.io/github/tag/twbworld/proxy?logo=github)](https://github.com/twbworld/proxy)
![](https://img.shields.io/badge/language-golang-cyan)
[![](https://img.shields.io/github/license/twbworld/proxy)](https://github.com/twbworld/proxy/blob/main/LICENSE)

## 简介
**翻墙订阅链接服务 和 用户管理**

> 本项目配合 [trojan-go](https://github.com/p4gefau1t/trojan-go) 使用, 并使用到其规定的数据库

### 本项目有两个作用:

1. 用户管理(增删改查)
    > 程序作为中间人, 通过与`telegram-bot`进行交互,实现对接[trojan-go](https://github.com/p4gefau1t/trojan-go)数据库, 从而进行用户管理  
    > 如喜欢用文件而不是`telegram-bot`进行用户管理, 请使用`v0`版本  
    > 如喜欢用shell进行用户管理, 建议出门左转使用[Jrohy/trojan](https://github.com/Jrohy/trojan)哟
2. 程序返回客户端可识别的翻墙配置, 即订阅功能
    > 访问含用户名的特定订阅链接, 程序返回`trojan-go`和`v2ray`客户端可以识别的base64码或`clash`可识别配置文件; 在客户端上配置订阅链接即可翻墙;

## 目录结构 : 
``` sh
├── .editorconfig
├── .gitattributes
├── .github/
│   └── workflows/          #存放GitHub-Actions的工作流文件
│       ├── ci.yml
│       ├── main.yml
│       ├── push-images.yaml
│       ├── review.yml
│       └── test.yaml
├── .gitignore
├── LICENSE
├── README.md
├── Dockerfile          #构建docker镜像
├── config/
│   ├── appConfig.go    #项目配置
│   ├── clash.ini       #clash配置模板
│   ├── .trojan-go       #trojan-go配置文件,需自行新增
│   ├── config.go
│   ├── trojanGoConfig.go #trojan-go配置
│   ├── .env            #机密配置文件,数据库之类的
│   ├── .env.example    #配置模板
│   └── env.go
├── controller/         #MVC模式的C
│   └── index.go
├── dao/
│   ├── db_handle.go    #用户数据操作
│   ├── mysql.go
│   └── db.sql       #数据库文件
├── global/
│   ├── app.go          #全局变量的初始化
│   ├── config.go
│   └── log.go
├── go.mod
├── go.sum
├── initialize/
│   ├── server.go       #启动代理订阅的服务
│   ├── tg.go           #初始化对接telegram-bot
│   └── userHandle.go   #命令行处理用户数据库的相关代码
├── log/
│   ├── gin.log         #gin日志
│   ├── .gitkeep
│   └── run.log         #业务日志
├── main.go             #入口
├── model/
│   ├── users.go
│   ├── system_info.go
│   └── usersJson.go
├── router/
│   └── router.go       #gin路由
├── server
├── service/
│   ├── index.go        #处理数据库的相关代码
│   └── tg.go          #与telegram-bot交互的逻辑
├── static/
│   └── favicon.ico
└── utils/
    └── tool.go
```

## 准备
* 请准备数据库, 点击查看[数据库结构](https://github.com/twbworld/proxy/blob/main/dao/db.sql); 如已安装 [trojan-go](https://github.com/p4gefau1t/trojan-go) 及其 [数据库](https://p4gefau1t.github.io/trojan-go/basic/full-config/#mysql数据库选项),则自行新增余下的表
* 配置数据库有两种方式; 为了与`trojan-go`使用同一个数据库,程序首先会识别config目录下是否存在名为`.trojan-go`的trojan-go配置文件, 如果文件不存在, 则读取相关环境变量, 如下
    ### 环境变量参数
    |  变量值   |  解释  | 默认 |
    |  ----  | ----  | ---- |
    | MYSQL_HOST  | 地址 | "127.0.0.1" |
    | MYSQL_PORT  | 端口 | "3306" |
    | MYSQL_DBNAME  | 库名 | "trojan" |
    | MYSQL_USERNAME  | 用户名 | "root" |
    | MYSQL_PASSWORD  | 密码 | "" |
* 自行创建`telegram-bot`, 将token/id/domain信息配置到`config/.env`
* 配置订阅相关信息: `config/.env`
  > 配置文件下的`trojan`配置`Port`为443时, 默认域名使用cdn, 程序返回的配置使用`WebSocket`协议, 请看[service/index.go](https://github.com/twbworld/proxy/blob/main/service/index.go)代码
* 配置监听端口等信息: `config/trojanGoConfig.go`


## 安装

### docker-compose
``` yaml
version: "3"
services:
    mysql:
        image: mysql:8.0
        environment:
            MYSQL_ALLOW_EMPTY_PASSWORD: true
        volumes:
            - ${PWD}/dao/db.sql:/docker-entrypoint-initdb.d/db.sql:ro
    trojan-go:
        image: p4gefau1t/trojan-go:latest
        depends_on:
            - mysql
        ports:
            - 443:443
        volumes:
            - ${PWD}/trojan-go.json:/etc/trojan-go/config.json:rw

    proxy:
        image: ghcr.io/twbworld/proxy:latest
        container_name: trojan
        depends_on:
        - mysql
        ports:
            - 80:80
        # environment:
        #     MYSQL_HOST: mysql
        #     MYSQL_DBNAME: trojan
        volumes:
            - ${PWD}/config/.env.example:/app/config/.env:ro
            - ${PWD}/trojan-go.json:config/.trojan-go:rw #需要用到trojan-go配置文件下的mysql配置
```

### 打包本地运行
```sh
$ cp config/.env.example config/.env

$ cp trojan-go.json config/.trojan-go

$ go mod tidy && go build -o server main.go

$ ./server
```

## 使用

> 本项目利用了 `GitHub-Actions` 作为 `持续集成` 处理数据, [相关代码](https://github.com/twbworld/proxy/blob/main/.github/workflows/ci.yml), 也可使用命令行, 下面有介绍

### telegram-bot聊天框交互 :
![](https://cdn.jsdelivr.net/gh/twbworld/hosting@main/img/2023081038595.jpg)

### 流量上下行的记录清零
```sh
$ docker exec -it trojan /app/server -a clear
或
$ ./server -a clear
```

### 过期用户处理
```sh
$ docker exec -it trojan /app/server -a expiry
或
$ ./server -a expiry
```


### 客户端订阅
* `trojan-go`和`v2ray`订阅地址例子: `trojan.domain.com/username.html`
* `clash`订阅地址例子: `clash.domain.com/username.html`
    > `clash`与前两者不同, 其识别的是配置文件, 所以clash需不同的网址, 且以clash开头的域名, 请自行解析域名, 而前两者则不要求;[相关代码](https://github.com/twbworld/proxy/blob/main/controller/index.go)
> 提示: 这个客户端使用的`订阅域名`, 跟`连接trojan-go的域名`是不一样哦; 可以理解为: 利用`订阅域名`获取连接信息, 这些连接信息就包含了用于连接trojan-go服务的域名;
