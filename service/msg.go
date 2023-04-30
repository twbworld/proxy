package service

import (
	"fmt"
	"github.com/twbworld/proxy/global"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TgSend(text string) {
	if text == "" || global.Config.Env.Debug || global.Config.Env.Telegram.Token == ""{
		return
	}
	bot, err := tg.NewBotAPI(global.Config.Env.Telegram.Token)
	if err != nil {
		global.Log.Warn("通知初始化失败[iowhei]: ", err)
		return
	}
	bot.Debug = global.Config.Env.Debug
	msg := tg.NewMessage(global.Config.Env.Telegram.Id, fmt.Sprintf("[%s]%s", global.Config.AppConfig.ProjectName, text))
	if _, err := bot.Send(msg); err != nil {
		global.Log.Warn("发送通知失败[oiuj0fasd]: ", err)
		return
	}
}
