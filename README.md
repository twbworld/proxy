
**Proxy**
===========
[![](https://github.com/twbworld/proxy/workflows/ci/badge.svg?branch=main)](https://github.com/twbworld/proxy/actions)
[![](https://img.shields.io/github/tag/twbworld/proxy?logo=github)](https://github.com/twbworld/proxy)
![](https://img.shields.io/badge/language-golang-cyan)
[![](https://img.shields.io/github/license/twbworld/proxy)](https://github.com/twbworld/proxy/blob/main/LICENSE)

### 简介
翻墙代理服务器的订阅和用户管理 控制代码 ; 可作为订阅服务器 , 以及 , 通过 `GitHub-Actions` 作为 `持续集成` , 自动化更新数据库(安装 ***`Jrohy/trojan`*** 作为前提)
> 推荐在Docker内使用

### 准备
基于 [Jrohy/trojan](https://github.com/Jrohy/trojan) 控制面板, 首先进行安装
``` sh
$ wget -N --no-check-certificate -q -O install_trojan_go.sh "https://git.io/trojan-install" && chmod +x install_trojan_go.sh && ./install_trojan_go.sh
```

### 目录结构 : 
``` sh
├── .editorconfig
├── .gitattributes
├── .github/
│   └── workflows/          #存放GitHub-Actions的工作流文件
│       ├── ci.yml
│       ├── clear.yml
│       └── expiry.yml
├── .gitignore
├── LICENSE
├── README.md
├── config/
│   ├── appConfig.go
│   ├── appConfig.json  #项目配置
│   ├── clash.ini       #clash配置模板
│   ├── config.go
│   ├── .env            #机密配置文件,数据库之类的
│   ├── .env.example    #配置模板
│   └── env.go
├── controller/         #MVC模式的C
│   └── index.go
├── dao/
│   ├── db_handle.go    #用户数据操作
│   ├── mysql.go
│   ├── users.json      #用户,要同步到数据库
│   └── users.sql       #数据库文件
├── global/
│   ├── app.go          #全局变量的初始化
│   ├── config.go
│   └── log.go
├── go.mod
├── go.sum
├── initialize/
│   ├── server.go       #启动代理订阅的服务
│   └── userHandle.go   #命令行处理用户数据库的相关代码
├── log/
│   ├── gin.log
│   ├── .gitkeep
│   └── run.log
├── main.go             #入口
├── model/
│   ├── users.go
│   └── usersJson.go
├── router/
│   └── router.go       #gin路由
├── server
├── service/
│   ├── index.go        #处理数据库的相关代码
│   └── msg.go          #发送通知代码
├── static/
│   └── favicon.ico
└── utils/
    └── tool.go
```
### 使用
**自行 clone项目 和 安装go环境 和 安装`Jrohy/trojan`项目**

**连接配置`Jrohy/trojan`数据库: config/.env**
```sh
$ cd src && go mod tidy && go build -o server main.go
```

#### 运行订阅服务
```sh
$ ./server
```

> 配置监听端口: config/appConfig.json  
> 订阅地址例子: `www.domain.com/u=username`

#### 根据dao/users.json更新数据库用户信息
```sh
$ ./server -a handle
```

#### 流量上下行的记录清零
```sh
$ ./server -a clear
```

#### 过期用户处理
```sh
$ ./server -a expiry
```

#### 持续集成

利用 `GitHub-Actions`, 可以作为参考
