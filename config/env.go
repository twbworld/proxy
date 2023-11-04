package config

type RealityOpts struct {
	PublicKey string `json:"public-key" mapstructure:"public-key" info:"公钥"`
	ShortId   string `json:"short-id" mapstructure:"short-id" info:"客户端短ID"`
}
type WsOpts struct {
	Path    string  `json:"path" mapstructure:"path" info:"路径"`
	Headers Headers `json:"headers" mapstructure:"headers" info:"头配置"`
}
type Headers struct {
	Host string `json:"host" mapstructure:"host" info:"地址"`
}
type Proxy struct {
	Name              string      `json:"name" mapstructure:"name" info:"唯一名称"`
	Type              string      `default:"vless" json:"type" mapstructure:"type" info:"proxy类型(如vless/vmess/trojan)"`
	Server            string      `json:"server" mapstructure:"server" info:"地址(域名或ip)"`
	Port              string      `default:"443" json:"port" mapstructure:"port" info:"端口"`
	Tls               bool        `json:"tls" mapstructure:"tls" info:"开启tls"`
	Udp               bool        `json:"udp" mapstructure:"udp" info:"开启udp"`
	SkipCertVerify    bool        `json:"skip-cert-verify" mapstructure:"skip-cert-verify" info:"跳过证书"`
	ClientFingerprint string      `default:"chrome" json:"client-fingerprint" mapstructure:"client-fingerprint" info:"utls指纹(如chrome/safari)"`
	Alpn              []string    `json:"alpn" mapstructure:"alpn" info:"alpn"`
	Sni               string      `json:"sni" mapstructure:"sni" info:"sni域名"`
	Uuid              string      `json:"uuid" mapstructure:"uuid" info:"用户ID"`
	Flow              string      `json:"flow" mapstructure:"flow" info:"流控类型"`
	Network           string      `default:"tcp" json:"network" mapstructure:"network" info:"传输协议(如tcp/ws/grpc)"`
	RealityOpts       RealityOpts `json:"reality-opts" mapstructure:"reality-opts" info:"Reality协议配置"`
	WsOpts            WsOpts      `json:"ws-opts" mapstructure:"ws-opts" info:"WebSocket协议配置"`
	Root              bool        `json:"root,omitempty" mapstructure:"root" info:"是否管理员(quota=-1)用户使用"`
}

type Telegram struct {
	Token string `json:"token" mapstructure:"token" info:"BotFather新建的聊天室"`
	Id    int64  `json:"id" mapstructure:"id" info:"userinfobot获取"`
}

type Env struct {
	Debug    bool     `json:"debug" mapstructure:"debug" info:"项目环境"`
	Domain   string   `json:"domain" mapstructure:"domain" info:"项目域名,端口默认80, webhook需要"`
	Proxy    []Proxy  `json:"proxy" mapstructure:"proxy" info:"代理配置"`
	Telegram Telegram `json:"telegram" mapstructure:"telegram" info:"Telegram聊天室配置"`
}
