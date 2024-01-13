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

	ginfile, err := os.OpenFile(global.Config.AppConfig.GinLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic("打开文件错误[nvcbjkdfgu]: " + err.Error())
	}
	gin.DefaultWriter = io.MultiWriter(ginfile) //记录所有日志
	gin.DefaultErrorWriter = global.Log.Out
	gin.DisableConsoleColor() //将日志写入文件时不需要控制台颜色
	mode := gin.ReleaseMode
	if global.Config.Env.Debug {
		mode = gin.DebugMode
	}
	gin.SetMode(mode)

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
			global.Log.Panic("服务出错[isjfio]: ", err.Error()) //外部并不能捕获Panic
		}
	}()

	TgInit()

	global.Log.Infof("启动成功, port: %s, pid: %d", global.Config.AppConfig.GinAddr, syscall.Getpid())
	service.TgSend("启动成功")

	//监听关闭(ctrl+C)指令
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	<-ctx.Done() //阻塞等待

	//来到这 证明有关闭指令,将进行平滑优雅关闭服务

	global.Log.Infof("程序关闭中..., port: %s, pid: %d", global.Config.AppConfig.GinAddr, syscall.Getpid())

	stop()

	//给程序最多5秒处理余下请求
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//关闭监听端口
	if err := server.Shutdown(timeoutCtx); err != nil {
		global.Log.Panicln("服务关闭出错[oijojiud]", err)
	}
	service.TgSend("服务退出成功")
	TgClear()
	global.Log.Infoln("服务退出成功")

}
