<?php
# 执行这个文件, 更新用户资料
# 可使用CCI/CD执行
# 目前使用GitHub Actions, 位于.github/workflows/cd.yml

error_reporting(E_ALL ^ E_NOTICE);
date_default_timezone_set('Asia/Shanghai');

require 'userHandle.php';
$userHandleInfo = new UserHandle;

/*$shell = 'cd /usr/share/nginx/proxy && git checkout -- . && git pull origin master 2>&1'; //使用2>&1可输出错误信息
$str = exec($shell, $return);
$logStr = json_encode((array) $return, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES);
$userHandleInfo->log(['      拉取代码 :      ' . $logStr]);
exit('oooooooooook :    ' . $logStr);*/

try {
    $userHandleInfo->handle();

} catch (Exception $e) {
    $userHandleInfo::log(['ERROR: ' . $e->getMessage()]);
    // header('HTTP/1.1 404 Not Found');
    exit('Error');
}
