package global

import (
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/twbworld/proxy/utils"

	"github.com/twbworld/proxy/global"
)

func (*GlobalInit) initLog() error {
	if err := utils.CreateFile(global.Config.RunLogPath); err != nil {
		return fmt.Errorf("创建文件错误[oirdtug]: %w", err)
	}

	global.Log = logrus.New()
	global.Log.SetFormatter(&logrus.JSONFormatter{})
	global.Log.SetLevel(logrus.InfoLevel)

	runfile, err := os.OpenFile(global.Config.RunLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("打开文件错误[0atrpf]: %w", err)
	}
	global.Log.SetOutput(io.MultiWriter(os.Stdout, runfile)) //同时输出到终端和日志
	return nil
}
