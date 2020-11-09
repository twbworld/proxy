<?php
$pwd = $_GET['pwd'];
if (!isset($pwd) || empty($pwd) || $pwd !== 'pwd') {
    header('HTTP/1.1 404 Not Found');
    exit();
}

file_put_contents('test.txt', PHP_EOL . 'all:   ' . file_get_contents('php://input'), FILE_APPEND | LOCK_EX);
file_put_contents('test.txt', PHP_EOL . 'get:   ' . json_encode((array) $_GET), FILE_APPEND | LOCK_EX);
file_put_contents('test.txt', PHP_EOL . 'post:   ' . json_encode((array)$_POST), FILE_APPEND | LOCK_EX);

require 'userHandle.php';
$userHandleInfo = new UserHandle();
$shell = 'cd /usr/share/nginx/proxy && git pull origin master'; //这里要提前人工输入"yes"(git要确认私钥)

try {
    $str = shell_exec($shell);
    $userHandleInfo->handle();
    exit('ok');
} catch (Exception $e) {
    $userHandleInfo->log([$e->getMessage()]);
    header('HTTP/1.1 404 Not Found');
    exit();
}
