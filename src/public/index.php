<?php

error_reporting(E_ALL ^ E_NOTICE);
date_default_timezone_set('Asia/Shanghai');

require __DIR__ . '/../library/Subscribe.php';

$subscribe = Library\Subscribe::getInstance();
echo $subscribe->response();
exit;
