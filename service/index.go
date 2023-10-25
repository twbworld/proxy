package service

import (
	"fmt"
	"os"
	"strings"

	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/model"
	"github.com/twbworld/proxy/utils"
)

func ClashHandle(user *model.Users) string {
	subscription := ""
	if len(global.Config.Env.Trojan) < 1 {
		return subscription
	}

	f, err := os.ReadFile(global.Config.AppConfig.ClashPath)
	if err != nil || len(f) < 1 {
		return subscription
	}

	tmp := `{name: %s, server: %s, port: %s, type: %s, password: %s, sni: %s, udp: true, tls: true, skip-cert-verify: false, network: tcp, alpn: ["h2", "http/1.1"], client-fingerprint: chrome}`
	tmpCdn := `{name: %s, server: %s, port: %s, type: %s, password: %s, sni: %s, udp: true, tls: true, skip-cert-verify: false, network: ws, ws-opts: {path: %s, headers: {Host: %s}}}`
	proxies := ""
	proxiesName := ""

	for _, value := range global.Config.Env.Trojan {
		if value.Domain == "" || (value.Root && user.Quota != -1) {
			continue
		}
		name := "外网信息复杂_理智分辨真假" + "_" + value.Domain + "_" + value.Port
		proxiesName += name + ","

		if value.Port == "443" && !value.Root {
			//使用cdn
			proxies += fmt.Sprintf(tmpCdn, name, value.Domain, value.Port, "trojan", user.Username, value.Domain, value.WsPath, value.Domain) + ","
			continue
		}
		proxies += fmt.Sprintf(tmp, name, value.Domain, value.Port, "trojan", user.Username, value.Domain) + ","
	}

	replacer := strings.NewReplacer(`%proxies%`, "["+strings.Trim(proxies, ",")+"]", `%proxies_name%`, "["+strings.Trim(proxiesName, ",")+"]")
	subscription += replacer.Replace(string(f))

	return subscription

}

func TrojanGoHandle(user *model.Users) string {
	subscription := ""
	if len(global.Config.Env.Trojan) < 1 {
		return subscription
	}
	if len(global.Config.Env.SuperUrl) > 0 {
		subscription += strings.Join(global.Config.Env.SuperUrl, "\n") + "\n"
	}

	tmp := `trojan://%s@%s:%s?security=tls&headerType=none&fp=chrome&uTLS=chrome&mux=1&sni=%s&type=tcp&alpn=h2,http/1.1#外网信息复杂_理智分辨真假`
	tmpCdn := `trojan://%s@%s:%s?security=tls&headerType=none&fp=chrome&uTLS=chrome&mux=1&sni=%s&type=ws&path=%s&host=%s#外网信息复杂_理智分辨真假`

	for _, value := range global.Config.Env.Trojan {
		if value.Domain == "" || (value.Root && user.Quota != -1) {
			continue
		}
		if value.Port == "443" && !value.Root {
			//使用cdn
			subscription += fmt.Sprintf(tmpCdn, user.Username, value.Domain, value.Port, value.Domain, value.WsPath, value.Domain) + "\n"
			continue
		}
		subscription += fmt.Sprintf(tmp, user.Username, value.Domain, value.Port, value.Domain) + "\n"
	}

	return utils.Base64Encode(subscription)
}
