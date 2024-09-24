package system

import (
	"github.com/robfig/cron/v3"
	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/service"
	"github.com/twbworld/proxy/task"
)

var c *cron.Cron

// startCronJob 启动一个新的定时任务
func startCronJob(task func() error, schedule, name string) error {
	_, err := c.AddFunc(schedule, func() {
		defer func() {
			text := "任务完成"
			if p := recover(); p != nil {
				text = "任务出错[gqxj]: " + p.(string)
			}
			service.Service.UserServiceGroup.TgService.TgSend(name + text)
		}()
		if err := task(); err != nil {
			panic(err)
		}
	})
	return err
}

func timerStart() error {
	c = cron.New([]cron.Option{
		cron.WithLocation(global.Tz),
		// cron.WithSeconds(), //精确到秒
	}...)

	if err := startCronJob(task.Clear, "0 0 1 * *", "流量清零"); err != nil {
		return err
	}

	if err := startCronJob(task.Expiry, "0 0 * * *", "处理过期用户"); err != nil {
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
