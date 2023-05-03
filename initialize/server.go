package initialize

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/router"
	"github.com/twbworld/proxy/service"

	"github.com/gin-gonic/gin"
)

func Init(){
	ginfile, err := os.OpenFile(global.Config.AppConfig.GinLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalln("打开文件错误: ", err)
	}
	gin.DefaultWriter = io.MultiWriter(ginfile)

	mode := gin.ReleaseMode
	if global.Config.Env.Debug {
		mode = gin.DebugMode
	}

	gin.SetMode(mode)

	gin.DisableConsoleColor() //禁用控制台颜色,将日志写入文件时不需要控制台颜色

	ginServer := gin.Default()


	router.Init(ginServer)

	// ginServer.Run(":8081")
	server := http.Server{Addr: global.Config.AppConfig.GinAddr, Handler: ginServer}

	//协程启动服务
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalln(err)
		}
	}()

	log.Println("启动成功")
	service.TgSend("启动成功")

	closeBy(&server)

}

// 平滑优雅关闭服务
func closeBy(server *http.Server) {
	sb := make(chan os.Signal, 1)
	signal.Notify(sb, syscall.SIGINT, syscall.SIGTERM) //监听关闭(ctrl+C)指令
	<-sb                                               //阻塞等待

	service.TgSend("关闭中...")

	//来到这 证明有关闭指令
	c, f := context.WithTimeout(context.Background(), 5*time.Second) //如果有连接就超时5s后关闭
	defer f()

	//关闭监听端口
	if err := server.Shutdown(c); nil != err {
		log.Fatalln(err)
	}
	log.Println("服务关闭!")
	service.TgSend("程序关闭成功")
}
