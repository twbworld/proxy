package middleware

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/twbworld/proxy/global"
)

// 跨域
func CorsHandle() gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowOrigins = global.Config.Cors
	config.AllowMethods = []string{"OPTIONS", "POST", "GET"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "authorization"}
	return cors.New(config)
}

func OptionsMethod(ctx *gin.Context) {
	if ctx.Request.Method == "OPTIONS" {
		ctx.AbortWithStatus(http.StatusNoContent)
	}
}
