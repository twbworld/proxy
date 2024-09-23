package common

import (
	"github.com/twbworld/proxy/model/config"
)

type ClashVlessVisionReality struct {
	*config.Proxy
	WsOpts         int `json:"-"`
	Alpn           int `json:"-"`
	SkipCertVerify int `json:"-"`
}
type ClashVlessVision struct {
	*config.Proxy
	WsOpts      int `json:"-"`
	RealityOpts int `json:"-"`
}
type ClashVlessWs struct {
	*config.Proxy
	RealityOpts int `json:"-"`
}
type ClashTrojanWs struct {
	*config.Proxy
	RealityOpts int    `json:"-"`
	Password    string `json:"password"`
}
type ClashTrojan struct {
	*config.Proxy
	RealityOpts int    `json:"-"`
	WsOpts      int    `json:"-"`
	Uuid        string `json:"-" info:"用户ID或trojan的password"`
	Password    string `json:"password"`
}
