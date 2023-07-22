package main

import (
	"flag"
	"log"
	"github.com/twbworld/proxy/dao"
	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/initialize"
)

func main() {

	var act string
	flag.StringVar(&act, "a", "", `行为,默认为空,即启动服务; "clear": 清除上下行流量记录; "expiry": 处理过期用户; "handle": 更新用户`)
	flag.Parse()

	global.Init()
	dao.InitMysql()

	switch act {
	case "":
		initialize.Init()
	case "clear":
		initialize.Clear()
	case "expiry":
		initialize.Expiry()
	case "handle":
		initialize.Handle()
	default:
		log.Println("参数可选: clear|expiry|handle")
	}

	global.Log.Info("完成")

}
