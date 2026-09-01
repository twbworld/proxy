package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/service"
)

type BotApi struct{}

func (t *BotApi) Webhook(ctx *gin.Context) {
	defer func() {
		if p := recover(); p != nil {
			global.Log.Errorln(p)
			global.Notifier.Send(`系统错误, 请按"/start"重新设置`)
			// 兜底返回 200, 避免 TG 官方服务器无限重试
			ctx.Status(http.StatusOK)
		}
	}()

	if err := service.Admin.BotService.Webhook(ctx); err != nil {
		global.Log.Errorf("处理 TG Webhook 请求失败: %v", err)
	}

	ctx.Status(http.StatusOK)
}
