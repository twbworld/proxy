package main

import (
	"context"
	"fmt"
	"time"

	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/initialize"
	"github.com/twbworld/proxy/task"
)

func main() {
	startTime := time.Now()

	taskManager := task.NewManager()
	initSvc := initialize.New(taskManager)

	if err := initSvc.InitTz(); err != nil {
		panic(fmt.Sprintf("初始化时区失败: %v", err))
	}

	if err := initSvc.InitLog(); err != nil {
		panic(fmt.Sprintf("初始化日志失败[fbvk89]: %v", err))
	}

	defer func() {
		if p := recover(); p != nil {
			global.Log.Errorln(p)
		}
	}()

	initCtx, cancelInit := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelInit()
	if err := initSvc.Run(initCtx); err != nil {
		global.Log.Fatalf("关键服务初始化失败，程序终止: %v", err)
	}
	defer initSvc.Close()

	initSvc.InitLogger()

	if initialize.Act != "" {
		dispatchAction(initialize.Act, taskManager)
		return
	}

	initialize.Start(initSvc, taskManager, startTime)
}

func dispatchAction(action string, taskManager *task.Manager) {
	global.Log.Infof("开始执行后台任务: %s", action)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
	defer cancel()

	var err error
	switch action {
	case "clear":
		err = taskManager.Clear(ctx)
	case "expiry":
		err = taskManager.Expiry(ctx)
	default:
		fmt.Println("未知的任务参数, 可选值: clear|expiry")
		return
	}

	if err == nil {
		global.Log.Infof("后台任务 '%s' 执行成功", action)
	} else {
		global.Log.Errorf("后台任务 '%s' 执行失败: %v", action, err)
	}
}
