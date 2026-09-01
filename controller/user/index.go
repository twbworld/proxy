package user

import (
	"cmp"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/model/db"
	"github.com/twbworld/proxy/service"
	"github.com/twbworld/proxy/utils"
)

type IndexApi struct{}

func (b *IndexApi) Subscribe(ctx *gin.Context) {
	user := ctx.MustGet(`user`).(*db.Users)

	// 获取最后一级子域名名称作为协议
	protocol, _, _ := strings.Cut(ctx.Request.Host, ".")

	res := service.User.IndexService.Handle(protocol, user)

	if utils.ContainsAny(ctx.Request.UserAgent(), []string{"clash", "v2ray"}) {
		fileName := cmp.Or(global.Config().Subscribe.Filename, user.Username)
		if fileName != "" {
			ctx.Header("content-disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.QueryEscape(fileName)))
		}
		if global.Config().Subscribe.UpdateInterval > 0 {
			ctx.Header("profile-update-interval", strconv.Itoa(int(global.Config().Subscribe.UpdateInterval)))
		}
		if global.Config().Subscribe.PageUrl != "" {
			ctx.Header("profile-web-page-url", global.Config().Subscribe.PageUrl)
		}
		if user.ExpiryDate != "" {
			if t, err := time.ParseInLocation(time.DateOnly, user.ExpiryDate, time.UTC); err == nil {
				ctx.Header("subscription-userinfo", fmt.Sprintf("upload=0; download=0; total=0; expire=%d", t.Unix()))
			}
		}
	}

	ctx.String(http.StatusOK, res)
}
