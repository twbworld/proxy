package system

import (
	"github.com/robfig/cron/v3"
	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/initialize"
	"github.com/twbworld/proxy/service"
	// "github.com/twbworld/proxy/task"
)

var c *cron.Cron

func timerStart() (err error) {
	c = cron.New([]cron.Option{
		cron.WithLocation(global.Tz),
		// cron.WithSeconds(), //精确到秒
	}...)

	_, err = c.AddFunc("0 0 1 * *", func() {
		defer func() {
			text := "流量清零任务完成"
			if p := recover(); p != nil {
				global.Log.Errorln("流量清零任务出错[gqnoj]: ", p)
				text = "流量清零任务出错[gqnoj]: " + p.(string)
			}
			service.Service.UserServiceGroup.TgService.TgSend(text)
		}()
		initialize.Clear()
	})
	if err != nil {
		return err
	}

	_, err = c.AddFunc("0 16 * * *", func() {
		defer func() {
			text := "处理过期用户任务完成"
			if p := recover(); p != nil {
				global.Log.Errorln("处理过期用户任务出错[54jni3]: ", p)
				text = "处理过期用户任务出错[54jni3]: " + p.(string)
			}
			service.Service.UserServiceGroup.TgService.TgSend(text)
		}()
		initialize.Expiry()
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
