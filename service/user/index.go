package user

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/model/common"
	"github.com/twbworld/proxy/model/config"
	"github.com/twbworld/proxy/model/db"
	"github.com/twbworld/proxy/utils"
)

type BaseService struct{}

type xray struct{}
type clash struct{}
type class interface {
	Handle(user *db.Users) string
}

func (b *BaseService) SetProtocol(t string) class {
	switch t {
	case "clash":
		return &clash{}
	default:
		return &xray{}
	}
}

func (c *clash) Handle(user *db.Users) string {

	if !checkUser(user) {
		return `proxies:
  - {name: "!!! 订阅已过期 !!!", type: trojan, server: cn.bing.com, port: 80, password: 0, network: tcp}
proxy-groups:
  - {name: "!!!!!! 订阅已过期 !!!!!!", type: select, proxies: ["!!! 订阅已过期 !!!"]}
  - {name: "🎯 全球直连", type: select, proxies: ["!!! 订阅已过期 !!!"]}`
	}

	if len(global.Config.Proxy) < 1 || !utils.FileExist(global.Config.ClashPath) {
		return ""
	}

	proxiesName := make([]string, 0, len(global.Config.Proxy))
	var proxies strings.Builder

	for _, value := range global.Config.Proxy {
		if value.Server == "" || value.Type == "" || (value.Root && user.Quota != -1) {
			continue
		}

		combinationType := c.getConfig(&value)
		if combinationType == nil {
			continue
		}

		b, e := json.Marshal(combinationType)
		if e != nil || b == nil {
			continue
		}

		proxies.WriteString("\n  - ") //yaml格式
		proxies.Write(b)
		proxiesName = append(proxiesName, value.Name)
	}

	if len(proxiesName) < 1 {
		return ""
	}

	bn, err := json.Marshal(proxiesName)
	if err != nil {
		return ""
	}

	fres, err := os.ReadFile(global.Config.ClashPath)
	if err != nil || len(fres) < 1 {
		return ""
	}

	replacer := strings.NewReplacer(` [proxies]`, proxies.String(), `[proxies_name]`, string(bn))

	return replacer.Replace(string(fres))
}

func (x *xray) Handle(user *db.Users) string {
	if !checkUser(user) {
		return utils.Base64Encode("vless://0@cn.bing.com:80?type=tcp#!!! 订阅已过期 !!!")
	}

	if len(global.Config.Proxy) < 1 {
		return ""
	}
	var subscription strings.Builder
	for _, value := range global.Config.Proxy {
		if value.Server == "" || value.Type == "" || (value.Root && user.Quota != -1) {
			continue
		}
		if link := x.getConfig(&value); link != "" {
			subscription.WriteString(link)
			subscription.WriteString("\n")
		}
	}

	return utils.Base64Encode(subscription.String())
}

func (c *clash) getConfig(value *config.Proxy) any {
	value.SetProxyDefault()

	switch {
	case value.Type == "vless" && value.Network == "ws" && value.WsOpts.Path != "" && value.Flow == "":
		// VLESS-WS-TLS
		return common.ClashVlessWs{Proxy: value}
	case value.Type == "vless" && value.Flow == "xtls-rprx-vision" && value.RealityOpts.PublicKey != "":
		// VLESS-TCP-XTLS-Vision-REALITY
		return common.ClashVlessVisionReality{Proxy: value}
	case value.Type == "vless" && value.Flow == "xtls-rprx-vision":
		// VLESS-TCP-XTLS-Vision
		return common.ClashVlessVision{Proxy: value}
	case value.Type == "trojan" && value.Network == "ws" && value.WsOpts.Path != "" && value.Flow == "":
		// TROJAN-WS-TLS
		trojan := common.ClashTrojanWs{Proxy: value}
		trojan.Password = value.Uuid
		return trojan
	case value.Type == "trojan":
		// TROJAN-TCP-TLS
		trojan := common.ClashTrojan{Proxy: value}
		trojan.Password = value.Uuid
		return trojan
	default:
		return nil
	}
}

func (x *xray) getConfig(value *config.Proxy) string {
	value.SetProxyDefault()

	if value.Type == "" || value.Server == "" || value.Port == "" {
		return ""
	}

	var link strings.Builder
	link.WriteString(value.Type)
	if value.Type == "vless" || value.Type == "trojan" {
		link.WriteString("://")
		link.WriteString(value.Uuid)
	}

	link.WriteString("@")
	link.WriteString(value.Server)
	link.WriteString(":")
	link.WriteString(value.Port)
	link.WriteString("?encryption=none&headerType=none&sni=")
	link.WriteString(value.Sni)
	link.WriteString("&fp=")
	link.WriteString(value.ClientFingerprint)
	link.WriteString("&type=")
	link.WriteString(value.Network)

	switch {
	case (value.Type == "vless" || value.Type == "trojan") && value.Network == "ws" && value.WsOpts.Path != "" && value.Flow == "":
		// VLESS-WS-TLS || TROJAN-WS-TLS
		link.WriteString("&alpn=")
		link.WriteString(strings.Join(value.Alpn, ","))
		link.WriteString("&host=")
		link.WriteString(value.WsOpts.Headers.Host)
		link.WriteString("&path=")
		link.WriteString(value.WsOpts.Path)
		link.WriteString("&security=tls")
	case value.Type == "vless" && value.Flow == "xtls-rprx-vision" && value.RealityOpts.PublicKey != "":
		// VLESS-TCP-XTLS-Vision-REALITY
		link.WriteString("&flow=")
		link.WriteString(value.Flow)
		link.WriteString("&pbk=")
		link.WriteString(value.RealityOpts.PublicKey)
		link.WriteString("&sid=")
		link.WriteString(value.RealityOpts.ShortId)
		link.WriteString("&security=reality")
	case value.Type == "vless" && value.Flow == "xtls-rprx-vision":
		// VLESS-TCP-XTLS-Vision
		link.WriteString("&alpn=")
		link.WriteString(strings.Join(value.Alpn, ","))
		link.WriteString("&flow=")
		link.WriteString(value.Flow)
		link.WriteString("&security=tls")
	case value.Type == "trojan":
		// TROJAN-TCP-TLS
		link.WriteString("&alpn=")
		link.WriteString(strings.Join(value.Alpn, ","))
		link.WriteString("&security=tls")
	}
	link.WriteString("#")
	link.WriteString(value.Name)

	return link.String()
}

// 检测过期
func checkUser(user *db.Users) bool {
	if *user.ExpiryDate == "" || *user.ExpiryDate == "0" {
		return true
	}

	t, err := time.ParseInLocation(time.DateOnly, *user.ExpiryDate, global.Tz)

	return err != nil || t.AddDate(0, 0, 1).After(time.Now().In(global.Tz))
}
