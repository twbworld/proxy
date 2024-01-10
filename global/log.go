package global

import (
	"io"
	"log"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/twbworld/proxy/utils"
)

func initLog(runLogPath string) {

	err := utils.CreateFile(runLogPath)
	if err != nil {
		log.Panicln("创建文件错误: ", err)
	}

	Log = logrus.New()
	Log.SetFormatter(&logrus.JSONFormatter{})
	Log.SetLevel(logrus.InfoLevel)

	runfile, err := os.OpenFile(runLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Panicln("打开文件错误: ", err)
	}
	Log.SetOutput(io.MultiWriter(os.Stdout, runfile)) //同时输出到终端和日志
}
