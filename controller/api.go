package controller

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/twbworld/proxy/dao"
	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/model"

	"github.com/twbworld/proxy/service"

	"github.com/gin-gonic/gin"
)

func Subscribe(ctx *gin.Context) {
	// ctx.Header("content-type", "text/html; charset=UTF-8")

	params := regexp.MustCompile(`^/(.*)\.html$`).FindStringSubmatch(ctx.Request.URL.Path)
	var user model.Users
	dao.GetUsersByUserName(&user, params[1])

	res := service.SetProtocol(ctx.Request.Host[:5]).Handle(&user)
	if res == "" {
		res = service.SetProtocol("").Handle(&user)
	}

	ctx.String(http.StatusOK, res)
}

func Tg(ctx *gin.Context) {
	defer func() {
		if p := recover(); p != nil {
			global.Log.Errorln(p)
			service.TgSend(`系统错误, 请按"/start"重新设置`)
			errMsg, _ := json.Marshal(map[string]string{"error": "系统出错"})
			ctx.Writer.WriteHeader(http.StatusBadRequest)
			ctx.Writer.Header().Set("Content-Type", "application/json")
			ctx.Writer.Write(errMsg)
		}
		service.IsTgSend = false
	}()

	service.IsTgSend = false

	if err := service.Webhook(ctx); err != nil {
		panic(err)
	}

}
