package telegram

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

// Config 定义 TG 服务所需的配置项
type Config struct {
	Token       string // Telegram Bot Token
	Domain      string // Webhook 回调域名
	ProjectName string // 项目名称(用于消息前缀标识)
	TelegramId  int64  // 接收通知的目标 TG 账号/群组 ID
	Debug       bool   // 是否开启调试模式
}

type Client struct {
	Config
	bot  *tg.BotAPI
	lock sync.RWMutex
	log  *logrus.Logger
}

// NewNoopClient 空操作的通知客户端
func NewNoopClient(log *logrus.Logger) *Client {
	return &Client{log: log}
}

// NewClient 初始化 TG 机器人并配置 Webhook，未配置或失败时降级为无操作模式
func NewClient(cfg Config, log *logrus.Logger) (*Client, error) {
	if cfg.Debug || cfg.Token == "" {
		if log != nil {
			log.Warnln("Telegram通知服务未启用: Debug模式或Token为空")
		}
		return &Client{
			log:         log,
			ProjectName: cfg.ProjectName,
			TelegramId:  cfg.TelegramId,
		}, nil
	}

	if cfg.Domain == "" {
		if log != nil {
			log.Warnln("Telegram配置缺失回调域名[Domain]，已降级为无操作模式")
		}
		return &Client{log: log, ProjectName: cfg.ProjectName, TelegramId: cfg.TelegramId}, nil
	}

	// 配置带短超时的 HTTP 客户端，避免网络受阻或不可达时长时间挂起阻塞启动
	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   2 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 2 * time.Second,
		},
		Timeout: 3 * time.Second,
	}

	bot, err := tg.NewBotAPIWithClient(cfg.Token, tg.APIEndpoint, httpClient)
	if err != nil {
		return nil, fmt.Errorf("bot初始化失败: %w", err)
	}
	bot.Debug = cfg.Debug

	setCommands := tg.NewSetMyCommands(tg.BotCommand{
		Command:     "start",
		Description: "开始",
	})
	if _, err = bot.Request(setCommands); err != nil {
		return nil, fmt.Errorf("设置Command失败: %w", err)
	}

	wh, _ := tg.NewWebhook(fmt.Sprintf("%s/wh/tg/%s", strings.TrimRight(cfg.Domain, "/"), bot.Token))
	if _, err = bot.Request(wh); err != nil {
		return nil, fmt.Errorf("设置webhook失败: %w", err)
	}

	if info, err := bot.GetWebhookInfo(); err == nil && info.LastErrorDate != 0 && log != nil {
		log.Warnf("获取tg信息告警: %s", info.LastErrorMessage)
	}

	if log != nil {
		log.Infof("成功配置TG机器人: %s", bot.Self.UserName)
	}

	return &Client{
		bot:         bot,
		log:         log,
		ProjectName: cfg.ProjectName,
		TelegramId:  cfg.TelegramId,
	}, nil
}

func (c *Client) Send(text string) error {
	if c == nil || c.bot == nil {
		if c != nil && c.log != nil {
			c.log.Debugf("Telegram未启用，跳过消息发送: %s", text)
		}
		return nil
	}
	if text == "" {
		return nil
	}

	c.lock.RLock()
	defer c.lock.RUnlock()

	var str strings.Builder
	if c.ProjectName != "" {
		str.WriteString("[")
		str.WriteString(c.ProjectName)
		str.WriteString("]")
	}
	str.WriteString(text)

	msg := tg.NewMessage(c.TelegramId, str.String())
	if _, err := c.bot.Send(msg); err != nil {
		if c.log != nil {
			c.log.Warnf("发送TG消息失败: %v", err)
		}
	}
	return nil
}

func (c *Client) SendRaw(msg tg.Chattable) error {
	if c == nil || c.bot == nil {
		if c != nil && c.log != nil {
			c.log.Debugln("Telegram未启用，跳过Raw消息发送")
		}
		return nil
	}
	if msg == nil {
		return nil
	}

	if _, err := c.bot.Send(msg); err != nil {
		if c.log != nil {
			c.log.Warnf("发送TG Raw消息失败: %v", err)
		}
	}
	return nil
}

func (c *Client) Clear() error {
	if c == nil || c.bot == nil {
		if c != nil && c.log != nil {
			c.log.Debugln("Telegram未启用或已清理")
		}
		return nil
	}

	if _, err := c.bot.Request(tg.DeleteWebhookConfig{}); err != nil {
		if c.log != nil {
			c.log.Warnf("删除TG webhook失败: %v", err)
		}
		return nil
	}

	if c.log != nil {
		c.log.Infoln("Telegram服务清理成功")
	}
	return nil
}
