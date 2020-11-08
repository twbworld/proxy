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
    private $quotaMax = '1073741824'; //1G*1024*1024*1024 = 1073741824byte
    private $db;

    public function __construct()
    {
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

    private function log($arr)
    {
        if (!is_array($arr)) {
            $log = PHP_EOL . date('Y-m-d H:i:s', time()) . '   ' . (string) $arr;
            file_put_contents($this->logFile, $log, FILE_APPEND | LOCK_EX);
        }
        if (count($arr) > 0) {
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
            'quota' => ($value['quota'] ?? -1) * $this->quotaMax,
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
            'quota' => ($value['quota'] ?? -1) * $this->quotaMax,
            'download' => 0,
            'upload' => 0,
        ]);

    }

    private function delUser($username)
    {
        $sql = 'DELETE FROM `users` where `id` = :id';
        $sth = $this->db->prepare($sql);
        $sth->execute([
            'id' => $id,
        ]);
    }

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
                    if ($usersMysqlNew[$value['username']]['passwordShow'] !== $this->base64($value['passwordShow']) || $usersMysqlNew[$value['username']]['quota'] != ($value['quota'] * $this->quotaMax)) {
                        $value['id'] = $usersMysqlNew[$value['username']]['id'];
                        $this->updateUser($value); //改
                        $log[] = 'update: ' . json_encode($usersMysqlNew[$value['username']]) . ' => ' . json_encode($value);
                    }
                } else {
                    $this->addUser($value); //增
                    $log[] = 'add: ' . json_encode($value);
                }
            });
        }

        $userDiff = array_diff(array_keys($usersMysqlNew), $userIsset);
        if (count($userDiff) > 0) {
            array_walk($userDiff, function ($value) use ($usersMysqlNew, &$log) {
                $this->delUser($usersMysqlNew[$value]['id']); //删
                $log[] = 'del: ' . json_encode($usersMysqlNew[$value]);
            });
        }

        $this->db->commit();

        $this->log($log);

    }

}

date_default_timezone_set('Asia/Shanghai');

(new UserHandle())->handle();
