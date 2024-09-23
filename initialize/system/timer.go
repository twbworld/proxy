package system

import (
	"github.com/robfig/cron/v3"
	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/task"
)

var c *cron.Cron

func timerStart() error {
	var option []cron.Option
	// option = append(option, cron.WithSeconds()) //精确到秒
	c = cron.New(option...)

	_, err := c.AddFunc("0 0 * * *", func() {
		task.Clean()
	})
	if err != nil {
		return err
	}

	c.Start() //已含协程
	global.Log.Infoln("定时器启动成功")
	return nil
}

func timerStop() error {
	if c == nil {
		global.Log.Warnln("定时器未启动")
		return nil
	}
	c.Stop()
	global.Log.Infoln("定时器停止成功")
	return nil
}
