package router

import (
	"database/sql"
	"net/http"

	"github.com/twbworld/proxy/controller"
	"github.com/twbworld/proxy/dao"
	"github.com/twbworld/proxy/model"

	"github.com/gin-gonic/gin"
)

func Init(ginServer *gin.Engine) {

	ginServer.Use(gin.Recovery())

	ginServer.StaticFile("/favicon.ico", "static/favicon.ico")


	ginServer.NoRoute(func(ctx *gin.Context) {
        // 实现内部重定向
        ctx.Redirect(http.StatusMovedPermanently, "/404.html")
    })

	//nginx: rewrite ^/(.*)\.html$ /index?u=$1 break;
	ginServer.GET("/index", validatorUri(), controller.Index)
}

func validatorUri() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userName := ctx.Query("u")
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
