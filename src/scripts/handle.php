<?php
# 执行这个文件, 更新用户资料
# 目前使用GitHub Actions执行, 位于.github/workflows/cd.yml

error_reporting(E_ALL ^ E_NOTICE);
date_default_timezone_set('Asia/Shanghai');

require __DIR__ . '/../library/UserHandle.php';
$userHandleInfo = new Library\UserHandle();

try {
    $userHandleInfo->handle();
    echo '用户处理 成功' . PHP_EOL;
} catch (Exception $e) {
    $userHandleInfo::log(['ERROR: ' . $e->getMessage()]);
    echo '!!!!!!!!!!!!!!!!!!!!!!!!!!用户处理 失败!!!!!!!!!!!!!!!!!1' . PHP_EOL;
}
