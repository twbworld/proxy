package config

type Trojan struct{
	Domain string `json:"domain" mapstructure:"domain" info:"域名"`
	Port string `json:"port" mapstructure:"port" info:"端口"`
	WsPath string `json:"wsPath" mapstructure:"wsPath" info:"WebSocket路径"`
}

type Telegram struct{
	Token string `json:"token" mapstructure:"token" info:"BotFather新建的聊天室"`
	Id int64 `json:"id" mapstructure:"id" info:"userinfobot获取"`
}

type Env struct{
	Debug bool `json:"debug" mapstructure:"debug" info:"项目环境"`
	SuperUrl []string `json:"superUrl" mapstructure:"superUrl" info:"除trojan外的连接"`
	Trojan []Trojan `json:"trojan" mapstructure:"trojan" info:"可配置多个trojan地址"`
	Telegram Telegram `json:"telegram" mapstructure:"telegram" info:"Telegram聊天室配置"`
}
