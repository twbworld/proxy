package global

import (
	"log"
	"os"

	"github.com/sirupsen/logrus"
)

func initLog(runLogPath string) {
	Log = logrus.New()
	Log.SetFormatter(&logrus.JSONFormatter{})
	Log.SetLevel(logrus.InfoLevel)

	runfile, err := os.OpenFile(runLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalln("打开文件错误: ", err)
	}
	Log.SetOutput(runfile)
}
