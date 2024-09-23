package user

import (
	"github.com/gin-gonic/gin"

	"net/http"
	"regexp"

	"github.com/twbworld/proxy/model/db"
	"github.com/twbworld/proxy/service"
)

type BaseApi struct{}

var urlPattern = regexp.MustCompile(`^/(.*)\.html$`)

func (b *BaseApi) Subscribe(ctx *gin.Context) {
	user := ctx.MustGet(`user`).(*db.Users)

	protocol := ""
	if len(ctx.Request.Host) >= 5 {
		protocol = ctx.Request.Host[:5]
	}
	res := service.Service.UserServiceGroup.SetProtocol(protocol).Handle(user)
	if res == "" && protocol != "" {
		res = service.Service.UserServiceGroup.SetProtocol("").Handle(user)
	}

	ctx.String(http.StatusOK, res)
}
