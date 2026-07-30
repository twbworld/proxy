package config

import "fmt"

type RealityOpts struct {
	PublicKey string `json:"public-key" mapstructure:"public-key" yaml:"public-key"`
	ShortId   string `json:"short-id" mapstructure:"short-id" yaml:"short-id"`
}
type GrpcOpts struct {
	GrpcServiceName string `json:"grpc-service-name" mapstructure:"grpc-service-name" yaml:"grpc-service-name"`
}
type DownloadSettings struct {
	Server            string      `json:"server,omitempty" mapstructure:"server" yaml:"server"`
	Port              string      `json:"port,omitempty" mapstructure:"port" yaml:"port"`
	Servername        string      `json:"servername,omitempty" mapstructure:"servername" yaml:"servername"`
	ClientFingerprint string      `json:"client-fingerprint,omitempty" mapstructure:"client-fingerprint" yaml:"client-fingerprint"`
	RealityOpts       RealityOpts `json:"reality-opts,omitempty" mapstructure:"reality-opts" yaml:"reality-opts"`
	Path              string      `json:"path,omitempty" mapstructure:"path" yaml:"path"`
	Mode              string      `json:"mode,omitempty" mapstructure:"mode" yaml:"mode"`
}
type XhttpOpts struct {
	Mode             string                 `json:"mode,omitempty" mapstructure:"mode" yaml:"mode"`
	Path             string                 `json:"path,omitempty" mapstructure:"path" yaml:"path"`
	DownloadSettings *DownloadSettings      `json:"download-settings,omitempty" mapstructure:"download-settings" yaml:"download-settings"`
	Extra            map[string]interface{} `json:"extra,omitempty" mapstructure:"extra" yaml:"extra"`
}
type Proxies struct {
	Name              string      `json:"name" mapstructure:"name" yaml:"name"`
	Type              string      `json:"type" mapstructure:"type" yaml:"type"`
	Server            string      `json:"server" mapstructure:"server" yaml:"server"`
	Port              string      `json:"port" mapstructure:"port" yaml:"port"`
	Tls               bool        `json:"tls" mapstructure:"tls" yaml:"tls"`
	Udp               bool        `json:"udp" mapstructure:"udp" yaml:"udp"`
	SkipCertVerify    bool        `json:"skip-cert-verify" mapstructure:"skip-cert-verify" yaml:"skip-cert-verify"`
	ClientFingerprint string      `json:"client-fingerprint" mapstructure:"client-fingerprint" yaml:"client-fingerprint"`
	Alpn              []string    `json:"alpn" mapstructure:"alpn" yaml:"alpn"`
	Servername        string      `json:"servername" mapstructure:"servername" yaml:"servername"`
	Uuid              string      `json:"uuid" mapstructure:"uuid" yaml:"uuid"`
	Flow              string      `json:"flow" mapstructure:"flow" yaml:"flow"`
	Network           string      `json:"network" mapstructure:"network" yaml:"network"`
	RealityOpts       RealityOpts `json:"reality-opts" mapstructure:"reality-opts" yaml:"reality-opts"`
	GrpcOpts          GrpcOpts    `json:"grpc-opts" mapstructure:"grpc-opts" yaml:"grpc-opts"`
	XhttpOpts         XhttpOpts   `json:"xhttp-opts" mapstructure:"xhttp-opts" yaml:"xhttp-opts"`
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

func (p *Proxies) SetProxyDefault() {
	// 如果 Servername 仍为空，且 Server 看起来像域名（简单判断），则兜底
	if p.Servername == "" && p.Server != "" {
		p.Servername = p.Server
	}

	if p.Name == "" {
		p.Name = fmt.Sprintf("外网信息复杂_理智分辨真假_%s_%s", p.Server, p.Port)
	}
	if p.ClientFingerprint == "" {
		p.ClientFingerprint = "chrome"
	}
	if len(p.Alpn) == 0 {
		// 默认 ALPN，针对不同协议可能需要调整，这里保持默认
		p.Alpn = []string{"h2", "http/1.1"}
	}

	// 强制开启 TLS/UDP
	p.Tls = true
	p.Udp = true
}
