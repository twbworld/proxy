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
	flag.StringVar(&act, "a", "", `行为,默认为空,即启动服务; "clear": 清除上下行流量记录; "expiry": 处理过期用户`)
	flag.Parse()

	defer func() {
		if p := recover(); p != nil {
			global.Log.Println(p)
		}
		if dao.DB != nil {
			if err := dao.Close(); err != nil {
				global.Log.Println("数据库关闭出错[joiasjofg]", err)
			}
		}
	}()

	switch act {
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
