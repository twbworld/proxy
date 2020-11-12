<?php
$pwd = $_GET['pwd'];
if (!isset($pwd) || empty($pwd) || $pwd !== 'pwd') {
    header('HTTP/1.1 404 Not Found');
    exit();
}

require 'userHandle.php';
$userHandleInfo = new UserHandle();
$shell = 'cd /usr/share/nginx/proxy && git checkout -- . && git pull origin master 2>&1';

try {
    $str = exec($shell,$return);
    $userHandleInfo->handle();

    $userHandleInfo->log(['      拉取代码 :      ' . json_encode((array) $return, JSON_UNESCAPED_UNICODE|JSON_UNESCAPED_SLASHES)]);
    exit('ok');
} catch (Exception $e) {
    $userHandleInfo->log(['ERROR: ' . $e->getMessage()]);
    header('HTTP/1.1 404 Not Found');
    exit();
}
