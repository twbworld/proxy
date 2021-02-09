<?php
namespace Test;

use Library\Subscribe;
use PHPUnit\Framework\TestCase;

require __DIR__ . '/../src/library/Subscribe.php';

class SubscribeTest extends TestCase
{

    protected $object;

    protected function setUp(): void
    {
        $this->object = Subscribe::getInstance();
    }

    /**
     * @covers Library\Subscribe::getUsers
     * @todo   Implement testGetUsers().
     */
    public function testGetUsers(): void
    {
        $users = $this->object->getUsers();
        $this->assertIsArray($users);
    }

    /**
     * @dataProvider dataUsers
     * @covers Library\Subscribe::response
     * @todo   Implement testResponse().
     */
    public function testResponse($user): void
    {
        $_GET['u'] = $user;

        $this->assertStringEndsWith('=', $this->object->response());
    }

    /**
     * 测试用户数据
     * @dateTime 2021-02-09T00:02:43+0800
     * @author   twb<1174865138@qq.com>
     * @return   array                   用户数据
     */
    public function dataUsers(): array
    {
        $subscribe = Subscribe::getInstance();
        $userData = [];
        $userData[0] = array_map(function ($value) {
            return $value['password'];
        }, $subscribe->getUsers());

        return $userData;
        /*return [
            ['kxl'],
            ['twbworld'],
            ['abc']
        ];*/
    }

}
