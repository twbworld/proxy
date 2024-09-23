package controller

import "github.com/twbworld/proxy/controller/user"
import "github.com/twbworld/proxy/controller/admin"

var Api = new(ApiGroup)

type ApiGroup struct {
	UserApiGroup  user.ApiGroup
	AdminApiGroup admin.ApiGroup
}
