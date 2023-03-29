<?php
/**
 * 代理订阅
 * @authors 忐忑 (1174865138@qq.com)
 * @date    2020-11-07 17:24:59
 * @version 1.0
 */

namespace Library;

class Subscribe
{
    private static $_instance = null;
    private static $usersPath = __DIR__ . '/../data/users.json';
    private static $envPath = __DIR__ . '/../config/.env';
    private static $clashPath = __DIR__ . '/../config/clash.ini';

    // @codeCoverageIgnoreStart

    private function __construct()
    {

    }

    public function __clone()
    {
        exit('Clone is not allowed.' . E_USER_ERROR);
    }

    public static function getInstance()
    {
        if (!(self::$_instance instanceof Subscribe)) {
            self::$_instance = new Subscribe();
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

    function exit() {
        if ($_ENV['phpunit'] === '1') {
            throw new \Exception("出现错误");
        } else {
            header('HTTP/1.1 404 Not Found');
            exit();
        }
    }

    // @codeCoverageIgnoreEnd

    /**
     * 读取所有的用户信息
     * @dateTime 2020-11-07T23:36:44+0800
     * @author   twb<1174865138@qq.com>
     * @return   [type]                   [description]
     */
    public static function getUsers()
    {

        $usersData = self::loadJsonFile(self::$usersPath);
        $usersDataEnable = [];
        if (is_array($usersData) && count($usersData) > 0) {
            array_walk($usersData, function ($value) use (&$usersDataEnable) {
                if ($value['enable'] === true) {
                    $usersDataEnable[] = $value;
                }
            });
        }

        return $usersDataEnable;
    }

    public function response()
    {
        $user = trim($_GET['u']);
        if (empty($user) || strlen($user) < 3 || strlen($user) > 50) {
            self::exit();
        }
        $usersData = self::getUsers();
        if (!is_array($usersData) || count($usersData) < 1) {
            self::exit();
        }

        foreach ($usersData as $value) {
            if (empty($value['username']) || $value['username'] !== $user) {
                continue;
            }
            $subscription = '';
            $env = self::loadJsonFile(self::$envPath);
            $isClash = 'clash' == substr($_SERVER['SERVER_NAME'], 0, 4); //判断三级域名
            $clashIni = $isClash ? file_get_contents(self::$clashPath) : '';
            '' == $clashIni && $isClash = false;
            if (is_array($env['trojan']) && count($env['trojan']) > 0) {
                foreach ($env['trojan'] as $val) {
                    if (empty($val['domain'])) {
                        continue;
                    }

                    if (isset($val['port']) && $val['port'] != '443') {
                        //直连分享链接(trojan)
                        if ($isClash) {
                            $wsConfig = '';
                            $subscription .= str_replace(['%username%', '%domain%', '%port%', '%ws%'], [$value['username'], $val['domain'], $val['port'], $wsConfig], $clashIni);
                        }else {
                            $subscription .= 'trojan://' . $value['username'] . '@' . $val['domain'] . ':' . $val['port'] . '?security=tls&headerType=none&fp=chrome&uTLS=chrome&mux=1&alpn=h2,http/1.1&type=tcp&sni=' . $val['domain'] . '#外网信息复杂_理智分辨真假' . PHP_EOL;
                        }



                    }else {
                        //cdn分享链接(trojan)
                        if ($isClash) {
                            $wsConfig = ', network: ws, ws-opts: {path: /trojan-go-ws/, headers: {Host: ' . $val['domain'] . '}}';
                            $subscription .= str_replace(['%username%', '%domain%', '%port%', '%ws%'], [$value['username'], $val['domain'], $val['port'], $wsConfig], $clashIni);
                        }else {
                            $subscription .= 'trojan://' . $value['username'] . '@' . $val['domain'] . ':443?security=tls&headerType=none&fp=chrome&uTLS=chrome&mux=1&type=ws&path=/trojan-go-ws/&host=' . $val['domain'] . '&sni=' . $val['domain'] . '#外网信息复杂_理智分辨真假' . PHP_EOL;
                        }
                    }


                }
            }
            if (!$isClash && isset($value['level']) && $value['level'] > 0 && is_array($env['superUrl']) && count($env['superUrl']) > 0) {
                $subscription .= implode(PHP_EOL, $env['superUrl']); //其他分享链接,vmess
            }

            return $isClash ? $subscription : base64_encode(trim($subscription, PHP_EOL));

        }

        self::exit();

    }

}
