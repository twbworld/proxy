package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/twbworld/proxy/dao"
	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/model/common"
	"github.com/twbworld/proxy/model/db"
)

func ValidatorBotToken(ctx *gin.Context) {
	token := ctx.Param("token")
	// Debug 模式下允许任意非空测试 Token
	if (global.Config().Debug && token != "") || (global.Config().Telegram.Token != "" && global.Config().Telegram.Token == token) {
		//业务前执行
		ctx.Next()
		//业务后执行
		return
	}

	common.FailAuth(ctx, "认证失败: 无效的令牌")
	ctx.Abort()
}

func ValidatorSubscribe(ctx *gin.Context) {
	userName := ctx.GetString("userName")
	if userName == "" || len(userName) < 3 || len(userName) > 50 {
		ctx.Abort()
		ctx.Redirect(http.StatusMovedPermanently, "/404.html")
		return
	}

	user := new(db.Users)
	if dao.App.GetByCol(user, "username", userName) != nil {
		ctx.Abort()
		ctx.Redirect(http.StatusMovedPermanently, "/404.html")
		return
	}

	ctx.Set(`user`, user)

	ctx.Next() //中间件处理完后往下走,也可以使用Abort()终止
}
