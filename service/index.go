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

type clashVlessVisionReality struct {
	*config.Proxy
	WsOpts         int `json:"ws-opts,omitempty" mapstructure:"ws-opts"` //用于覆盖前面的WsOpts以隐藏json的ws-opts
	Alpn           int `json:"alpn,omitempty" mapstructure:"alpn"`
	SkipCertVerify int `json:"skip-cert-verify,omitempty" mapstructure:"skip-cert-verify"`
}
type clashVlessVision struct {
	*config.Proxy
	WsOpts      int `json:"ws-opts,omitempty" mapstructure:"ws-opts"`
	RealityOpts int `json:"reality-opts,omitempty" mapstructure:"reality-opts"`
}
type clashVlessWs struct {
	*config.Proxy
	RealityOpts int `json:"reality-opts,omitempty" mapstructure:"reality-opts"`
}

//检测过期
func checkUser(user *model.Users) bool {
	if *user.ExpiryDate == "" || *user.ExpiryDate == "0" {
		return true
	}

	tz, e := time.LoadLocation("Asia/Shanghai")
	t, err := time.Parse(time.DateOnly, *user.ExpiryDate)

	return e != nil || err != nil || t.AddDate(0, 0, 1).After(time.Now().In(tz))
}

func setDefaultValue(value *config.Proxy) {
	domain := value.Server
	if value.WsOpts.Headers.Host != "" {
		//如果套cdn,则避免host不等于server(使用了优选ip)
		domain = value.WsOpts.Headers.Host
	}
	value.Name = "外网信息复杂_理智分辨真假" + "_" + domain + "_" + value.Port
	value.Tls = true
	value.Udp = true
	value.SkipCertVerify = false
	value.ClientFingerprint = "chrome"
	value.Alpn = []string{"h2", "http/1.1"}
	value.Sni = domain
	value.WsOpts.Headers.Host = domain
}

func getClashConfig(value *config.Proxy) any {
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
	} else if value.Type == "trojan" {
		// trojan
		return nil
	}
	return nil
}

func getConfig(value *config.Proxy) string {
	setDefaultValue(value)

	if value.Type == "" || value.Server == "" || value.Port == "" {
		return ""
	}

	link := ""
	link += value.Type
	if value.Type == "vless" {
		link += "://" + value.Uuid
	} else if value.Type == "trojan" {
		link += "://" + value.Uuid
	}
	link += "@" + value.Server
	link += ":" + value.Port
	link += "?encryption=none"
	link += "&headerType=none"
	link += "&sni=" + value.Sni
	link += "&fp=" + value.ClientFingerprint
	link += "&type=" + value.Network

	if value.Type == "vless" && value.Network == "ws" && value.WsOpts.Path != "" && value.Flow == "" {
		// VLESS-TCP-TLS-WS
		link += "&alpn=" + strings.Join(value.Alpn, ",")
		link += "&host=" + value.WsOpts.Headers.Host
		link += "&path=" + value.WsOpts.Path
		link += "&security=tls"
	} else if value.Type == "vless" && value.Flow == "xtls-rprx-vision" && value.RealityOpts.PublicKey != "" {
		// VLESS-TCP-XTLS-Vision-REALITY
		link += "&flow=" + value.Flow
		link += "&pbk=" + value.RealityOpts.PublicKey
		link += "&sid=" + value.RealityOpts.ShortId
		link += "&security=reality"
	} else if value.Type == "vless" && value.Flow == "xtls-rprx-vision" {
		// VLESS-TCP-XTLS-Vision
		link += "&alpn=" + strings.Join(value.Alpn, ",")
		link += "&flow=" + value.Flow
		link += "&security=tls"
	} else if value.Type == "trojan" {
		// trojan
		return ""
	}

	return link + "#" + value.Name
}

func ClashHandle(user *model.Users) string {

	if !checkUser(user) {
		return `proxies: [{name: "!!! 订阅已过期 !!!", type: trojan, server: cn.bing.com, port: 80, password: 0, network: tcp}]
proxy-groups: [{name: "!!!!!! 订阅已过期 !!!!!!", type: select, proxies: ["!!! 订阅已过期 !!!"]}, {name: "🎯 全球直连", type: select, proxies: ["!!! 订阅已过期 !!!"]}]`
	}

	subscription := ""
	if len(global.Config.Env.Proxy) < 1 {
		return subscription
	}

	f, err := os.ReadFile(global.Config.AppConfig.ClashPath)
	if err != nil || len(f) < 1 {
		return subscription
	}

	proxies := ""
	proxiesName := []string{}

	for _, value := range global.Config.Env.Proxy {
		if value.Server == "" || value.Type == "" || (value.Root && user.Quota != -1) {
			continue
		}

		combinationType := getClashConfig(&value)
		if combinationType == nil {
			continue
		}

		b, e := json.Marshal(combinationType)
		if e != nil || b == nil {
			continue
		}

		proxies += string(b) + ","
		proxiesName = append(proxiesName, value.Name)
	}

	if len(proxiesName) < 1 {
		return subscription
	}

	bn, err := json.Marshal(proxiesName)
	if err != nil {
		return subscription
	}

	replacer := strings.NewReplacer(`%proxies%`, "["+strings.Trim(proxies, ",")+"]", `%proxies_name%`, string(bn))
	subscription += replacer.Replace(string(f))

	return subscription

}

func XrayHandle(user *model.Users) string {
	if !checkUser(user) {
		return utils.Base64Encode("vless://0@cn.bing.com:80?type=tcp#!!! 订阅已过期 !!!")
	}

	subscription := ""
	if len(global.Config.Env.Proxy) < 1 {
		return subscription
	}

	for _, value := range global.Config.Env.Proxy {
		if value.Server == "" || value.Type == "" || (value.Root && user.Quota != -1) {
			continue
		}
		if link := getConfig(&value); link != "" {
			subscription += link + "\n"
		}
	}

	return utils.Base64Encode(subscription)
}
