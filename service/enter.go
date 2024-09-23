package service

import "github.com/twbworld/proxy/service/user"
import "github.com/twbworld/proxy/service/admin"

var Service = new(ServiceGroup)

type ServiceGroup struct {
	UserServiceGroup  user.ServiceGroup
	AdminServiceGroup admin.ServiceGroup
}
