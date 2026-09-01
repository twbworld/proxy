package controller

import "github.com/twbworld/proxy/controller/user"
import "github.com/twbworld/proxy/controller/admin"

var (
	User  user.ApiGroup  //用户前台
	Admin admin.ApiGroup //管理后台
)
