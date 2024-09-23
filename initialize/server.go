package initialize

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/router"
	"github.com/twbworld/proxy/service"
	"github.com/twbworld/proxy/utils"

	"github.com/gin-gonic/gin"
)

var server *http.Server

func InitializeLogger() {
	ginfile, err := os.OpenFile(global.Config.GinLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		global.Log.Fatalf("打开文件错误[fsmk89]: %v", err)
	}
	gin.DefaultWriter, gin.DefaultErrorWriter = io.MultiWriter(ginfile), global.Log.Out //记录所有日志
	gin.DisableConsoleColor()                                                           //将日志写入文件时不需要控制台颜色
}

func Start() {
	initializeGinServer()
	//协程启动服务
	go startServer()

	logStartupInfo()

	service.Service.UserServiceGroup.TgService.TgSend("已启动")

	waitForShutdown()
}

func initializeGinServer() {
	mode := gin.ReleaseMode
	if global.Config.Debug {
		mode = gin.DebugMode
	}
	gin.SetMode(mode)

	ginServer := gin.Default()
	router.Start(ginServer)

	ginServer.ForwardedByClientIP = true

	// ginServer.Run(":80")
	server = &http.Server{
		Addr:    global.Config.GinAddr,
		Handler: ginServer,
	}
}

// 启动HTTP服务器
func startServer() {
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		global.Log.Panic("服务出错[isjfio]: ", err.Error()) //外部并不能捕获Panic
	}
}

// 记录启动信息
func logStartupInfo() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	global.Log.Infof("已启动, version: %s, port: %s, pid: %d, mem: %gMiB", runtime.Version(), global.Config.GinAddr, syscall.Getpid(), utils.NumberFormat(float32(m.Alloc)/1024/1024))

}

// 等待关闭信号(ctrl+C)
func waitForShutdown() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done() //阻塞等待

	//来到这 证明有关闭指令,将进行平滑优雅关闭服务

	global.Log.Infof("程序关闭中..., port: %s, pid: %d", global.Config.GinAddr, syscall.Getpid())

	shutdownServer()
}

// 平滑关闭服务器
func shutdownServer() {
	//给程序最多5秒处理余下请求
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//关闭监听端口
	if err := server.Shutdown(timeoutCtx); err != nil {
		global.Log.Panicln("服务关闭出错[oijojiud]", err)
	}
	service.Service.UserServiceGroup.TgService.TgSend("服务退出成功")
	global.Log.Infoln("服务退出成功")
}
