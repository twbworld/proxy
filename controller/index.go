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

func Index(ctx *gin.Context) {
	// ctx.Header("content-type", "text/html; charset=UTF-8")

	params := regexp.MustCompile(`^/(.*)\.html$`).FindStringSubmatch(ctx.Request.URL.Path)
	var user model.Users
	dao.GetUsersByUserName(&user, params[1])

	if ctx.Request.Host[:5] == "clash" {
		if res := service.ClashHandle(&user); res != "" {
			ctx.String(http.StatusOK, res)
			return
		}
	}
	ctx.String(http.StatusOK, service.XrayHandle(&user))
}

func Tg(ctx *gin.Context) {
	var err error

	defer func() {
		if p := recover(); p != nil || err != nil {
			if p != nil {
				global.Log.Errorln(p)
			} else {
				global.Log.Errorln(err)
			}
			service.TgSend(`系统错误, 请按"/start"重新设置`)
			errMsg, _ := json.Marshal(map[string]string{"error": "系统出错"})
			ctx.Writer.WriteHeader(http.StatusBadRequest)
			ctx.Writer.Header().Set("Content-Type", "application/json")
			_, _ = ctx.Writer.Write(errMsg)
		}
		service.IsTgSend = false
	}()

	service.IsTgSend = false
	err = service.TgWebhookHandle(ctx)

}
