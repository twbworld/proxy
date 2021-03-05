<?php
namespace Library;

// @codeCoverageIgnoreStart

abstract class db
{
    abstract public static function getInstance();
    abstract public function beginTransaction();
    abstract public function commit();
    abstract public function rollBack();
    abstract public function add($sql, $param);
    abstract public function del($sql, $param);
    abstract public function update($sql, $param);
    abstract public function select($sql);
    abstract public function expiry($sql);
}

class Users extends db
{
    private static $envPath = __DIR__ . '/../config/.env';
    private static $db = null;
    private static $_instance = null;

    private function __construct()
    {
        if (!file_exists(self::$envPath)) {
            throw new \Exception('配置文件' . self::$envPath . '不存在');
        }
        $dbConfig = json_decode(file_get_contents(self::$envPath), true);
        (json_last_error() != JSON_ERROR_NONE || !isset($dbConfig['mysqlConfig'])) && $dbConfig = [];

        $dsn = 'mysql:dbname=' . $dbConfig['mysqlConfig']['dbname'] . ';host=' . $dbConfig['mysqlConfig']['host'];

        try {
            self::$db = new \PDO($dsn, $dbConfig['mysqlConfig']['username'], $dbConfig['mysqlConfig']['password']);
            self::$db->exec('set names utf8');
        } catch (\PDOException $e) {
            echo '数据库连接失败丫 : ' . $e->getMessage();
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
        self::$db->beginTransaction();
    }

    public function commit()
    {
        self::$db->commit();
    }

    public function rollBack()
    {
        self::$db->rollBack();
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
    public function expiry($sql)
    {
        $sth = self::$db->prepare($sql);
        return $sth->execute($param);
    }
}

interface Factory
{
    public static function beginTransaction();
    public static function commit();
    public static function rollBack();
    public static function addUser($value);
    public static function delUser($id);
    public static function updateUser($value);
    public static function selectUser($where);
    public static function clear();
    public static function expiry($id);
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
    public static function rollBack()
    {
        (Users::getInstance())->rollBack();
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
            'useDays' => $value['useDays'], //天数,0为无限制
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

        $sql = 'UPDATE `users` SET `password` = :password, `passwordShow` = :passwordShow, `quota` = :quota, `useDays` = :useDays, `expiryDate` = :expiryDate WHERE `id` = :id';
        return (Users::getInstance())->update($sql, [
            'id' => $value['id'],
            'password' => $value['password'],
            'passwordShow' => $value['passwordShow'],
            'quota' => $value['quota'] ?? -1,
            'expiryDate' => $value['expiryDate'],
            'useDays' => $value['useDays']
        ]);
    }
    public static function selectUser($where = '')
    {
        $sql = 'SELECT `id`, `username`, `passwordShow`, `quota`, `useDays`, `expiryDate` FROM `users`';
        !empty($where) && $sql .= ' WHERE ' . trim($where);
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

    public static function expiry($id)
    {
        $sql = 'UPDATE `users` SET `quota` = 0 WHERE `id` = :id';
        return (Users::getInstance())->update($sql, [
            'id' => $id
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
    private static $quotaMax = '1073741824'; //流量单位转换,入库需要, 1G*1024*1024*1024 = 1073741824byte

    public function getUsersByJson()
    {
        $usersData = json_decode(file_get_contents(self::$usersFile), true);
        if (!is_array($usersData)) {
            self::log(['ERROR: 会员json数据错误']);
            if ($_ENV['phpunit'] === '1') {
                throw new \Exception('ERROR: 会员json数据错误');
            }else{
                exit;
            }
        }
        $usersDataEnable = [];
        if (is_array($usersData) && count($usersData) > 0) {
            array_walk($usersData, function ($value) use (&$usersDataEnable) {
                if (
                    !isset($value['username'])
                    || strlen($value['username']) < 3
                    || strlen($value['username']) > 15
                    || !isset($value['password'])
                    || strlen($value['password']) < 3
                    || strlen($value['password']) > 15
                    || !isset($value['quota'])
                    || !isset($value['enable'])
                    || !isset($value['level'])
                    || !isset($value['expiryDate'])
                    || (!empty($value['expiryDate']) && strtotime(date('Y-m-d', strtotime($value['expiryDate']))) !== strtotime($value['expiryDate']))
                ) {
                    self::log(['ERROR: 会员json数据错误']);
                    if ($_ENV['phpunit'] === '1') {
                        throw new \Exception('ERROR: 会员json数据错误');
                    }else{
                        exit;
                    }
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

    // @codeCoverageIgnoreStart

    private static function base64($str)
    {
        return base64_encode($str);
    }

    private static function hash($str)
    {
        return hash('sha224', $str);
    }

    // @codeCoverageIgnoreEnd

    /**
     * 更新用户表
     * @dateTime 2020-11-11T23:47:23+0800
     * @author   twb<1174865138@qq.com>
     * @return   [type]                   [description]
     */
    public function handle()
    {
        date_default_timezone_set('Asia/Shanghai');

        $usersJson = $this->getUsersByJson();
        $usersMysql = UsersDbHandle::selectUser();
        $usersMysqlNew = $userIsset = $log = [];

        if (is_array($usersMysql) && count($usersMysql) > 0) {
            array_walk($usersMysql, function ($value) use (&$usersMysqlNew) {
                $usersMysqlNew[$value['username']] = $value;
            });
        }


        try {

            UsersDbHandle::beginTransaction();

            if (is_array($usersJson) && count($usersJson) > 0) {
                array_walk($usersJson, function ($value) use ($usersMysqlNew, &$userIsset, &$log) {
                    $value['useDays'] = $value['level'] > 0 ? 0 : 30;
                    if (isset($usersMysqlNew[$value['username']])) {
                        $userIsset[] = $value['username'];
                        if (
                            $usersMysqlNew[$value['username']]['passwordShow'] !== $value['passwordShow']
                            || $usersMysqlNew[$value['username']]['quota'] != $value['quota']
                            || $usersMysqlNew[$value['username']]['useDays'] != $value['useDays']
                            || $usersMysqlNew[$value['username']]['expiryDate'] != $value['expiryDate']
                        ) {
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
            if (is_array($userDiff) && count($userDiff) > 0) {
                array_walk($userDiff, function ($value) use ($usersMysqlNew, &$log) {
                    UsersDbHandle::delUser($usersMysqlNew[$value]['id']); //删
                    $log[] = 'del: ' . json_encode($usersMysqlNew[$value], JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES);
                });
            }

            UsersDbHandle::commit();

            $this->expiry();

        // @codeCoverageIgnoreStart

        } catch (Exception $e) {
            UsersDbHandle::rollBack();
            $err = 'Error: 数据库操作回滚;detail: ' . $e->getMessage();
            $log[] = $err;
            if ($_ENV['phpunit'] === '1') {
                throw new \Exception($err);
            }
        }

        // @codeCoverageIgnoreEnd

        self::log($log);

        if ($_ENV['phpunit'] === '1') {
            return $log;
        }

    }

    /**
     * 清除流量上下行的记录
     * @dateTime 2020-11-11T23:46:04+0800
     * @author   twb<1174865138@qq.com>
     */
    public function clear()
    {
        UsersDbHandle::beginTransaction();
        try {
            UsersDbHandle::clear();
            UsersDbHandle::commit();
            $log[] = '!!!!!!!!!!!!!!!!!!!!! Clear: 流量清零 !!!!!!!!!!!!!!!!!!!!!!';

        // @codeCoverageIgnoreStart

        }catch (Exception $e) {
            UsersDbHandle::rollBack();
            $err = 'Error: 出现错误;detail: ' . $e->getMessage();
            $log[] = $err;
            if ($_ENV['phpunit'] === '1') {
                throw new \Exception($err);
            }
        }

        // @codeCoverageIgnoreEnd

        self::log($log);
        echo '流量清零完成' . PHP_EOL;
    }

    /**
     * 处理过期用户
     * @dateTime 2021-03-05T16:35:25+0800
     * @author   twb<1174865138@qq.com>
     * @return   [type]                   [description]
     */
    public function expiry()
    {
        // $date = date("Y-m-d",strtotime("-1 day"));
        $date = date("Y-m-d");
        $usersMysql = UsersDbHandle::selectUser('`quota` != 0 AND `useDays` != 0 AND `expiryDate` = "' . $date . '"');
        if (!is_array($usersMysql) || count($usersMysql) < 1) {
            return true;
        }

        try {
            UsersDbHandle::beginTransaction();
            $users = '';
            array_walk($usersMysql, function($value) use(&$users){
                UsersDbHandle::expiry($value['id']);
                $users .= '[' . $value['id'] . ']' . $value['username'];
            });

            UsersDbHandle::commit();
            $log[] = '过期用户处理: ' . $users;

        // @codeCoverageIgnoreStart

        }catch (Exception $e) {
            UsersDbHandle::rollBack();
            $err = 'Error: 出现错误;detail: ' . $e->getMessage();
            $log[] = $err;
            if ($_ENV['phpunit'] === '1') {
                throw new \Exception($err);
            }
        }

        // @codeCoverageIgnoreEnd

        self::log($log);
        echo '过期用户处理完成' . PHP_EOL;
    }

    public static function log($arr)
    {
        try {
            if (!is_array($arr)) {
                $log = PHP_EOL . date('Y-m-d H:i:s', time()) . '   ' . (string) $arr;
                file_put_contents(self::$logFile, $log, FILE_APPEND | LOCK_EX);
                return true;
            } elseif (count($arr) > 0) {
                array_walk($arr, function ($value) {
                    $log = PHP_EOL . date('Y-m-d H:i:s', time()) . '   ' . $value;
                    file_put_contents(self::$logFile, $log, FILE_APPEND | LOCK_EX);
                });
                return true;
            }
            return false;

        // @codeCoverageIgnoreStart

        }catch (Exception $e) {
            $err = 'Error: 出现错误;detail: ' . $e->getMessage();
            $log[] = $err;
            if ($_ENV['phpunit'] === '1') {
                throw new \Exception($err);
            }
        }

        // @codeCoverageIgnoreEnd

    }

}

// (new Library/UserHandle)->handle();
