<?php
namespace Test;

use Library\UserHandle;
use PHPUnit\Framework\TestCase;

require __DIR__ . '/../src/library/UserHandle.php';

class UserHandleTest extends TestCase
{

    protected $object;

    protected function setUp(): void
    {
        $this->object = new UserHandle();
    }

    /**
     * @dataProvider dataHandle
     * @covers Library\UserHandle::handle
     * @todo   Implement testHandle().
     */
    public function testHandle(array $data): void
    {
        //在类UserHandle下模拟一个getUsersByJson的方法,并设定返回值
        $mock = $this->getMockBuilder(UserHandle::class)->setMethods(['getUsersByJson'])->getMock();
        $mock->expects($this->once())->method('getUsersByJson')->willReturn(reset($data));
        $logs = $mock->handle();

        $this->assertStringStartsWith(key($data), $logs[0]);
    }

    /**
     * @covers Library\UserHandle::getUsersByJson
     * @todo   Implement testGetUsersByJson().
     */
    public function testGetUsersByJson(): void
    {
        $result = $this->object->getUsersByJson();

        $this->assertIsArray($result);
        $this->assertIsString($result[0]['password']);
    }

    /**
     * @covers Library\UserHandle::clear
     * @todo   Implement testClear().
     */
    public function testClear(): void
    {
        $this->object->clear();

        $this->expectOutputString('流量清零完成' . PHP_EOL);
    }

    /**
     * @dataProvider dataLog
     * @covers Library\UserHandle::log
     * @todo   Implement testLog().
     */
    public function testLog($data): void
    {
        $this->assertTrue($this->object::log($data));
    }

    public function dataHandle(): array
    {
        return [
            [[
                'add' => [
                    [
                        "username" => "test",
                        "password" => "90a3ed9e32b2aaf4c61c410eb925426119e1a9dc53d4286ade99a809",
                        "quota" => 107374182400,
                        "enable" => true,
                        "level" => 0,
                        "expiryDate" => "2050-01-01",
                        "passwordShow" => "dGVzdA==",
                    ],
                    [
                        "username" => "test2",
                        "password" => "90a3ed9e32b2aaf4c61c410eb925426119e1a9dc53d4286ade99a809",
                        "quota" => 107374182400,
                        "enable" => true,
                        "level" => 0,
                        "expiryDate" => "2050-01-01",
                        "passwordShow" => "dGVzdA==",
                    ],
                ],
            ]],
            [[
                'del' => [
                    [
                        "username" => "test",
                        "password" => "90a3ed9e32b2aaf4c61c410eb925426119e1a9dc53d4286ade99a809",
                        "quota" => 107374182400,
                        "enable" => true,
                        "level" => 0,
                        "expiryDate" => "2050-01-01",
                        "passwordShow" => "dGVzdA==",
                    ],
                ],
            ]],
            [[
                'update' => [
                    [
                        "username" => "test",
                        "password" => "90a3ed9e32b2aaf4c61c410eb925426119e1a9dc53d4286ade99a809",
                        "quota" => 107374182400,
                        "enable" => true,
                        "level" => 0,
                        "expiryDate" => "2050-01-02",
                        "passwordShow" => "dGVzdA==",
                    ],
                ],
            ]],
        ];
    }

    public function dataLog(): array
    {
        return [
            ['测试'],
            [['测试2', '测试3']],
        ];
    }

}
