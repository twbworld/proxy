package router

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/jxskiss/ginregex"
	"github.com/twbworld/proxy/controller"
	"github.com/twbworld/proxy/dao"
	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/model"

	"github.com/gin-gonic/gin"
)

func Start(ginServer *gin.Engine) {
	ginServer.StaticFile("/favicon.ico", "static/favicon.ico")
	ginServer.StaticFile("/robots.txt", "static/robots.txt")
	ginServer.LoadHTMLGlob("static/*.html")

	ginServer.NoRoute(func(ctx *gin.Context) {
		//内部重定向
		ctx.Request.URL.Path = "/404.html"
		ginServer.HandleContext(ctx)
		//http重定向
		// ctx.Redirect(http.StatusMovedPermanently, "/404.html")
	})
	ginServer.GET("404.html", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "404.html", gin.H{"status": "404"})
	})
	ginServer.GET("40x.html", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "404.html", gin.H{"status": "40x"})
	})
	ginServer.GET("50x.html", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "404.html", gin.H{"status": "50x"})
	})
	ginServer.POST("404.html", func(ctx *gin.Context) {
		ctx.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "错误[fhua]"})
	})
	ginServer.POST("40x.html", func(ctx *gin.Context) {
		ctx.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "错误[fhua]"})
	})
	ginServer.POST("50x.html", func(ctx *gin.Context) {
		ctx.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "错误[fhua]"})
	})

	wh := ginServer.Group("wh")
	{
		wh.POST("/tg/:token", validatorTgToken(), controller.Tg)
	}

	// gin路由不支持正则,我服了
	regexRouter := ginregex.New(ginServer, nil)
	regexRouter.GET(`^/(.*)\.html$`, validatorUri(), controller.Subscribe)
}

func validatorUri() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		params := regexp.MustCompile(`^/(.*)\.html$`).FindStringSubmatch(ctx.Request.URL.Path)
		if len(params) < 2 {
			ctx.Abort()
			ctx.Redirect(http.StatusMovedPermanently, "/404.html")
			return
		}
		userName := params[1]
		if userName == "" || len(userName) < 3 || len(userName) > 50 {
			ctx.Abort()
			ctx.Redirect(http.StatusMovedPermanently, "/404.html")
			return
		}

		var user model.Users
		if err := dao.GetUsersByUserName(&user, userName); err == sql.ErrNoRows || err != nil {
			ctx.Abort()
			ctx.Redirect(http.StatusMovedPermanently, "/404.html")
			return
		}

		ctx.Next() //中间件处理完后往下走,也可以使用Abort()终止
	}
}

func validatorTgToken() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.Param("token")
		if !global.Config.Env.Debug && global.Config.Env.Telegram.Token == token {
			ctx.Next()
			return
		}

		ctx.Abort()
		if global.Config.Env.Telegram.Token != token {
			errMsg, _ := json.Marshal(map[string]string{"error": "Lack of token"})
			ctx.Writer.WriteHeader(http.StatusBadRequest)
			ctx.Writer.Header().Set("Content-Type", "application/json")
			_, _ = ctx.Writer.Write(errMsg)
		} else {
			ctx.Redirect(http.StatusMovedPermanently, "/404.html")
		}
	}
}
