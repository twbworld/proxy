<?php
class Env
{
    private static $path = './.env';
    private static $prefix = 'PHP_';
    private static $_instance = null;

    private function __construct()
    {
        date_default_timezone_set('Asia/Shanghai');
        self::loadFile();
    }

    public function __clone()
    {
        exit('Clone is not allowed.' . E_USER_ERROR);
    }

    public static function getInstance()
    {
        if (!(self::$_instance instanceof Env)) {
            self::$_instance = new Env();
        }
        return self::$_instance;
    }

    /**
     * 加载配置文件
     * @dateTime 2021-01-18T14:01:57+0800
     * @author   twb<1174865138@qq.com>
     * @return   [type]                   [description]
     */
    private static function loadFile()
    {
        if (!file_exists(self::$path)) {
            throw new \Exception('配置文件' . self::$path . '不存在');
        }

        //返回二位数组
        $env = parse_ini_file(self::$path, true);
        foreach ($env as $key => $val) {
            $prefix = self::$prefix . strtoupper($key);
            if (is_array($val)) {
                foreach ($val as $k => $v) {
                    $item = $prefix . '_' . strtoupper($k);
                    putenv("$item=$v");
                }
            } else {
                putenv("$prefix=$val");
            }
        }
    }

    /**
     * 获取环境变量值
     * @dateTime 2021-01-18T14:06:11+0800
     * @author   twb<1174865138@qq.com>
     * @param    string                   $name    [description]
     * @param    [type]                   $default [description]
     * @return   [type]                            [description]
     */
    public function get(string $name, $default = null)
    {
        $result = getenv(self::$prefix . strtoupper(str_replace('.', '_', $name)));

        if (false !== $result) {
            if ('false' === $result) {
                $result = false;
            } elseif ('true' === $result) {
                $result = true;
            }
            return $result;
        }
        return $default;
    }

}
