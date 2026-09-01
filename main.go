package main

import (
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

	if err := initSvc.Run(); err != nil {
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
	var err error
	switch action {
	case "clear":
		err = taskManager.Clear()
	case "expiry":
		err = taskManager.Expiry()
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
