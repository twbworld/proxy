package user

import (
	"cmp"
	"encoding/json/v2"
	"fmt"
	"maps"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/model/config"
	"github.com/twbworld/proxy/model/db"
	"github.com/twbworld/proxy/utils"
)

type IndexService interface {
	Handle(protocol string, user *db.Users) string
}

type indexService struct{}

type handler struct {
	user *db.Users
}

func NewIndexService() IndexService {
	return &indexService{}
}

func (s *indexService) Handle(protocol string, user *db.Users) string {
	handler := new(handler{
		user: user,
	})

	res := handler.generate(protocol)
	if res == "" && protocol != "" {
		res = handler.generate("")
	}
	return res
}

func (h *handler) generate(protocol string) string {
	switch protocol {
	case "clash":
		return h.buildClash()
	default:
		return h.buildV2ray()
	}
}

// buildClash 处理 Clash/Mihomo 订阅
func (h *handler) buildClash() string {
	if !h.checkUser() {
		return `proxies:
  - {name: "!!! 订阅已过期 !!!", type: trojan, server: cn.bing.com, port: 80, password: 0, network: tcp}
proxy-groups:
  - {name: "!!!!!! 订阅已过期 !!!!!!", type: select, proxies: ["!!! 订阅已过期 !!!"]}`
	}

	if len(global.Config().Proxies) < 1 || !utils.FileExist(global.Config().ClashPath) {
		return ""
	}

	proxiesName := make([]string, 0, len(global.Config().Proxies))
	var proxiesBuilder strings.Builder

	for _, value := range global.Config().Proxies {
		if value.Server == "" || value.Type == "" || (value.Root && h.user.Quota != -1) {
			continue
		}

		proxyConfig, name := h.getClashConfig(&value)
		if name == "" || proxyConfig == nil {
			continue
		}

		b, err := json.Marshal(proxyConfig)
		if err != nil || len(b) == 0 {
			continue
		}

		proxiesBuilder.WriteString("\n  - ")
		proxiesBuilder.Write(b)
		proxiesName = append(proxiesName, name)
	}

	if len(proxiesName) < 1 {
		return ""
	}

	bn, err := json.Marshal(proxiesName)
	if err != nil {
		return ""
	}

	tpl, err := os.ReadFile(global.Config().ClashPath)
	if err != nil || len(tpl) < 1 {
		return ""
	}

	replacer := strings.NewReplacer(
		`[proxies]`, proxiesBuilder.String(),
		`[proxies_name]`, string(bn),
	)

	return replacer.Replace(string(tpl))
}

// buildV2ray 处理 v2rayN 订阅
func (h *handler) buildV2ray() string {
	if !h.checkUser() {
		return utils.Base64Encode("vless://0@cn.bing.com:80?type=tcp#!!! 订阅已过期 !!!")
	}

	if len(global.Config().Proxies) < 1 {
		return ""
	}

	var subscription strings.Builder
	for _, value := range global.Config().Proxies {
		if value.Server == "" || value.Type == "" || (value.Root && h.user.Quota != -1) {
			continue
		}
		if link := h.getV2rayConfig(&value); link != "" {
			subscription.WriteString(link)
			subscription.WriteString("\n")
		}
	}

	return utils.Base64Encode(subscription.String())
}

// getClashConfig 生成 Clash 配置对象
func (h *handler) getClashConfig(value *config.Proxies) (any, string) {
	p := *value
	p.SetProxyDefault()

	if p.Type != "vless" {
		return nil, ""
	}

	switch {
	case p.Flow == "xtls-rprx-vision" && p.RealityOpts.PublicKey != "":
		p.XhttpOpts = config.XhttpOpts{}
		return &p, p.Name
	case p.Network == "xhttp":
		p.RealityOpts = config.RealityOpts{}
		return &p, p.Name
	case p.Network == "tcp":
		p.RealityOpts = config.RealityOpts{}
		p.XhttpOpts = config.XhttpOpts{}
		return &p, p.Name
	default:
		return nil, ""
	}
}

// getV2rayConfig 生成 v2rayN 链接
func (h *handler) getV2rayConfig(value *config.Proxies) string {
	p := *value
	p.SetProxyDefault()

	if p.Type == "" || p.Server == "" || p.Port == 0 {
		return ""
	}

	var link strings.Builder
	link.WriteString(p.Type)
	link.WriteString("://")
	link.WriteString(p.Uuid)
	link.WriteString("@")
	link.WriteString(net.JoinHostPort(strings.Trim(p.Server, "[]"), fmt.Sprint(p.Port)))
	link.WriteString("?encryption=none")

	if p.RealityOpts.PublicKey != "" {
		link.WriteString("&security=reality")
	} else if p.Tls {
		link.WriteString("&security=tls")
	}

	if p.Servername != "" {
		link.WriteString("&sni=")
		link.WriteString(p.Servername)
	}

	if p.ClientFingerprint != "" {
		link.WriteString("&fp=")
		link.WriteString(p.ClientFingerprint)
	}

	if p.Flow != "" {
		link.WriteString("&flow=")
		link.WriteString(p.Flow)
	}

	if p.Network != "" {
		link.WriteString("&type=")
		link.WriteString(p.Network)
	}

	if len(p.Alpn) > 0 {
		link.WriteString("&alpn=")
		link.WriteString(url.QueryEscape(strings.Join(p.Alpn, ",")))
	}

	switch p.Network {
	case "tcp":
		link.WriteString("&headerType=none")
		if p.RealityOpts.PublicKey != "" {
			link.WriteString("&pbk=")
			link.WriteString(p.RealityOpts.PublicKey)
			link.WriteString("&sid=")
			link.WriteString(p.RealityOpts.ShortId)
		}
	case "xhttp":
		if p.XhttpOpts.Path != "" {
			link.WriteString("&path=")
			link.WriteString(url.QueryEscape(p.XhttpOpts.Path))
		}
		if p.XhttpOpts.Mode != "" {
			link.WriteString("&mode=")
			link.WriteString(url.QueryEscape(p.XhttpOpts.Mode))
		}
		if p.Servername != "" {
			link.WriteString("&host=")
			link.WriteString(url.QueryEscape(p.Servername))
		}

		extraMap := make(map[string]any, len(p.XhttpOpts.Extra)+1)
		if len(p.XhttpOpts.Extra) > 0 {
			maps.Copy(extraMap, p.XhttpOpts.Extra)
		}

		if p.XhttpOpts.DownloadSettings != nil {
			ds := p.XhttpOpts.DownloadSettings
			v2rayDs := make(map[string]any)
			if ds.Server != "" {
				v2rayDs["address"] = ds.Server
			}
			v2rayDs["port"] = cmp.Or(ds.Port, 443)
			v2rayDs["network"] = "xhttp"

			xhttpSet := make(map[string]any)
			if ds.Path != "" {
				xhttpSet["path"] = ds.Path
			}
			if ds.Mode != "" {
				xhttpSet["mode"] = ds.Mode
			}
			if len(xhttpSet) > 0 {
				v2rayDs["xhttpSettings"] = xhttpSet
			}

			if ds.RealityOpts.PublicKey != "" {
				v2rayDs["security"] = "reality"
				realitySet := map[string]any{
					"publicKey": ds.RealityOpts.PublicKey,
					"shortId":   ds.RealityOpts.ShortId,
				}
				if ds.Servername != "" {
					realitySet["serverName"] = ds.Servername
				}
				if ds.ClientFingerprint != "" {
					realitySet["fingerprint"] = ds.ClientFingerprint
				}
				v2rayDs["realitySettings"] = realitySet
			} else {
				v2rayDs["security"] = "tls"
				tlsSet := make(map[string]any)
				if ds.Servername != "" {
					tlsSet["serverName"] = ds.Servername
				}
				if ds.ClientFingerprint != "" {
					tlsSet["fingerprint"] = ds.ClientFingerprint
				}
				if len(tlsSet) > 0 {
					v2rayDs["tlsSettings"] = tlsSet
				}
			}

			if len(v2rayDs) > 0 {
				extraMap["downloadSettings"] = v2rayDs
			}
		}

		if len(extraMap) > 0 {
			extraBytes, err := json.Marshal(extraMap)
			if err == nil {
				link.WriteString("&extra=")
				link.WriteString(url.QueryEscape(string(extraBytes)))
			}
		}
	}

	link.WriteString("#")
	link.WriteString(url.QueryEscape(p.Name))
	return link.String()
}

// checkUser 检测是否过期
func (h *handler) checkUser() bool {
	if h.user.ExpiryDate == "" || h.user.ExpiryDate == "0" {
		return true
	}
	t, err := time.ParseInLocation(time.DateOnly, h.user.ExpiryDate, time.UTC)
	return err != nil || t.AddDate(0, 0, 1).After(time.Now().In(time.UTC))
}
