package initialize

import (
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/twbworld/proxy/global"
)

func TgInit() {
	if global.Config.Env.Debug || global.Config.Env.Telegram.Token == "" {
		return
	}

	bot, err := tg.NewBotAPI(global.Config.Env.Telegram.Token)
	if err != nil {
		global.Log.Warn("bot初始化失败[jfsertyu]: ", err)
		return
	}
	global.Bot = bot
	global.Bot.Debug = global.Config.Env.Debug

	setCommands := tg.NewSetMyCommands(tg.BotCommand{
		Command:     "start",
		Description: "开始",
	})
	if _, err := global.Bot.Request(setCommands); err != nil {
		global.Log.Warn("设置Command失败[podritgfd]: ", err)
	}

	if global.Config.Env.Domain != "" {
		wh, _ := tg.NewWebhook("https://" + global.Config.Env.Domain + "/wh/tg/" + global.Bot.Token)
		if _, err = global.Bot.Request(wh); err != nil {
			global.Log.Warn("设置webhook失败[oifoghe]: ", err)
			return
		}

		info, err := global.Bot.GetWebhookInfo()
		if err != nil {
			global.Log.Warn("获取webhook失败[iuieee]: ", err)
			return
		}

		if info.LastErrorDate != 0 {
			global.Log.Warn("Telegram callback failed[fosdjfoisj]: ", info.LastErrorMessage)
			return
		}
		global.Log.Printf("成功配置tg[doiasjo]: %s", global.Bot.Self.UserName)
	}
}
