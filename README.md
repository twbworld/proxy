
**Proxy**
===========
[![CI](https://github.com/twbworld/proxy/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/twbworld/proxy/actions/workflows/ci.yaml)
[![](https://img.shields.io/github/tag/twbworld/proxy?logo=github)](https://github.com/twbworld/proxy)
![](https://img.shields.io/badge/language-golang-cyan)
[![](https://img.shields.io/github/license/twbworld/proxy)](https://github.com/twbworld/proxy/blob/main/LICENSE)

## 简介
**翻墙订阅链接服务 和 用户管理**

### 本项目有两个作用:

1. 用户管理(增删改查)
    > 程序作为中间人, 通过与`telegram-bot`进行交互,实现对接用户数据库, 进行用户管理
    > 如喜欢用shell进行用户管理, 建议出门左转使用[Jrohy/trojan](https://github.com/Jrohy/trojan)哟
2. 程序返回客户端可识别的翻墙配置, 即订阅功能
    > 访问含用户名的特定订阅链接, 程序返回`v2ray`和`clash`等客户端可以识别的base64码或`clash`配置文件; 在客户端上配置订阅链接即可翻墙;

## 目录结构 :
``` sh
├── clash.yaml              # Clash 配置模板
├── config.example.yaml     # 配置文件示例 (复制为 config.yaml)
├── controller/             # 控制器 (API 入口)
│   ├── admin/
│   │   └── bot.go
│   ├── enter.go
│   └── user/
│       └── index.go
├── dao/                    # 数据访问层 (SQL 操作)
│   ├── common.go
│   ├── enter.go
│   ├── system_info.go
│   └── users.go
├── Dockerfile
├── .editorconfig
├── .github/
│   └── workflows/
├── global/                 # 全局变量与接口
│   ├── global.go
│   └── interface.go
├── go.mod
├── go.sum
├── initialize/             # 框架与各组件初始化
│   ├── config.go
│   ├── cron.go
│   ├── enter.go
│   ├── log.go
│   ├── server.go           # Gin Web 服务
│   └── services.go
├── internal/               # 外部依赖解耦层 (用于随时无缝切换底层驱动)
│   ├── mysql/              # MySQL 适配与 DDL
│   ├── redis/              # Redis 适配与分布式锁
│   ├── sqlite/             # SQLite 适配与 DDL
│   └── telegram/           # Telegram Bot 通信驱动
├── log/
│   └── .gitkeep
├── main.go                 # 程序入口
├── middleware/             # 路由中间件与验参
│   ├── auth.go
│   └── other.go
├── model/                  # 结构体定义
│   ├── common/
│   ├── config/             # 配置文件结构体
│   ├── db/                 # 数据库模型结构体
│   └── enum/
├── README.md
├── router/
│   └── router.go
├── service/                # 核心业务逻辑
│   ├── admin/
│   │   └── bot.go          # TG Bot 业务逻辑
│   ├── common/
│   │   └── async.go
│   ├── enter.go
│   └── user/
│       └── index.go
├── static/                 # 静态资源
├── task/                   # 后台定时任务
│   ├── cleanup_logs.go     # 日志清理
│   ├── clear.go            # 流量清零
│   ├── enter.go
│   └── expiry.go           # 过期用户处理
└── utils/
    └── tool.go             # 通用工具函数
```

## 准备
* 请准备数据库(默认使用sqlite, 没有db文件则自动在根目录下新增proxy.db文件, 数据库结构参考`initialize/system/db.go`), 并配置`config.yaml` database选项, 可用mysql,参考 `config.example.yaml`
* 自行创建`telegram-bot`, 将token/id/domain信息配置到`config.yaml`, 可实现tg交互管理用户

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
    > `clash`与前者不同,其响应的是yaml格式配置,域名以clash开头;[相关代码](https://github.com/twbworld/proxy/blob/main/controller/user/base.go)
> 提示: 这个客户端使用的`订阅域名`, 跟`连接xray等服务端的域名`是不一样哦; 可以理解为: 利用`订阅域名`获取连接信息, 这些连接信息就包含了用于连接xray服务的域名;
