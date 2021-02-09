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
     * @covers Library\UserHandle::handle
     * @todo   Implement testHandle().
     */
    public function testHandle(): void
    {
        $users = $this->object->handle();
        // $this->assertIsArray($users);
    }


}
