#!/bin/bash

# 清除流量上下行的记录
# 安装crontab后, 把这个文件放到"/etc/cron.monthly/"下,可每月执行一次
php -r "require '/usr/share/nginx/proxy/userHandle.php';(UserHandle::getInstance())::clear();exit();"
