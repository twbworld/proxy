package initialize

import (
	"context"
	"io"
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

func Init() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM) //监听关闭(ctrl+C)指令
	defer stop()

	TgInit()

	ginfile, err := os.OpenFile(global.Config.AppConfig.GinLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		global.Log.Fatalln("打开文件错误: ", err)
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

	// ginServer.Run(":80")
	server := &http.Server{
		Addr:    global.Config.AppConfig.GinAddr,
		Handler: ginServer,
	}

	//协程启动服务
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			global.Log.Fatalln(err)
		}
	}()

	global.Log.Infof("启动成功, port: %s, pid: %d", global.Config.AppConfig.GinAddr, syscall.Getpid())
	service.TgSend("启动成功")

	<-ctx.Done() //阻塞等待
	//来到这 证明有关闭指令,将进行平滑优雅关闭服务

	stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) //如果有连接就超时5s后关闭
	defer cancel()

	//关闭监听端口
	if err := server.Shutdown(ctx); err != nil {
		global.Log.Fatalln("Server forced to shutdown: ", err)
	}
	service.TgSend("程序关闭成功")
	service.TgWebhookClear()
	global.Log.Fatalln("服务关闭!")
}
