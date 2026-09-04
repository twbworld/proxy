package initialize

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/task"
)

func (i *Initializer) timerStart(taskManager *task.Manager) error {
	i.cron = cron.New([]cron.Option{
		cron.WithLocation(global.Tz),
		cron.WithChain(cron.Recover(cron.DefaultLogger)),
	}...)

	// 每天凌晨1点执行，作为兜底
	if err := i.startCronJob(taskManager.Clear, "0 0 1 * *", "流量清零"); err != nil {
		return err
	}

	if err := i.startCronJob(taskManager.Expiry, "0 0 * * *", "处理过期用户"); err != nil {
		return err
	}

	if err := i.startCronJob(taskManager.CleanUpLogs, "0 3 * * *", "清除日志"); err != nil {
		return err
	}

	i.cron.Start() //已含协程
	global.Log.Infoln("定时器启动成功")
	return nil
}

func (i *Initializer) timerStop() {
	if i.cron == nil {
		global.Log.Warnln("定时器未启动")
		return
	}
	i.cron.Stop()
	global.Log.Infoln("定时器停止成功")
}

// 启动一个新的定时任务
func (i *Initializer) startCronJob(task func(context.Context) error, schedule string, name string) error {
	_, err := i.cron.AddFunc(schedule, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
		defer cancel()

		if err := task(ctx); err != nil {
			global.Log.Errorf("定时任务执行失败 [%s]: %v", name, err)
		}
	})
	return err
}
