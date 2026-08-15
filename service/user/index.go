package user

import (
	"cmp"
	"encoding/json"
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

type BaseService struct{}

type v2ray struct{}
type clash struct{}
type class interface {
	Handle(user *db.Users) string
}

func (b *BaseService) SetProtocol(t string) class {
	switch t {
	case "clash":
		return &clash{}
	default:
		return &v2ray{}
	}
}

// Handle 处理 Clash/Mihomo 订阅
func (c *clash) Handle(user *db.Users) string {
	if !checkUser(user) {
		return `proxies:
  - {name: "!!! 订阅已过期 !!!", type: trojan, server: cn.bing.com, port: 80, password: 0, network: tcp}
proxy-groups:
  - {name: "!!!!!! 订阅已过期 !!!!!!", type: select, proxies: ["!!! 订阅已过期 !!!"]}`
	}

	if len(global.Config.Proxies) < 1 || !utils.FileExist(global.Config.ClashPath) {
		return ""
	}

	proxiesName := make([]string, 0, len(global.Config.Proxies))
	var proxiesBuilder strings.Builder

	for _, value := range global.Config.Proxies {
		if value.Server == "" || value.Type == "" || (value.Root && user.Quota != -1) {
			continue
		}

		proxyConfig, name := c.getConfig(&value)
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

	tpl, err := os.ReadFile(global.Config.ClashPath)
	if err != nil || len(tpl) < 1 {
		return ""
	}

	replacer := strings.NewReplacer(
		`[proxies]`, proxiesBuilder.String(),
		`[proxies_name]`, string(bn),
	)

	return replacer.Replace(string(tpl))
}

// Handle 处理 v2rayN 订阅
func (x *v2ray) Handle(user *db.Users) string {
	if !checkUser(user) {
		return utils.Base64Encode("vless://0@cn.bing.com:80?type=tcp#!!! 订阅已过期 !!!")
	}

	if len(global.Config.Proxies) < 1 {
		return ""
	}

	var subscription strings.Builder
	for _, value := range global.Config.Proxies {
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

// getConfig 生成 Clash 配置对象
func (c *clash) getConfig(value *config.Proxies) (any, string) {
	p := *value
	p.SetProxyDefault()

	// 允许 vless 下所有类型,包括支持最新的 xhttp
	if p.Type != "vless" {
		return nil, ""
	}

	// 根据 Network 和 Flow 判断具体类型
	switch {
	case p.Flow == "xtls-rprx-vision" && p.RealityOpts.PublicKey != "":
		// VLESS Reality (TCP)
		p.GrpcOpts = config.GrpcOpts{}
		p.XhttpOpts = config.XhttpOpts{}
		return &p, p.Name
	case p.Network == "grpc":
		// VLESS gRPC
		p.RealityOpts = config.RealityOpts{}
		p.XhttpOpts = config.XhttpOpts{}
		return &p, p.Name
	case p.Network == "xhttp":
		// VLESS XHTTP
		p.RealityOpts = config.RealityOpts{}
		p.GrpcOpts = config.GrpcOpts{}
		return &p, p.Name
	case p.Network == "tcp":
		// VLESS TCP (TLS/Vision)
		p.RealityOpts = config.RealityOpts{}
		p.GrpcOpts = config.GrpcOpts{}
		p.XhttpOpts = config.XhttpOpts{}
		return &p, p.Name
	default:
		return nil, ""
	}
}

// getConfig 生成 v2rayN 链接
func (x *v2ray) getConfig(value *config.Proxies) string {
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
	// JoinHostPort判断IPv6会加上[]
	link.WriteString(net.JoinHostPort(strings.Trim(p.Server, "[]"), fmt.Sprint(p.Port)))

	// 公共参数
	link.WriteString("?encryption=none")

	// Security
	if p.RealityOpts.PublicKey != "" {
		link.WriteString("&security=reality")
	} else if p.Tls {
		link.WriteString("&security=tls")
	}

	// Servername / SNI
	if p.Servername != "" {
		link.WriteString("&sni=")
		link.WriteString(p.Servername)
	}

	// Fingerprint
	if p.ClientFingerprint != "" {
		link.WriteString("&fp=")
		link.WriteString(p.ClientFingerprint)
	}

	// Flow (Vision)
	if p.Flow != "" {
		link.WriteString("&flow=")
		link.WriteString(p.Flow)
	}

	// Network Type (只有 tcp, grpc, xhttp)
	if p.Network != "" {
		link.WriteString("&type=")
		link.WriteString(p.Network)
	}

	// ALPN
	if len(p.Alpn) > 0 {
		link.WriteString("&alpn=")
		link.WriteString(url.QueryEscape(strings.Join(p.Alpn, ",")))
	}

	// 协议特定参数
	switch p.Network {
	case "tcp":
		link.WriteString("&headerType=none")
		if p.RealityOpts.PublicKey != "" {
			link.WriteString("&pbk=")
			link.WriteString(p.RealityOpts.PublicKey)
			link.WriteString("&sid=")
			link.WriteString(p.RealityOpts.ShortId)
		}
	case "grpc":
		if p.GrpcOpts.GrpcServiceName != "" {
			link.WriteString("&serviceName=")
			link.WriteString(p.GrpcOpts.GrpcServiceName)
		}
		link.WriteString("&mode=gun")
		// gRPC 通常需要 authority
		if p.Servername != "" {
			link.WriteString("&authority=")
			link.WriteString(p.Servername)
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

		// 汇总整合 Extra 和 DownloadSettings 给 v2ray 使用 (映射为规范驼峰)
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

			// 构建 xhttpSettings 层级
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

			// 构建 realitySettings 或 tlsSettings 层级
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

// 检测过期
func checkUser(user *db.Users) bool {
	if user.ExpiryDate == nil || *user.ExpiryDate == "" || *user.ExpiryDate == "0" {
		return true
	}

	t, err := time.ParseInLocation(time.DateOnly, *user.ExpiryDate, time.UTC)

	return err != nil || t.AddDate(0, 0, 1).After(time.Now().In(time.UTC))
}
