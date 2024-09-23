package middleware

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/twbworld/proxy/dao"
	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/model/db"
)

var urlPattern = regexp.MustCompile(`^/(.*)\.html$`)

// 验证TG的token
func ValidatorTgToken(ctx *gin.Context) {
	token := ctx.Param("token")
	if !global.Config.Debug && global.Config.Telegram.Token == token {
		//业务前执行
		ctx.Next()
		//业务后执行
		return
	}

	ctx.Abort()

	ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "token"})
	ctx.Redirect(http.StatusMovedPermanently, "/404.html")
}

func ValidatorSubscribe(ctx *gin.Context) {
	params := urlPattern.FindStringSubmatch(ctx.Request.URL.Path)
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

	var user db.Users
	if dao.App.UsersDb.GetUsersByUserName(&user, userName) != nil {
		ctx.Abort()
		ctx.Redirect(http.StatusMovedPermanently, "/404.html")
		return
	}

	ctx.Set(`user`, &user)

	ctx.Next() //中间件处理完后往下走,也可以使用Abort()终止
}
