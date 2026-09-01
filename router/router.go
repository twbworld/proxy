package router

import (
	"net/http"
	"strings"

	"github.com/twbworld/proxy/controller"
	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/middleware"
	"github.com/twbworld/proxy/model/common"

	"github.com/gin-gonic/gin"
)

func Start(ginServer *gin.Engine) {
	// 限制form内存(默认32MiB)
	ginServer.MaxMultipartMemory = 32 << 20

	ginServer.Use(middleware.CorsHandle(), middleware.OptionsMethod) //全局中间件

	ginServer.StaticFile("/favicon.ico", global.Config().StaticDir+"/favicon.ico")
	ginServer.StaticFile("/robots.txt", global.Config().StaticDir+"/robots.txt")
	ginServer.LoadHTMLGlob(global.Config().StaticDir + "/*.html")
	ginServer.StaticFS("/static", http.Dir(global.Config().StaticDir))

	// 错误处理路由
	errorRoutes := []string{"404.html", "40x.html", "50x.html"}
	for _, route := range errorRoutes {
		ginServer.GET(route, func(ctx *gin.Context) {
			ctx.HTML(http.StatusOK, "404.html", gin.H{"status": route[:3]})
		})
		ginServer.POST(route, func(ctx *gin.Context) {
			common.FailNotFound(ctx)
		})
	}

	ginServer.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "404.html", nil)
	})

	ginServer.NoRoute(func(ctx *gin.Context) {
		if strings.Contains(ctx.Request.Header.Get("Accept"), "text/html") {
			ctx.HTML(http.StatusNotFound, "404.html", gin.H{"status": "404"})
		} else {
			common.FailNotFound(ctx)
		}
	})

	wh := ginServer.Group("wh")
	{
		wh.POST("/tg/:token", middleware.ValidatorBotToken, controller.Admin.BotApi.Webhook)
	}

	// gin路由不支持正则,我服了
	ginServer.GET("/:filename", func(ctx *gin.Context) {
		filename := ctx.Param("filename")
		userName, ok := strings.CutSuffix(filename, ".html")
		if !ok || userName == "404" {
			ctx.Request.URL.Path = "/404.html"
			ginServer.HandleContext(ctx)
			ctx.Abort()
			return
		}
		ctx.Set("userName", userName)
	}, middleware.ValidatorSubscribe, controller.User.IndexApi.Subscribe)

}
