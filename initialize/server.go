package initialize

import (
	"context"
	"github.com/twbworld/proxy/task"
	"io"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/router"
	"github.com/twbworld/proxy/service"
	"github.com/twbworld/proxy/service/admin"
	"github.com/twbworld/proxy/service/common"
	"github.com/twbworld/proxy/service/user"
	"github.com/twbworld/proxy/utils"
)

var server *http.Server

// Start 负责初始化并启动所有服务
func Start(initializer *Initializer, taskManager *task.Manager, startTime time.Time) {
	initializer.StartSystem(taskManager)

	service.Common = common.NewServiceGroup()
	service.User = user.NewServiceGroup()
	service.Admin = admin.NewServiceGroup(service.Common.AsyncService, global.Notifier)

	initGinServer()

	go startServer()

	logStartupInfo(startTime)

	waitForShutdown()
}

// 日志初始化
func (i *Initializer) InitLogger() {
	ginfile, err := i.setupLogFile(global.Config().GinLogPath)
	if err != nil {
		global.Log.Fatalf("初始化Gin日志失败: %v", err)
	}

	if global.Config().Debug {
		gin.DefaultWriter = io.MultiWriter(os.Stdout, ginfile)
	} else {
		gin.DefaultWriter = ginfile
	}

	gin.DefaultErrorWriter = io.MultiWriter(os.Stderr, ginfile)
	gin.DisableConsoleColor() //将日志写入文件时不需要控制台颜色
}

func initGinServer() {
	mode := gin.ReleaseMode
	if global.Config().Debug {
		mode = gin.DebugMode
	}
	gin.SetMode(mode)

	ginServer := gin.New()
	// 非Debug模式下,不记录常规的请求日志
	if global.Config().Debug {
		ginServer.Use(gin.Logger())
	}
	ginServer.Use(gin.Recovery())

	router.Start(ginServer)

	ginServer.ForwardedByClientIP = true

	csrf := http.NewCrossOriginProtection()
	if global.Config().Domain != "" {
		csrf.AddTrustedOrigin(global.Config().Domain)
	}

	server = &http.Server{
		Addr:    global.Config().GinAddr,
		Handler: csrf.Handler(ginServer),
	}
}

// 启动HTTP服务器
func startServer() {
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		global.Log.Panic("服务出错[isjfio]: ", err.Error())
	}
}

// 记录启动信息
func logStartupInfo(startTime time.Time) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	global.Log.Infof("已启动, 耗时: %v, 版本: %s, Go: %s, 端口: %s, 模式: %s, PID: %d, 内存: %gMiB", time.Since(startTime), global.Version, runtime.Version(), global.Config().GinAddr, gin.Mode(), syscall.Getpid(), utils.NumberFormat(float32(m.Alloc)/1024/1024))

	service.Common.AsyncService.GoSafe(func() {
		if err := global.Notifier.Send("已启动"); err != nil {
			global.Log.Warnf("发送TG通知失败: %v", err)
		}
	})
}

// 阻塞并等待关闭信号(ctrl+C)
func waitForShutdown() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done() //阻塞等待

	//来到这 证明有关闭指令,将进行平滑优雅关闭服务

	global.Log.Infof("程序关闭中..., port: %s, pid: %d", global.Config().GinAddr, syscall.Getpid())

	shutdownServer()
}

// 平滑关闭服务器
func shutdownServer() {
	//给程序最多5秒处理余下请求
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//关闭监听端口
	if err := server.Shutdown(timeoutCtx); err != nil {
		global.Log.Panicln("出错[oijojiud]", err)
	}
	global.Log.Infoln("退出成功")
	service.Common.AsyncService.GoSafe(func() {
		if err := global.Notifier.Send("退出成功"); err != nil {
			global.Log.Warnf("发送TG通知失败: %v", err)
		}
	})
}
