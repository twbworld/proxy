package initialize

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/utils"
)

// setupLogFile 是一个辅助函数，用于创建和打开一个每日轮转的日志文件。
func (i *Initializer) setupLogFile(logPath string) (*os.File, error) {
	// 日志命名规范,如:gin.log.2025-10-28
	dateSuffix := time.Now().In(global.Tz).Format("2006-01-02")
	dailyLogPath := fmt.Sprintf("%s.%s", logPath, dateSuffix)

	if err := utils.CreateFile(dailyLogPath); err != nil {
		return nil, fmt.Errorf("创建日志文件 '%s' 失败: %w", dailyLogPath, err)
	}

	file, err := os.OpenFile(dailyLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("打开日志文件 '%s' 失败: %w", dailyLogPath, err)
	}

	i.logFileClosers = append(i.logFileClosers, file)
	return file, nil
}

type CustomJSONFormatter struct {
	logrus.JSONFormatter
}

func (f *CustomJSONFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	entry.Time = entry.Time.In(global.Tz)
	return f.JSONFormatter.Format(entry)
}

// InitLog 初始化logrus日志库
func (i *Initializer) InitLog() error {
	runfile, err := i.setupLogFile(global.Config().RunLogPath)
	if err != nil {
		return fmt.Errorf("初始化运行日志失败: %w", err)
	}

	global.Log = logrus.New()
	global.Log.SetFormatter(&CustomJSONFormatter{
		TimestampFormat: time.RFC3339,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "msg",
			logrus.FieldKeyTime:  "time",
		},
	})
	if global.Config().Debug {
		global.Log.SetLevel(logrus.DebugLevel)
	} else {
		global.Log.SetLevel(logrus.InfoLevel)
	}

	global.Log.SetOutput(io.MultiWriter(os.Stdout, runfile))
	return nil
}
