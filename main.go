package main

import (
	"log"

	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/initialize"
	initGlobal "github.com/twbworld/proxy/initialize/global"
	"github.com/twbworld/proxy/initialize/system"
	"github.com/twbworld/proxy/task"
)

func main() {
	initGlobal.New().Start()
	initialize.InitializeLogger()
	if err := system.DbStart(); err != nil {
		global.Log.Fatalf("连接数据库失败[fbvk89]: %v", err)
	}
	defer system.DbClose()

	defer func() {
		if p := recover(); p != nil {
			global.Log.Println(p)
		}
	}()

	switch initGlobal.Act {
	case "":
		initialize.Start()
	case "clear":
		task.Clear()
	case "expiry":
		task.Expiry()
	default:
		log.Println("参数可选: clear|expiry")
	}

}
