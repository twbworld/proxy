package service

import (
	"github.com/twbworld/proxy/service/admin"
	"github.com/twbworld/proxy/service/common"
	"github.com/twbworld/proxy/service/user"
)

var (
	User   user.ServiceGroup
	Admin  admin.ServiceGroup
	Common common.ServiceGroup
)
