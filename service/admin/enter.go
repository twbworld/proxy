package admin

import (
	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/service/common"
)

type ServiceGroup struct {
	BotService BotService
}

func NewServiceGroup(async common.AsyncService, notifier global.NotifierService) ServiceGroup {
	return ServiceGroup{
		BotService: NewBotService(async, notifier),
	}
}
