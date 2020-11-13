<?php
/**
 * 代理订阅
 * @authors 忐忑 (1174865138@qq.com)
 * @date    2020-11-07 17:24:59
 * @version 1.0
 */

class Proxy
{
    private static $trojanDomain = 'tg.twbhub.cf';
    private static $vmessUrl = 'vmess://ew0KICAidiI6ICIyIiwNCiAgInBzIjogIuWklue9keS/oeaBr+Wkjeadgl/nkIbmmbrliIbovqjnnJ/lgYciLA0KICAiYWRkIjogIjQ1Ljc2LjE5NC43OSIsDQogICJwb3J0IjogIjIwODMiLA0KICAiaWQiOiAiZTY2MDk5YjAtNGZmNi00MDBhLWJhOWQtZWQzMjk2MjU0OWRjIiwNCiAgImFpZCI6ICIwIiwNCiAgIm5ldCI6ICJ0Y3AiLA0KICAidHlwZSI6ICJub25lIiwNCiAgImhvc3QiOiAiIiwNCiAgInBhdGgiOiAiIiwNCiAgInRscyI6ICIiDQp9'; #免密码链接,高等级用户使用
    private static $trojanPort = '443';
    private static $usersFile = 'users.json';
    private static $_instance = null;

    private function __construct()
    {
        date_default_timezone_set('Asia/Shanghai');
    }

    public function __clone()
    {
        exit('Clone is not allowed.' . E_USER_ERROR);
    }

    public static function getInstance()
    {
        if (!(self::$_instance instanceof Proxy)) {
            self::$_instance = new Proxy();
        }
        return self::$_instance;
    }

    /**
     * 读取json文件,所有的用户信息
     * @dateTime 2020-11-07T23:36:44+0800
     * @author   twb<1174865138@qq.com>
     * @return   [type]                   [description]
     */
    private static function getUsers()
    {
        $usersData = json_decode(file_get_contents(self::$usersFile), true);
        $usersDataEnable = [];
        if (count($usersData) > 0) {
            array_walk($usersData, function ($value) use (&$usersDataEnable) {
                if ($value['enable'] === true) {
                    $usersDataEnable[] = $value;
                }
            });
        }

        return $usersDataEnable;
    }

    private static function exit() {
        header('HTTP/1.1 404 Not Found');
        exit();
    }

    public static function response()
    {
        $user = trim($_GET['u']);
        if (empty($user) || strlen($user) < 3 || strlen($user) > 15) {
            self::exit();
        }
        $usersData = self::getUsers();
        if (!is_array($usersData) || count($usersData) < 1) {
            self::exit();
        }
        if (count($usersData) > 0) {
            array_walk($usersData, function ($value) use ($user) {
                if ($value['username'] === $user) {
                    // trojan://trojan@tg.twbhub.cf:443?sni=tg.twbhub.cf#外网信息复杂_理智分辨真假
                    // $subscription = 'trojan://' . $value['username'] . '@' . self::$trojanDomain . ':' . self::$trojanPort . '?sni=' . self::$trojanDomain . '#外网信息复杂_理智分辨真假'; //trojan分享链接
                    // trojan://trojan@tg.twbhub.cf:443#外网信息复杂_理智分辨真假
                    $subscription = 'trojan://' . $value['username'] . '@' . self::$trojanDomain . ':' . self::$trojanPort . '#外网信息复杂_理智分辨真假'; //trojan分享链接
                    if (isset($value['level']) && $value['level'] > 0) {
                        $subscription .= PHP_EOL . self::$vmessUrl;
                    }

                    echo base64_encode($subscription);
                    exit();
                }
            });
        }
        self::exit();

    }

}

(Proxy::getInstance())::response();
