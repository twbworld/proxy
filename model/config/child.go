package config

import "fmt"

type RealityOpts struct {
	PublicKey string `json:"public-key" mapstructure:"public-key" yaml:"public-key"`
	ShortId   string `json:"short-id" mapstructure:"short-id" yaml:"short-id"`
}
type WsOpts struct {
	Path    string  `json:"path" mapstructure:"path" yaml:"path"`
	Headers Headers `json:"headers" mapstructure:"headers" yaml:"headers"`
}
type Headers struct {
	Host string `json:"host" mapstructure:"host" yaml:"host"`
}
type Proxy struct {
	Name              string      `json:"name" mapstructure:"name" yaml:"name"`
	Type              string      `json:"type" mapstructure:"type" yaml:"type"`
	Server            string      `json:"server" mapstructure:"server" yaml:"server"`
	Port              string      `json:"port" mapstructure:"port" yaml:"port"`
	Tls               bool        `json:"tls" mapstructure:"tls" yaml:"tls"`
	Udp               bool        `json:"udp" mapstructure:"udp" yaml:"udp"`
	SkipCertVerify    bool        `json:"skip-cert-verify" mapstructure:"skip-cert-verify" yaml:"skip-cert-verify"`
	ClientFingerprint string      `json:"client-fingerprint" mapstructure:"client-fingerprint" yaml:"client-fingerprint"`
	Alpn              []string    `json:"alpn" mapstructure:"alpn" yaml:"alpn"`
	Sni               string      `json:"sni" mapstructure:"sni" yaml:"sni"`
	Uuid              string      `json:"uuid" mapstructure:"uuid" yaml:"uuid"`
	Flow              string      `json:"flow" mapstructure:"flow" yaml:"flow"`
	Network           string      `json:"network" mapstructure:"network" yaml:"network"`
	RealityOpts       RealityOpts `json:"reality-opts" mapstructure:"reality-opts" yaml:"reality-opts"`
	WsOpts            WsOpts      `json:"ws-opts" mapstructure:"ws-opts" yaml:"ws-opts"`
	Root              bool        `json:"root,omitempty" mapstructure:"root" yaml:"root"`
}

type Subscribe struct {
	Filename       string `json:"filename" mapstructure:"filename" yaml:"filename"`
	UpdateInterval uint16 `json:"update_interval" mapstructure:"update_interval" yaml:"update_interval"`
	PageUrl        string `json:"page_url" mapstructure:"page_url" yaml:"page_url"`
}

type Database struct {
	Type          string `json:"type" mapstructure:"type" yaml:"type"`
	SqlitePath    string `json:"sqlite_path" mapstructure:"sqlite_path" yaml:"sqlite_path"`
	MysqlHost     string `json:"mysql_host" mapstructure:"mysql_host" yaml:"mysql_host"`
	MysqlPort     string `json:"mysql_port" mapstructure:"mysql_port" yaml:"mysql_port"`
	MysqlDbname   string `json:"mysql_dbname" mapstructure:"mysql_dbname" yaml:"mysql_dbname"`
	MysqlUsername string `json:"mysql_username" mapstructure:"mysql_username" yaml:"mysql_username"`
	MysqlPassword string `json:"mysql_password" mapstructure:"mysql_password" yaml:"mysql_password"`
}

type Telegram struct {
	Token string `json:"token" mapstructure:"token" yaml:"token"`
	Id    int64  `json:"id" mapstructure:"id" yaml:"id"`
}

func (p *Proxy) SetProxyDefault() {
	domain := p.WsOpts.Headers.Host
	if domain == "" {
		//套cdn(如使用优选ip),则host/sni不等于server
		//PS: 这可判断Server是否为域名
		domain = p.Server
		p.WsOpts.Headers.Host = domain
	}
	if p.Sni == "" && domain != "" {
		p.Sni = domain
	}
	if p.Name == "" {
		p.Name = fmt.Sprintf("外网信息复杂_理智分辨真假_%s_%s", p.Server, p.Port)
	}
	if p.ClientFingerprint == "" {
		p.ClientFingerprint = "chrome"
	}
	if len(p.Alpn) == 0 {
		p.Alpn = []string{"h2", "http/1.1"}
	}
	p.Tls = true
	p.Udp = true
	// p.SkipCertVerify = false
}
