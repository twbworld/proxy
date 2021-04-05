
**Proxy**
===========
[![](https://github.com/twbworld/proxy/workflows/ci/badge.svg?branch=master)](https://github.com/twbworld/proxy/actions)
[![](https://img.shields.io/github/tag/twbworld/proxy?logo=github)](https://github.com/twbworld/proxy)
![](https://img.shields.io/badge/language-PHP-orange)
[![](https://img.shields.io/github/license/twbworld/proxy)](https://github.com/twbworld/proxy/blob/master/LICENSE)
[![codecov](https://codecov.io/gh/twbworld/proxy/branch/master/graph/badge.svg?token=08N3AJSVCR)](https://codecov.io/gh/twbworld/proxy)
[![Codacy Badge](https://app.codacy.com/project/badge/Coverage/d6a59fd45f234d8aab947e5be874f5cd?branch=master)](https://www.codacy.com/gh/twbworld/test/dashboard?utm_source=github.com&utm_medium=referral&utm_content=twbworld/proxy&utm_campaign=Badge_Coverage)
[![Codacy Badge](https://app.codacy.com/project/badge/Grade/d6a59fd45f234d8aab947e5be874f5cd)](https://www.codacy.com/gh/twbworld/proxy/dashboard?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=twbworld/proxy&amp;utm_campaign=Badge_Grade)

### 简介
翻墙代理服务器的订阅和用户管理 控制代码 ; 可作为订阅服务器 , 以及 , 通过 `GitHub-Actions` 作为 `持续集成` , 自动化更新数据库(安装 ***`Jrohy/trojan`*** 作为前提)
> 推荐在Docker内使用

### 准备
基于 [Jrohy/trojan](https://github.com/Jrohy/trojan) 控制面板, 首先进行安装
```
$ wget -N --no-check-certificate -q -O install_trojan_go.sh "https://git.io/trojan-install" && chmod +x install_trojan_go.sh && ./install_trojan_go.sh
```

### 目录结构 : 
``` sh
├── phpunit.xml              #单元测试配置
├── tests/                   #单元测试目录
├── .github/
│    └── workflows/          #存放GitHub-Actions的工作流文件
├── src/
     ├── config/
     │   ├── .env            #配置文件,数据库啥的
     │   └── .env.example    #配置例子
     ├── data/
     │   ├── users.json      #用户,要同步到数据库
     │   └── users.sql       #数据库文件
     ├── library/
     │   ├── Subscribe.php   #代理订阅的相关代码
     │   └── UserHandle.php  #处理数据库的相关代码
     ├── logs/
     │   ├── .gitkeep
     │   └── userHandle.log
     ├── public/
     │   └── index.php        #订阅入口
     └── scripts/
         └── bash.php       #脚本入口
```
### 使用
利用 `GitHub-Actions` 作为 `持续集成` , 位于 `.github/workflows` 下 , 可以作为参考


### 单元测试(如有需要)

使用 `PHPunit` 工具, 首先 `composer` 安装依赖
``` sh
$ composer install
```
在 `tests` 下编写单元测试代码后, 执行
``` sh
$ ./vendor/bin/phpunit
```

