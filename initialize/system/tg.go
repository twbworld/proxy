package system

import (
	"fmt"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/twbworld/proxy/global"
)

func tgStart() error {
	if global.Config.Debug || len(global.Config.Telegram.Token) < 1 {
		global.Log.Warnln("Telegram服务未启动: Debug模式或Token为空")
		return nil
	}

	bot, err := tg.NewBotAPI(global.Config.Telegram.Token)
	if err != nil {
		return fmt.Errorf("bot初始化失败: %w", err)
	}
	global.Bot = bot
	global.Bot.Debug = global.Config.Debug

	setCommands := tg.NewSetMyCommands(tg.BotCommand{
		Command:     "start",
		Description: "开始",
	})
	if _, err := global.Bot.Request(setCommands); err != nil {
		return fmt.Errorf("设置Command失败[ofoan]: %w", err)
	}

	if global.Config.Domain != "" {
		wh, _ := tg.NewWebhook(fmt.Sprintf(`%s/wh/tg/%s`, global.Config.Domain, global.Bot.Token))
		if _, err = global.Bot.Request(wh); err != nil {
			return fmt.Errorf("设置webhook失败: %w", err)
		}

		info, err := global.Bot.GetWebhookInfo()
		if err != nil {
			return fmt.Errorf("获取webhook失败: %w", err)
		}

		if info.LastErrorDate != 0 {
			return fmt.Errorf("获取tg信息错误[9e0rtji]: %s", info.LastErrorMessage)
		}
		global.Log.Printf("成功配置tg[doiasjo]: %s", global.Bot.Self.UserName)
	}
	return nil
}

func tgClear() error {
	if global.Bot == nil {
		global.Log.Warnln("Telegram服务未启动或已清理")
		return nil
	}
	_, err := global.Bot.Request(tg.DeleteWebhookConfig{})
	if err != nil {
		return fmt.Errorf("删除webhook失败[wtuina]: %w", err)
	}
	global.Log.Infoln("Telegram服务清理成功")
	return nil
}
