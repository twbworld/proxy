<?php
namespace Library;

// @codeCoverageIgnoreStart

abstract class db
{
    abstract public static function getInstance();
    abstract public function beginTransaction();
    abstract public function commit();
    abstract public function add($sql, $param);
    abstract public function del($sql, $param);
    abstract public function update($sql, $param);
    abstract public function select($sql);
}

class Users extends db
{
    private static $dbname = 'trojan';
    private static $host = 'mysql';
    private static $username = 'root';
    private static $password = 'tp';
    private static $db = null;
    private static $_instance = null;

    private function __construct()
    {
        $dsn = 'mysql:dbname=' . self::$dbname . ';host=' . self::$host;

        try {
            self::$db = new \PDO($dsn, self::$username, self::$password);
            self::$db->exec('set names utf8');
        } catch (\PDOException $e) {
            echo 'Connection failed: ' . $e->getMessage();
        }
    }

    public function __clone()
    {
        exit('Clone is not allowed.' . E_USER_ERROR);
    }

    public static function getInstance()
    {
        if (!(self::$_instance instanceof Users)) {
            self::$_instance = new Users();
        }
        return self::$_instance;
    }

    public function beginTransaction()
    {
        $sth = self::$db->beginTransaction;
    }

    public function commit()
    {
        $sth = self::$db->commit;
    }

    public function add($sql, $param)
    {
        $sth = self::$db->prepare($sql);
        return $sth->execute($param);
    }
    public function del($sql, $param)
    {
        $sth = self::$db->prepare($sql);
        return $sth->execute($param);
    }
    public function update($sql, $param)
    {
        $sth = self::$db->prepare($sql);
        return $sth->execute($param);
    }
    public function select($sql)
    {
        $sth = self::$db->query($sql);
        return $sth->fetchAll(\PDO::FETCH_ASSOC);
    }
}

interface Factory
{
    public static function beginTransaction();
    public static function commit();
    public static function addUser($value);
    public static function delUser($id);
    public static function updateUser($value);
    public static function selectUser();
    public static function clear();
}

class UsersDbHandle implements Factory
{
    public static function beginTransaction()
    {
        (Users::getInstance())->beginTransaction();
    }
    public static function commit()
    {
        (Users::getInstance())->commit();
    }
    public static function addUser($value)
    {
        $sql = 'INSERT INTO `users`(`username`, `password`, `passwordShow`, `quota`, `download`, `upload`, `useDays`, `expiryDate`)
                VALUES(:username, :password, :passwordShow, :quota, :download, :upload, :useDays, :expiryDate)';
        return (Users::getInstance())->add($sql, [
            'username' => $value['username'],
            'password' => $value['password'],
            'passwordShow' => $value['passwordShow'],
            'quota' => $value['quota'] ?? -1,
            'expiryDate' => $value['expiryDate'], //2021-02-02
            'useDays' => 30,
            'download' => 0,
            'upload' => 0,
        ]);
    }

    public static function delUser($id)
    {
        $sql = 'DELETE FROM `users` where `id` = :id';
        return (Users::getInstance())->del($sql, [
            'id' => $id,
        ]);
    }
    public static function updateUser($value)
    {
        $sql = 'SELECT `username` FROM `users` WHERE `id` = ' . $value['id'] . ' FOR UPDATE';
        (Users::getInstance())->select($sql);

        $sql = 'UPDATE `users` SET `password` = :password, `passwordShow` = :passwordShow, `quota` = :quota, `expiryDate` = :expiryDate WHERE `id` = :id';
        return (Users::getInstance())->update($sql, [
            'id' => $value['id'],
            'password' => $value['password'],
            'passwordShow' => $value['passwordShow'],
            'quota' => $value['quota'] ?? -1,
            'expiryDate' => $value['expiryDate']
        ]);
    }
    public static function selectUser()
    {
        $sql = 'SELECT `id`, `username`, `passwordShow`, `quota`, `useDays`, `expiryDate` FROM `users`';
        return (Users::getInstance())->select($sql);
    }

    public static function clear()
    {
        $sql = 'UPDATE `users` SET `download` = :download, `upload` = :upload';
        return (Users::getInstance())->update($sql, [
            'download' => 0,
            'upload' => 0,
        ]);
    }

}

// @codeCoverageIgnoreEnd

/**
 * 处理用户信息,对mysql数据库作处理
 */
class UserHandle
{

    private static $usersFile = __DIR__ . '/../data/users.json';
    private static $logFile = __DIR__ . '/../logs/userHandle.log';
    private static $quotaMax = '1073741824'; //入库需要, 1G*1024*1024*1024 = 1073741824byte

    private static function getUsersByJson()
    {
        $usersData = json_decode(file_get_contents(self::$usersFile), true);
        if (!is_array($usersData)) {
            self::log(['ERROR: 会员json数据错误']);
            exit;
        }
        $usersDataEnable = [];
        if (count($usersData) > 0) {
            array_walk($usersData, function ($value) use (&$usersDataEnable) {
                if (!isset($value['username']) || strlen($value['username']) < 3 || strlen($value['username']) > 15 || !isset($value['password']) || strlen($value['password']) < 3 || strlen($value['password']) > 15 || !isset($value['quota']) || !isset($value['enable']) || !isset($value['level']) || !isset($value['expiryDate']) || (!empty($value['expiryDate']) && strtotime(date('Y-m-d', strtotime($value['expiryDate']))) !== strtotime($value['expiryDate']))) {
                    self::log(['ERROR: 会员json数据错误']);
                    exit;
                }
                if ($value['enable'] === true) {
                    $value['quota'] > 0 && ($value['quota'] = $value['quota'] * self::$quotaMax);
                    $value['passwordShow'] = self::base64($value['password']);
                    $value['password'] = self::hash($value['password']);
                    $value['expiryDate'] = trim($value['expiryDate']);
                    $usersDataEnable[] = $value;
                }
            });
        }

        return $usersDataEnable;
    }

    private static function base64($str)
    {
        return base64_encode($str);
    }

    private static function hash($str)
    {
        return hash('sha224', $str);
    }

    /**
     * 更新用户表
     * @dateTime 2020-11-11T23:47:23+0800
     * @author   twb<1174865138@qq.com>
     * @return   [type]                   [description]
     */
    public function handle()
    {
        date_default_timezone_set('Asia/Shanghai');

        $usersJson = self::getUsersByJson();
        $usersMysql = UsersDbHandle::selectUser();
        $usersMysqlNew = $userIsset = $log = [];

        if (count($usersMysql) > 0) {
            array_walk($usersMysql, function ($value) use (&$usersMysqlNew) {
                $usersMysqlNew[$value['username']] = $value;
            });
        }

        UsersDbHandle::beginTransaction();

        if (count($usersJson) > 0) {
            array_walk($usersJson, function ($value) use ($usersMysqlNew, &$userIsset, &$log) {
                if (isset($usersMysqlNew[$value['username']])) {
                    $userIsset[] = $value['username'];
                    if ($usersMysqlNew[$value['username']]['passwordShow'] !== $value['passwordShow'] || $usersMysqlNew[$value['username']]['quota'] != $value['quota'] || $usersMysqlNew[$value['username']]['expiryDate'] != $value['expiryDate']) {
                        $value['id'] = $usersMysqlNew[$value['username']]['id'];
                        UsersDbHandle::updateUser($value); //改
                        $log[] = 'update: ' . json_encode($usersMysqlNew[$value['username']], JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES) . ' => ' . json_encode($value, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES);
                    }
                } else {
                    UsersDbHandle::addUser($value); //增
                    $log[] = 'add: ' . json_encode($value, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES);
                }
            });
        }

        $userDiff = array_diff(array_keys($usersMysqlNew), $userIsset);
        if (count($userDiff) > 0) {
            array_walk($userDiff, function ($value) use ($usersMysqlNew, &$log) {
                UsersDbHandle::delUser($usersMysqlNew[$value]['id']); //删
                $log[] = 'del: ' . json_encode($usersMysqlNew[$value], JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES);
            });
        }

        UsersDbHandle::commit();

        self::log($log);

    }

    /**
     * 清除流量上下行的记录
     * @dateTime 2020-11-11T23:46:04+0800
     * @author   twb<1174865138@qq.com>
     */
    public function clear()
    {
        UsersDbHandle::clear();
        self::log(['!!!!!!!!!!!!!!!!!!!!! Clear: 流量清零 !!!!!!!!!!!!!!!!!!!!!!']);
        echo '流量清零完成' . PHP_EOL;
    }

    public static function log($arr)
    {
        if (!is_array($arr)) {
            $log = PHP_EOL . date('Y-m-d H:i:s', time()) . '   ' . (string) $arr;
            file_put_contents(self::$logFile, $log, FILE_APPEND | LOCK_EX);
        } elseif (count($arr) > 0) {
            array_walk($arr, function ($value) {
                $log = PHP_EOL . date('Y-m-d H:i:s', time()) . '   ' . $value;
                file_put_contents(self::$logFile, $log, FILE_APPEND | LOCK_EX);
            });
        }
    }

}

// (new Library/UserHandle)->handle();
