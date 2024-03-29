package global

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/twbworld/proxy/utils"
)

func initLog(runLogPath string) {
	if err := utils.CreateFile(runLogPath); err != nil {
		panic("创建文件错误[oirdtug]: " + err.Error())
	}

	Log = logrus.New()
	Log.SetFormatter(&logrus.JSONFormatter{})
	Log.SetLevel(logrus.InfoLevel)

	runfile, err := os.OpenFile(runLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic("打开文件错误[0atrpf]: " + err.Error())
	}
	Log.SetOutput(io.MultiWriter(os.Stdout, runfile)) //同时输出到终端和日志
}
