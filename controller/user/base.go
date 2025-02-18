package user

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"net/http"
	"net/url"

	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/model/db"
	"github.com/twbworld/proxy/service"
	"github.com/twbworld/proxy/utils"
)

type BaseApi struct{}

func (b *BaseApi) Subscribe(ctx *gin.Context) {
	user := ctx.MustGet(`user`).(*db.Users)

	//获取最后一级子域名名称
	protocol := strings.Split(ctx.Request.Host, ".")[0]
	res := service.Service.UserServiceGroup.SetProtocol(protocol).Handle(user)
	if res == "" && protocol != "" {
		res = service.Service.UserServiceGroup.SetProtocol("").Handle(user)
	}

	//https://www.clashverge.dev/guide/url_schemes.html
	if utils.ContainsAny(ctx.Request.UserAgent(), []string{"clash", "v2ray"}) {
		fileName := global.Config.Subscribe.Filename
		if fileName == "" {
			fileName = user.Username
		}
		if fileName != "" {
			ctx.Header("content-disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.QueryEscape(fileName)))
		}
		if global.Config.Subscribe.UpdateInterval > 0 {
			ctx.Header("profile-update-interval", strconv.Itoa(int(global.Config.Subscribe.UpdateInterval)))
		}
		if global.Config.Subscribe.PageUrl != "" {
			ctx.Header("profile-web-page-url", global.Config.Subscribe.PageUrl)
		}
		if *user.ExpiryDate != "" {
			t, err := time.ParseInLocation(time.DateOnly, *user.ExpiryDate, global.Tz)
			if err == nil {
				//暂不支持流量获取
				ctx.Header("subscription-userinfo", fmt.Sprintf("upload=0; download=0; total=0; expire=%d", t.Unix()))
			}
		}
	}

	ctx.String(http.StatusOK, res)
}
