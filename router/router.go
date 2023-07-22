package router

import (
	"database/sql"
	"net/http"
	"regexp"

	"github.com/jxskiss/ginregex"
	"github.com/twbworld/proxy/controller"
	"github.com/twbworld/proxy/dao"
	"github.com/twbworld/proxy/model"

	"github.com/gin-gonic/gin"
)

func Init(ginServer *gin.Engine) {
	ginServer.StaticFile("/favicon.ico", "static/favicon.ico")
	ginServer.LoadHTMLGlob("static/*.html")

	ginServer.NoRoute(func(ctx *gin.Context) {
		//内部重定向
		ctx.Request.URL.Path = "/404.html"
		ginServer.HandleContext(ctx)
		//http重定向
		// ctx.Redirect(http.StatusMovedPermanently, "/404.html")
	})
	ginServer.GET("/404.html", func(ctx *gin.Context) {
		ctx.HTML(200, "404.html", gin.H{"status": "404"})
	})
	ginServer.GET("/40x.html", func(ctx *gin.Context) {
		ctx.HTML(200, "404.html", gin.H{"status": "40x"})
	})
	ginServer.GET("/50x.html", func(ctx *gin.Context) {
		ctx.HTML(200, "404.html", gin.H{"status": "50x"})
	})

	// gin路由不支持正则,我服了
	regexRouter := ginregex.New(ginServer, nil)
	regexRouter.GET(`^/(.*)\.html$`, validatorUri(), controller.Index)
}


func validatorUri() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		params := regexp.MustCompile(`^/(.*)\.html$`).FindStringSubmatch(ctx.Request.URL.Path)
		userName := params[1]
		if userName == "" || len(userName) < 3 || len(userName) > 50 {
			ctx.Abort()
			ctx.Redirect(http.StatusMovedPermanently, "/404.html")
			return
		}

		var user model.Users
		err := dao.GetUsersByUserName(&user, userName)

		if err == sql.ErrNoRows || err != nil {
			ctx.Abort()
			ctx.Redirect(http.StatusMovedPermanently, "/404.html")
			return
		}

		ctx.Next() //中间件处理完后往下走,也可以使用Abort()终止
	}
}
