package service

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/utils"

	"github.com/gin-gonic/gin"
)

func ClashHandle(ctx *gin.Context, f []byte) string {
	params := regexp.MustCompile(`^/(.*)\.html$`).FindStringSubmatch(ctx.Request.URL.Path)
	userName := params[1]
	subscription := ""
	if len(global.Config.Env.Trojan) > 0 {
		for _, value := range global.Config.Env.Trojan {
			if value.Domain == "" {
				continue
			}
			ws := ""
			if value.Port == "443" {
				//使用cdn
				ws = ", network: ws, ws-opts: {path: " + value.WsPath + ", headers: {Host: " + value.Domain + "}}"
			}
			replacer := strings.NewReplacer(`%username%`, userName, `%domain%`, value.Domain, `%port%`, value.Port, `%ws%`, ws)
			subscription += replacer.Replace(string(f))

			return subscription //目前只支持单节点vpn
		}

	}
	return subscription

}

func TrojanGoHandle(ctx *gin.Context) string {
	params := regexp.MustCompile(`^/(.*)\.html$`).FindStringSubmatch(ctx.Request.URL.Path)
	userName := params[1]
	subscription := ""
	if len(global.Config.Env.Trojan) > 0 {
		for _, value := range global.Config.Env.Trojan {
			if value.Domain == "" {
				continue
			}
			if value.Port == "443" {
				//使用cdn
				subscription += fmt.Sprintf("trojan://%s@%s:443?security=tls&headerType=none&fp=chrome&uTLS=chrome&mux=1&type=ws&path=%s&host=%s&sni=%s#外网信息复杂_理智分辨真假\n", userName, value.Domain, value.WsPath, value.Domain, value.Domain)
			} else {
				//直连
				subscription += fmt.Sprintf("trojan://%s@%s:%s?security=tls&headerType=none&fp=chrome&uTLS=chrome&mux=1&alpn=h2,http/1.1&type=tcp&path=%s&host=%s&sni=%s#外网信息复杂_理智分辨真假\n", userName, value.Domain, value.Port, value.WsPath, value.Domain, value.Domain)
			}
		}
	}

	if len(global.Config.Env.SuperUrl) > 0 {
		subscription += strings.Join(global.Config.Env.SuperUrl, "\n")
	}

	return utils.Base64Encode(subscription)
}
