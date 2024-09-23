package main

import (
	"log"

	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/initialize"
	initGlobal "github.com/twbworld/proxy/initialize/global"
	"github.com/twbworld/proxy/initialize/system"
)

func main() {
	initGlobal.New().Start()
	initialize.InitializeLogger()
	sys := system.Start()
	defer sys.Stop()

	defer func() {
		if p := recover(); p != nil {
			global.Log.Println(p)
		}
	}()

	switch initGlobal.Act {
	case "":
		initialize.Start()
	case "clear":
		initialize.Clear()
	case "expiry":
		initialize.Expiry()
	default:
		log.Println("参数可选: clear|expiry")
	}

}
