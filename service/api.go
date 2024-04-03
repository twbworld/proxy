package service

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/twbworld/proxy/config"
	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/model"
	"github.com/twbworld/proxy/utils"
)

type xray struct{}
type clash struct{}
type class interface {
	Handle(user *model.Users) string
}

type clashVlessVisionReality struct {
	*config.Proxy
	WsOpts         int `json:"-"` //用于覆盖上面的属性
	Alpn           int `json:"-"`
	SkipCertVerify int `json:"-"`
}
type clashVlessVision struct {
	*config.Proxy
	WsOpts      int `json:"-"`
	RealityOpts int `json:"-"`
}
type clashVlessWs struct {
	*config.Proxy
	RealityOpts int `json:"-"`
}
type clashTrojanWs struct {
	*config.Proxy
	RealityOpts int    `json:"-"`
	Password    string `json:"password"`
}
type clashTrojan struct {
	*config.Proxy
	RealityOpts int    `json:"-"`
	WsOpts      int    `json:"-"`
	Uuid        string `json:"-" info:"用户ID或trojan的password"`
	Password    string `json:"password"`
}

func SetProtocol(t string) class {
	switch t {
	case "clash":
		return clash{}
	default:
		return xray{}
	}
}

func (clash) getConfig(value *config.Proxy) any {
	setDefaultValue(value)
	if value.Type == "vless" && value.Network == "ws" && value.WsOpts.Path != "" && value.Flow == "" {
		// VLESS-TCP-TLS-WS
		return clashVlessWs{
			Proxy: value,
		}
	} else if value.Type == "vless" && value.Flow == "xtls-rprx-vision" && value.RealityOpts.PublicKey != "" {
		// VLESS-TCP-XTLS-Vision-REALITY
		return clashVlessVisionReality{
			Proxy: value,
		}
	} else if value.Type == "vless" && value.Flow == "xtls-rprx-vision" {
		// VLESS-TCP-XTLS-Vision
		return clashVlessVision{
			Proxy: value,
		}
	} else if value.Type == "trojan" && value.Network == "ws" && value.WsOpts.Path != "" && value.Flow == "" {
		// TROJAN-TCP-TLS-WS
		trojan := clashTrojanWs{
			Proxy: value,
		}
		trojan.Password = value.Uuid
		return trojan
	} else if value.Type == "trojan" {
		// TROJAN-TCP-TLS
		trojan := clashTrojan{
			Proxy: value,
		}
		trojan.Password = value.Uuid
		return trojan
	}
	return nil
}

func (xray) getConfig(value *config.Proxy) string {
	setDefaultValue(value)

	if value.Type == "" || value.Server == "" || value.Port == "" {
		return ""
	}

	var link strings.Builder //比"+"拼接省资源

	link.WriteString(value.Type)

	if value.Type == "vless" || value.Type == "trojan" {
		link.WriteString("://")
		link.WriteString(value.Uuid)
	}

	link.WriteString(value.Uuid)
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

	if (value.Type == "vless" || value.Type == "trojan") && value.Network == "ws" && value.WsOpts.Path != "" && value.Flow == "" {
		// VLESS-TCP-TLS-WS || TROJAN-TCP-TLS-WS
		link.WriteString("&alpn=")
		link.WriteString(strings.Join(value.Alpn, ","))
		link.WriteString("&host=")
		link.WriteString(value.WsOpts.Headers.Host)
		link.WriteString("&path=")
		link.WriteString(value.WsOpts.Path)
		link.WriteString("&security=tls")
	} else if value.Type == "vless" && value.Flow == "xtls-rprx-vision" && value.RealityOpts.PublicKey != "" {
		// VLESS-TCP-XTLS-Vision-REALITY
		link.WriteString("&flow=")
		link.WriteString(value.Flow)
		link.WriteString("&pbk=")
		link.WriteString(value.RealityOpts.PublicKey)
		link.WriteString("&sid=")
		link.WriteString(value.RealityOpts.ShortId)
		link.WriteString("&security=reality")
	} else if value.Type == "vless" && value.Flow == "xtls-rprx-vision" {
		// VLESS-TCP-XTLS-Vision
		link.WriteString("&alpn=")
		link.WriteString(strings.Join(value.Alpn, ","))
		link.WriteString("&flow=")
		link.WriteString(value.Flow)
		link.WriteString("&security=tls")
	} else if value.Type == "trojan" {
		// TROJAN-TCP-TLS
		link.WriteString("&alpn=")
		link.WriteString(strings.Join(value.Alpn, ","))
		link.WriteString("&security=tls")
	}
	link.WriteString("#")
	link.WriteString(value.Name)

	return link.String()
}

func (c clash) Handle(user *model.Users) string {

	if !checkUser(user) {
		return `proxies: [{name: "!!! 订阅已过期 !!!", type: trojan, server: cn.bing.com, port: 80, password: 0, network: tcp}]
proxy-groups: [{name: "!!!!!! 订阅已过期 !!!!!!", type: select, proxies: ["!!! 订阅已过期 !!!"]}, {name: "🎯 全球直连", type: select, proxies: ["!!! 订阅已过期 !!!"]}]`
	}

	if len(global.Config.Env.Proxy) < 1 {
		return ""
	}

	f, err := os.ReadFile(global.Config.AppConfig.ClashPath)
	if err != nil || len(f) < 1 {
		return ""
	}

	proxiesName := make([]string, 0, len(global.Config.Env.Proxy))
	var proxies strings.Builder

	for _, value := range global.Config.Env.Proxy {
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

		proxies.WriteString(string(b))
		proxies.WriteString(",")
		proxiesName = append(proxiesName, value.Name)
	}

	if len(proxiesName) < 1 {
		return ""
	}

	bn, err := json.Marshal(proxiesName)
	if err != nil {
		return ""
	}

	replacer := strings.NewReplacer(`%proxies%`, "["+strings.Trim(proxies.String(), ",")+"]", `%proxies_name%`, string(bn))

	return replacer.Replace(string(f))
}

func (x xray) Handle(user *model.Users) string {
	if !checkUser(user) {
		return utils.Base64Encode("vless://0@cn.bing.com:80?type=tcp#!!! 订阅已过期 !!!")
	}

	if len(global.Config.Env.Proxy) < 1 {
		return ""
	}
	var subscription strings.Builder
	for _, value := range global.Config.Env.Proxy {
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

// 检测过期
func checkUser(user *model.Users) bool {
	if *user.ExpiryDate == "" || *user.ExpiryDate == "0" {
		return true
	}

	tz, e := time.LoadLocation("Asia/Shanghai")
	t, err := time.ParseInLocation(time.DateOnly, *user.ExpiryDate, tz)

	return e != nil || err != nil || t.AddDate(0, 0, 1).After(time.Now().In(tz))
}

func setDefaultValue(value *config.Proxy) {
	domain := value.Server
	if value.WsOpts.Headers.Host != "" {
		//如果套cdn,则避免host不等于server(使用了优选ip)
		domain = value.WsOpts.Headers.Host
	}
	value.Name = "外网信息复杂_理智分辨真假_" + value.Server + "_" + value.Port
	value.Tls = true
	value.Udp = true
	value.SkipCertVerify = false
	value.ClientFingerprint = "chrome"
	value.Alpn = []string{"h2", "http/1.1"}
	value.Sni = domain
	value.WsOpts.Headers.Host = domain
}
