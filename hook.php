<?php
$pwd = $_GET['pwd'];
if (!isset($pwd) || empty($pwd) || $pwd !== 'pwd') {
    header('HTTP/1.1 404 Not Found');
    exit();
}

require 'userHandle.php';
$userHandleInfo = new UserHandle;
$shell = 'cd /usr/share/nginx/proxy && git checkout -- . && git pull origin master 2>&1'; //使用2>&1可输出错误信息

try {
    $str = exec($shell, $return);
    $userHandleInfo->handle();

    $logStr = json_encode((array) $return, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES);
    $userHandleInfo->log(['      拉取代码 :      ' . $logStr]);
    exit('oooooooooook :    ' . $logStr);
} catch (Exception $e) {
    $userHandleInfo->log(['ERROR: ' . $e->getMessage()]);
    header('HTTP/1.1 404 Not Found');
    exit();
}
