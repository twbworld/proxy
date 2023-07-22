package global

import (
	"github.com/twbworld/proxy/config"

	"github.com/sirupsen/logrus"
)

// 全局变量
var (
	Config      *config.Configuration = new(config.Configuration) //指针类型, 给与其内存空间
	Log         *logrus.Logger
)

func Init() {
	initConfig()
	initLog(Config.AppConfig.RunLogPath)
	initEnv(Config.AppConfig.EnvPath)
	initTrojanGoConfig(Config.AppConfig.TrojanGoConfigParh)
}
