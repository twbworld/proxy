
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
├── clash.yaml              #clash配置模板
├── config.example.yaml     #以其为例, 自行创建config.yaml
├── controller/
│   ├── admin/              #后台api
│   ├── enter.go
│   └── user/               #前台api
├── dao/                    #sql
├── Dockerfile              #构建docker镜像
├── .editorconfig
├── .gitattributes
├── .github/
│   └── workflows/          #存放GitHub-Actions的工作流文件
├── .gitignore
├── .gitmodules
├── global/
│   └── global.go           #全局变量的初始化
├── go.mod
├── go.sum
├── initialize/             #服务初始化相关
│   ├── server.go           #gin服务
│   ├── global/
│   └── system/
├── LICENSE
├── log/
│   ├── gin.log             #gin日志
│   ├── .gitkeep
│   └── run.log             #业务日志
├── main.go                 #入口
├── main_test.go            #测试
├── middleware/             #路由中间件以及验参
├── model/
│   ├── common/             #业务要用的结构体
│   ├── config/             #配置文件的结构体
│   └── db/                 #数据库模型结构体
├── README.md
├── router/                 #gin路由
├── service/
│   └── user/
│       ├── enter.go
│       ├── index.go        #处理数据库的相关代码
│       ├── tg.go           #与telegram-bot交互的逻辑
│       └── validator.go
├── static/                 #静态资源
├── task/                   #任务
│   ├── clear.go            #流量清零
│   └── expiry.go           #过期用户处理
└── utils/
    └── tool.go
```

## 准备
* 请准备数据库(默认使用sqlite, 没有db文件则自动在根目录下新增proxy.db文件, 数据库结构参考`initialize/system/db.go`), 并配置`config.yaml` db选项, 可用mysql, 甚至直接配置trojan-go的mysql数据库, 参考 `config.example.yaml`
* 自行创建`telegram-bot`, 将token/id/domain信息配置到`config.yaml`, 可实现tg交互管理用户
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
            - ${PWD}/config.yaml:/app/config.yaml:ro
            - ${PWD}/proxy.db:proxy.db:rw
```

### 打包本地运行
```sh
$ cp config.example.yaml config.yaml

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
  > 其中`www.domain.com`是自己的域名,指向该项目监听的端口; `username`是用户名, 如果数据库中存在该用户, 则显示在`config.yaml`下`proxy`选项所配置的vpn信息
* `clash`订阅地址例子: `clash.domain.com/username.html`
    > `clash`与前两者不同, 其识别的是配置文件, 所以clash需不同的网址, 且以clash开头的域名, 请自行解析域名;[相关代码](https://github.com/twbworld/proxy/blob/main/controller/user/base.go)
> 提示: 这个客户端使用的`订阅域名`, 跟`连接xray等服务端的域名`是不一样哦; 可以理解为: 利用`订阅域名`获取连接信息, 这些连接信息就包含了用于连接xray服务的域名;
