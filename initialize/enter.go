package initialize

import (
	"context"
	"io"
	"sync"

	"github.com/robfig/cron/v3"
	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/task"
	"golang.org/x/sync/errgroup"
)

// Initializer 统一管理项目的所有初始化工作
type Initializer struct {
	cron           *cron.Cron
	logFileClosers []io.Closer
	reloadLock     sync.Mutex // 用于热重载的锁
	taskManager    *task.Manager
}

// Run 并发执行所有核心服务的初始化
func (i *Initializer) Run(ctx context.Context) error {
	eg, egCtx := errgroup.WithContext(ctx)

	// 关键任务，失败会终止程序
	eg.Go(i.initDB)

	// 非关键任务，失败只打印日志，不影响启动
	eg.Go(func() error {
		if err := i.initRedis(egCtx); err != nil {
			global.Log.Warnf("初始化Redis客户端失败: %v", err)
		} else {
			global.Log.Info("初始化Redis服务成功")
		}
		return nil
	})
	eg.Go(func() error {
		if err := i.initNotify(); err != nil {
			global.Log.Warnf("初始化通知服务失败: %v", err)
		} else {
			global.Log.Info("初始化通知服务成功")
		}
		return nil
	})

	return eg.Wait()
}

// Close 优雅地关闭和释放所有资源
func (i *Initializer) Close() {
	if i.notifyClose() == nil {
		global.Log.Info("通知服务已关闭")
	}
	if i.redisClose() == nil {
		global.Log.Info("Redis客户端已关闭")
	}
	if i.dbClose() == nil {
		global.Log.Infof("%s已关闭", global.Config().Database.Type)
	}
	i.timerStop()
	_ = i.logClose()
}

// logClose 关闭或刷新日志组件
func (i *Initializer) logClose() error {
	var lastErr error
	for _, closer := range i.logFileClosers {
		if err := closer.Close(); err != nil {
			// 记录最后一个错误，但继续尝试关闭其他文件
			lastErr = err
		}
	}
	return lastErr
}

// StartSystem 启动系统级服务，如定时器和数据加载
func (i *Initializer) StartSystem(taskManager *task.Manager) {
	if err := i.timerStart(taskManager); err != nil {
		panic(err)
	}
}
