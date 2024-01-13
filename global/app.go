package global

import (
	"log"

	"github.com/twbworld/proxy/config"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

// 全局变量
var (
	Config *config.Configuration = new(config.Configuration) //指针类型, 给与其内存空间
	Log    *logrus.Logger
	Bot    *tg.BotAPI
)

func Init() {
	defer func() {
		if p := recover(); p != nil {
			log.Println(p)
		}
	}()

	initConfig()
	initLog(Config.AppConfig.RunLogPath)
	initEnv(Config.AppConfig.EnvPath)
}
