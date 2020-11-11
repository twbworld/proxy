<?php

/**
 * 处理用户信息,对mysql数据库作处理
 */
class UserHandle
{
    private $dbname = 'trojan';
    private $host = 'mysql';
    private $username = 'root';
    private $password = 'tp';
    private $usersFile = 'users.json';
    private $logFile = 'userHandle.log';
    private $quotaMax = '1073741824'; //入库需要, 1G*1024*1024*1024 = 1073741824byte
    private $db;

    public function __construct()
    {
        date_default_timezone_set('Asia/Shanghai');
        $dsn = 'mysql:dbname=' . $this->dbname . ';host=' . $this->host;

        try {
            $this->db = new \PDO($dsn, $this->username, $this->password);
            $this->db->exec('set names utf8');
        } catch (\PDOException $e) {
            echo 'Connection failed: ' . $e->getMessage();
        }
    }

    private function getUsersByJson()
    {
        $usersData = json_decode(file_get_contents($this->usersFile), true);
        if (!is_array($usersData)) {
            $this->log(['ERROR: 会员json数据错误']);
            exit;
        }
        $usersDataEnable = [];
        if (count($usersData) > 0) {
            array_walk($usersData, function ($value) use (&$usersDataEnable) {
                if (!isset($value['username']) || strlen($value['username']) < 3 || strlen($value['username']) > 15 || !isset($value['password']) || strlen($value['password']) < 3 || strlen($value['password']) > 15 || !isset($value['quota']) || !isset($value['enable']) || !isset($value['level'])) {
                    $this->log(['ERROR: 会员json数据错误']);
                    exit;
                }
                if ($value['enable'] === true) {
                    $value['quota'] > 0 && ($value['quota'] = $value['quota'] * $this->quotaMax);
                    $usersDataEnable[] = $value;
                }
            });
        }

        return $usersDataEnable;
    }

    private function getUsersByMysql()
    {
        $sql = 'SELECT `id`, `username`, `passwordShow`, `quota` FROM `users`';
        $sth = $this->db->query($sql);
        $usersData = $sth->fetchAll(\PDO::FETCH_ASSOC);
        return $usersData;
    }

    private function base64($str)
    {
        return base64_encode($str);
    }

    private function hash($str)
    {
        return hash('sha224', $str);
    }

    public function log($arr)
    {
        if (!is_array($arr)) {
            $log = PHP_EOL . date('Y-m-d H:i:s', time()) . '   ' . (string) $arr;
            file_put_contents($this->logFile, $log, FILE_APPEND | LOCK_EX);
        }elseif (count($arr) > 0) {
            array_walk($arr, function ($value) {
                $log = PHP_EOL . date('Y-m-d H:i:s', time()) . '   ' . $value;
                file_put_contents($this->logFile, $log, FILE_APPEND | LOCK_EX);
            });
        }
    }

    private function updateUser($value)
    {
        $sql = 'SELECT `username` FROM `users` WHERE `id` = ' . $value['id'] . ' FOR UPDATE';
        $sth = $this->db->query($sql);
        $sth->fetchAll(\PDO::FETCH_ASSOC);

        $sql = 'UPDATE `users` SET `password` = :password, `passwordShow` = :passwordShow, `quota` = :quota WHERE `id` = :id';
        $sth = $this->db->prepare($sql);
        $sth->execute([
            'id' => $value['id'],
            'password' => $this->hash($value['password']),
            'passwordShow' => $this->base64($value['password']),
            'quota' => $value['quota'] ?? -1,
        ]);

    }

    private function addUser($value)
    {
        $sql = 'INSERT INTO `users`(`username`, `password`, `passwordShow`, `quota`, `download`, `upload`)
                VALUES(:username, :password, :passwordShow, :quota, :download, :upload)';
        $sth = $this->db->prepare($sql);
        $sth->execute([
            'username' => $value['username'],
            'password' => $this->hash($value['password']),
            'passwordShow' => $this->base64($value['password']),
            'quota' => $value['quota'] ?? -1,
            'download' => 0,
            'upload' => 0,
        ]);

    }

    private function delUser($id)
    {
        $sql = 'DELETE FROM `users` where `id` = :id';
        $sth = $this->db->prepare($sql);
        $sth->execute([
            'id' => $id,
        ]);
    }

    /**
     * 更新用户表
     * @dateTime 2020-11-11T23:47:23+0800
     * @author   twb<1174865138@qq.com>
     * @return   [type]                   [description]
     */
    public function handle()
    {
        $usersJson = $this->getUsersByJson();
        $usersMysql = $this->getUsersByMysql();
        $usersMysqlNew = $userIsset = $log = [];

        if (count($usersMysql) > 0) {
            array_walk($usersMysql, function ($value) use (&$usersMysqlNew) {
                $usersMysqlNew[$value['username']] = $value;
            });
        }

        $this->db->beginTransaction();

        if (count($usersJson) > 0) {
            array_walk($usersJson, function ($value) use ($usersMysqlNew, &$userIsset, &$log) {
                if (isset($usersMysqlNew[$value['username']])) {
                    $userIsset[] = $value['username'];
                    if ($usersMysqlNew[$value['username']]['passwordShow'] !== $this->base64($value['password']) || $usersMysqlNew[$value['username']]['quota'] != $value['quota']) {
                        $value['id'] = $usersMysqlNew[$value['username']]['id'];
                        $this->updateUser($value); //改
                        $log[] = 'update: ' . json_encode($usersMysqlNew[$value['username']], JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES) . ' => ' . json_encode($value, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES);
                    }
                } else {
                    $this->addUser($value); //增
                    $log[] = 'add: ' . json_encode($value, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES);
                }
            });
        }

        $userDiff = array_diff(array_keys($usersMysqlNew), $userIsset);
        if (count($userDiff) > 0) {
            array_walk($userDiff, function ($value) use ($usersMysqlNew, &$log) {
                $this->delUser($usersMysqlNew[$value]['id']); //删
                $log[] = 'del: ' . json_encode($usersMysqlNew[$value], JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES);
            });
        }

        $this->db->commit();

        $this->log($log);

    }

    /**
     * 清除流量上下行的记录
     * @dateTime 2020-11-11T23:46:04+0800
     * @author   twb<1174865138@qq.com>
     */
    public function clear()
    {
        $sql = 'UPDATE `users` SET `download` = :download, `upload` = :upload';
        $sth = $this->db->prepare($sql);
        $sth->execute([
            'download' => 0,
            'upload' => 0,
        ]);
        $this->log(['!!!!!!!!!!!!!!!!!!!!! Clear: 流量清零 !!!!!!!!!!!!!!!!!!!!!!']);
    }

}
