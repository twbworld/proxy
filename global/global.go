package global

import (
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/twbworld/proxy/internal/telegram"
	"github.com/twbworld/proxy/model/config"
)

// 全局变量(业务逻辑禁止修改)
var (
	Version   string
	configPtr atomic.Pointer[config.Config] //无锁原子操作
	Log       *logrus.Logger
	Tz        *time.Location

	DBClient    DBService
	RedisClient RedisService
	Notifier    NotifierService = telegram.NewNoopClient(nil) // 默认绑定空操作通知实例
)

// InitConfig 初始化或安全替换配置
func InitConfig(c *config.Config) {
	configPtr.Store(c)
}

// Config 获取配置的只读快照 (并发安全)
func Config() *config.Config {
	return configPtr.Load()
}
