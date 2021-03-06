<?php
// 目前使用GitHub Actions执行, 位于.github/workflows/下

error_reporting(E_ALL ^ E_NOTICE);
date_default_timezone_set('Asia/Shanghai');

require __DIR__ . '/../library/UserHandle.php';
$userHandleInfo = new Library\UserHandle();

try {
    $param = trim($argv[1]);
    if (!in_array($param, ['clear', 'expiry', 'handle'])) {
        $userHandleInfo::log(['ERROR: 参数错误' . json_encode((array) $argv)]);
        echo '!!!!!!!!!!!!!!!![' . $param . ']参数错误!!!!!!!!!!!!!!!!!!!!' . PHP_EOL;
        return false;
    }
    $userHandleInfo->$param();
    echo '[' . $param . ']执行成功' . PHP_EOL;
} catch (Exception $e) {
    $userHandleInfo::log(['ERROR: ' . $e->getMessage()]);
    echo '!!!!!!!!!!!!!!!![' . $param . ']执行失败!!!!!!!!!!!!!!!!!!!!' . PHP_EOL;
}
