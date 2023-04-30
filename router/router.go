package router

import (
	"database/sql"
	"log"
	"github.com/twbworld/proxy/controller"
	"github.com/twbworld/proxy/dao"
	"github.com/twbworld/proxy/model"

	"github.com/gin-gonic/gin"
)

func Init(ginServer *gin.Engine) {

	ginServer.Use(gin.Recovery())

	ginServer.StaticFile("/favicon.ico", "static/favicon.ico")

	//nginx: rewrite ^/(.*)\.html$ /index?u=$1 break;
	ginServer.GET("/index", validatorUri(), controller.Index)
}

func validatorUri() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userName := ctx.Query("u")
		if userName == "" || len(userName) < 3 || len(userName) > 50 {
			ctx.Abort()
			return
		}

		var user model.Users
		err := dao.DB.Get(&user, "SELECT * FROM `users` WHERE `username`=?", userName)
		if err == sql.ErrNoRows {
			ctx.Abort()
			return
		} else if err != nil {
			ctx.Abort()
			log.Println("查询出错: ", err)
			return
		}

		ctx.Next() //中间件处理完后往下走,也可以使用Abort()终止
	}
}
