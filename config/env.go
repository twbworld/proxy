package config

type Trojan struct{
	Domain string `json:"domain" mapstructure:"domain" info:"域名"`
	Port string `json:"port" mapstructure:"port" info:"端口"`
	WsPath string `json:"wsPath" mapstructure:"wsPath" info:"WebSocket路径"`
}

type Mysql struct{
	Dbname string `json:"dbname" mapstructure:"dbname" info:"数据库名"`
	Host string `json:"host" mapstructure:"host" info:"地址"`
	Username string `json:"username" mapstructure:"username" info:"用户名"`
	Password string `json:"password" mapstructure:"password" info:"密码"`
}

type Telegram struct{
	Token string `json:"token" mapstructure:"token" info:"BotFather新建的聊天室"`
	Id int64 `json:"id" mapstructure:"id" info:"userinfobot获取"`
}

type Env struct{
	Debug bool `json:"debug" mapstructure:"debug" info:"项目环境"`
	Mysql Mysql `json:"mysql" mapstructure:"mysql" info:"数据库配置"`
	SuperUrl []string `json:"superUrl" mapstructure:"superUrl" info:"除trojan外的连接"`
	Trojan []Trojan `json:"trojan" mapstructure:"trojan" info:"可配置多个trojan地址"`
	Telegram Telegram `json:"tg" mapstructure:"tg" info:"Telegram聊天室配置"`
}
