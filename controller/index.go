package controller

import (
	"encoding/json"
	"github.com/twbworld/proxy/dao"
	"github.com/twbworld/proxy/model"
	"net/http"
	"regexp"

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
	if err := service.TgWebhookHandle(ctx); err != nil {
		errMsg, _ := json.Marshal(map[string]string{"error": err.Error()})
		ctx.Writer.WriteHeader(http.StatusBadRequest)
		ctx.Writer.Header().Set("Content-Type", "application/json")
		_, _ = ctx.Writer.Write(errMsg)
		return
	}
}
