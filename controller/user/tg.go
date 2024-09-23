package user

import (
	"github.com/gin-gonic/gin"

	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/service"
	"net/http"
)

type TgApi struct{}

func (t *TgApi) Tg(ctx *gin.Context) {
	defer func() {
		if p := recover(); p != nil {
			global.Log.Errorln(p)
			service.Service.UserServiceGroup.TgSend(`系统错误, 请按"/start"重新设置`)
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "系统错误"})
		}
	}()

	if err := service.Service.UserServiceGroup.Webhook(ctx); err != nil {
		panic(err)
	}

}
