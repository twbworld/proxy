<?php
/**
 * 代理订阅
 * @authors 忐忑 (1174865138@qq.com)
 * @date    2020-11-07 17:24:59
 * @version 1.0
 */

class Proxy
{
    private static $_instance = null;
    private static $usersPath = './users.json';
    private static $envPath = './.env';

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

    private static function loadJsonFile($path)
    {
        if (!file_exists($path)) {
            throw new \Exception('配置文件' . $path . '不存在');
        }
        $arr = json_decode(file_get_contents($path), true);
        return json_last_error() == JSON_ERROR_NONE ? $arr : [];
    }

    /**
     * 读取所有的用户信息
     * @dateTime 2020-11-07T23:36:44+0800
     * @author   twb<1174865138@qq.com>
     * @return   [type]                   [description]
     */
    private static function getUsers()
    {

        $usersData = self::loadJsonFile(self::$usersPath);
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

    /**
     * 读取配置
     * @dateTime 2020-11-07T23:36:44+0800
     * @author   twb<1174865138@qq.com>
     * @return   [type]                   [description]
     */
    private static function getEnv()
    {
        return self::loadJsonFile(self::$envPath);
    }

    function exit() {
        header('HTTP/1.1 404 Not Found');
        exit();
    }

    public function response()
    {
        $user = trim($_GET['u']);
        if (empty($user) || strlen($user) < 3 || strlen($user) > 15) {
            self::exit();
        }
        $usersData = self::getUsers();
        if (!is_array($usersData) || count($usersData) < 1) {
            self::exit();
        }

        array_walk($usersData, function ($value) use ($user) {
            if (!empty($value['username']) && $value['username'] === $user) {
                $subscription = '';
                $env = self::getEnv();
                if (count($env['trojan']) > 0) {
                    array_walk($env['trojan'], function ($val) use (&$subscription, $value) {
                        // trojan://trojan@www.trojanDomain.com:443?sni=www.trojanDomain.com#外网信息复杂_理智分辨真假
                        // trojan://trojan@www.trojanDomain.com:443#外网信息复杂_理智分辨真假
                        if (!empty($val['domain'])) {
                            $subscription .= 'trojan://' . $value['username'] . '@' . $val['domain'] . ':' . ($val['port'] ?? '443') . '#外网信息复杂_理智分辨真假' . PHP_EOL; //trojan分享链接
                        }
                    });
                }
                if (isset($value['level']) && $value['level'] > 0 && count($env['superUrl']) > 0) {
                    $subscription .= implode(PHP_EOL, $env['superUrl']); //其他分享链接,vmess
                }

                echo base64_encode(trim($subscription, PHP_EOL));
                exit();
            }
        });

    }

}

(Proxy::getInstance())->response();
