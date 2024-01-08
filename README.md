
**Proxy**
===========
[![](https://github.com/twbworld/proxy/workflows/ci/badge.svg?branch=main)](https://github.com/twbworld/proxy/actions)
[![](https://img.shields.io/github/tag/twbworld/proxy?logo=github)](https://github.com/twbworld/proxy)
![](https://img.shields.io/badge/language-golang-cyan)
[![](https://img.shields.io/github/license/twbworld/proxy)](https://github.com/twbworld/proxy/blob/main/LICENSE)

## 简介
**翻墙订阅链接服务 和 用户管理**

### 本项目有两个作用:

1. 用户管理(增删改查)
    > 程序作为中间人, 通过与`telegram-bot`进行交互,实现对接用户数据库, 进行用户管理  
    > 如喜欢用文件而不是`telegram-bot`进行用户管理, 请使用`v0`版本  
    > 如喜欢用shell进行用户管理, 建议出门左转使用[Jrohy/trojan](https://github.com/Jrohy/trojan)哟
2. 程序返回客户端可识别的翻墙配置, 即订阅功能
    > 访问含用户名的特定订阅链接, 程序返回`v2ray`和`clash`等客户端可以识别的base64码或`clash`配置文件; 在客户端上配置订阅链接即可翻墙;

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
│   ├── config.go
│   ├── .env            #机密配置文件,数据库之类的
│   ├── .env.example    #配置模板
│   └── env.go
├── controller/         #MVC模式的C
│   └── index.go
├── dao/
│   ├── db.go          #数据库初始化
│   └── dbHandle.go    #用户数据操作
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
│   └── systemInfo.go
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
* 请准备数据库(默认使用sqlite, 没有db文件则自动在dao目录下新增proxy.db文件, 数据库结构参考`dao/db.go`), 并配置`config/.env` db选项, 可用mysql, 甚至直接配置trojan-go的mysql数据库, 参考 `config/.env.example`
* 自行创建`telegram-bot`, 将token/id/domain信息配置到`config/.env`, 可实现tg交互管理用户
* 配置监听端口(默认80)等信息: `config/appConfig.go`
* 建议使用的xray很难做用户管理, 故项目不依赖其数据而是外置数据库, 缺点是不能利用xray的流量统计功能(使用trojan-go和其数据库, 则流量统计可用); 未来会对接xray数据, 也许吧!
* 未来支持环境变量配置, 也许吧!

## 安装

### docker-compose
``` yaml
version: "3"
services:
    proxy:
        image: ghcr.io/twbworld/proxy:latest
        ports:
            - 80:80
        volumes:
            - ${PWD}/config/.env:/app/config/.env:ro
            - ${PWD}/dao/proxy.db:dao/proxy.db:rw
```

### 打包本地运行
```sh
$ cp config/.env.example config/.env

$ go mod tidy && go build -o server main.go

$ ./server
```

## 使用

> 本项目利用了 `GitHub-Actions` 作为 `持续集成` 处理数据, [相关代码](https://github.com/twbworld/proxy/blob/main/.github/workflows/ci.yml), 也可使用命令行, 下面有介绍

### telegram-bot聊天框交互 :
![](https://cdn.jsdelivr.net/gh/twbworld/hosting@main/img/2023081038595.jpg)

### 流量上下行的记录清零
```sh
$ docker exec -it proxy /app/server -a clear
或
$ ./server -a clear
```

### 过期用户处理
```sh
$ docker exec -it proxy /app/server -a expiry
或
$ ./server -a expiry
```


### 客户端订阅
* `v2ray`订阅地址例子: `www.domain.com/username.html`
  > 其中`www.domain.com`是自己的域名,指向该项目监听的端口; `username`是用户名, 如果数据库中存在该用户, 则显示在`config/.env`下`proxy`选项所配置的vpn信息
* `clash`订阅地址例子: `clash.domain.com/username.html`
    > `clash`与前两者不同, 其识别的是配置文件, 所以clash需不同的网址, 且以clash开头的域名, 请自行解析域名;[相关代码](https://github.com/twbworld/proxy/blob/main/controller/index.go)
> 提示: 这个客户端使用的`订阅域名`, 跟`连接xray等服务端的域名`是不一样哦; 可以理解为: 利用`订阅域名`获取连接信息, 这些连接信息就包含了用于连接xray服务的域名;
