package controller

import (
	"net/http"
	"os"
	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/service"

	"github.com/gin-gonic/gin"
)

func Index(ctx *gin.Context) {
	// ctx.Header("content-type", "text/html; charset=UTF-8")
	if ctx.Request.Host[:5] == "clash" {
		f, err := os.ReadFile(global.Config.AppConfig.ClashPath)
		if err == nil && len(f) > 1 {
			ctx.String(http.StatusOK, service.ClashHandle(ctx, f))
			return
		}
	}
	ctx.String(http.StatusOK, service.TrojanGoHandle(ctx))
}
