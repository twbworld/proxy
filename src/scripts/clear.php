<?php
# 执行这个文件, 用户流量清零
# 目前使用GitHub Actions定时执行, 位于.github/workflows/cron.yml

error_reporting(E_ALL ^ E_NOTICE);
date_default_timezone_set('Asia/Shanghai');

require __DIR__ . '/../library/UserHandle.php';
$userHandleInfo = new Library\UserHandle();

try {
    $userHandleInfo->clear();
    echo '流量清零 成功' . PHP_EOL;
} catch (Exception $e) {
    $userHandleInfo::log(['ERROR: ' . $e->getMessage()]);
    echo '!!!!!!!!!!!!!!!!!!!!!!!!!!流量清零 失败!!!!!!!!!!!!!!!!!1' . PHP_EOL;
}
