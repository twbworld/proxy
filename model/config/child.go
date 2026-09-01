package config

import (
	"cmp"
	"fmt"
	"net"
	"strings"
)

type RealityOpts struct {
	PublicKey string `mapstructure:"public-key" json:"public-key" yaml:"public-key"`
	ShortId   string `mapstructure:"short-id" json:"short-id" yaml:"short-id"`
}
type DownloadSettings struct {
	Server            string      `mapstructure:"server" json:"server,omitzero" yaml:"server"`
	Port              uint        `mapstructure:"port" json:"port,omitzero" yaml:"port"`
	Servername        string      `mapstructure:"servername" json:"servername,omitzero" yaml:"servername"`
	ClientFingerprint string      `mapstructure:"client-fingerprint" json:"client-fingerprint,omitzero" yaml:"client-fingerprint"`
	RealityOpts       RealityOpts `mapstructure:"reality-opts" json:"reality-opts,omitzero" yaml:"reality-opts"`
	Path              string      `mapstructure:"path" json:"path,omitzero" yaml:"path"`
	Mode              string      `mapstructure:"mode" json:"mode,omitzero" yaml:"mode"`
}
type XhttpOpts struct {
	Mode             string            `mapstructure:"mode" json:"mode,omitzero" yaml:"mode"`
	Path             string            `mapstructure:"path" json:"path,omitzero" yaml:"path"`
	DownloadSettings *DownloadSettings `mapstructure:"download-settings" json:"download-settings,omitzero" yaml:"download-settings"`
	Extra            map[string]any    `mapstructure:"extra" json:"extra,omitempty" yaml:"extra"`
}
type Proxies struct {
	Name              string      `mapstructure:"name" json:"name" yaml:"name"`
	Type              string      `mapstructure:"type" json:"type" yaml:"type"`
	Server            string      `mapstructure:"server" json:"server" yaml:"server"`
	Port              uint        `mapstructure:"port" json:"port,omitzero" yaml:"port"`
	Tls               bool        `mapstructure:"tls" json:"tls" yaml:"tls"`
	Udp               bool        `mapstructure:"udp" json:"udp" yaml:"udp"`
	SkipCertVerify    bool        `mapstructure:"skip-cert-verify" json:"skip-cert-verify" yaml:"skip-cert-verify"`
	ClientFingerprint string      `mapstructure:"client-fingerprint" json:"client-fingerprint" yaml:"client-fingerprint"`
	Alpn              []string    `mapstructure:"alpn" json:"alpn" yaml:"alpn"`
	Servername        string      `mapstructure:"servername" json:"servername" yaml:"servername"`
	Uuid              string      `mapstructure:"uuid" json:"uuid" yaml:"uuid"`
	Flow              string      `mapstructure:"flow" json:"flow,omitzero" yaml:"flow"`
	Network           string      `mapstructure:"network" json:"network" yaml:"network"`
	RealityOpts       RealityOpts `mapstructure:"reality-opts" json:"reality-opts,omitzero" yaml:"reality-opts"`
	XhttpOpts         XhttpOpts   `mapstructure:"xhttp-opts" json:"xhttp-opts,omitzero" yaml:"xhttp-opts"`
	Root              bool        `mapstructure:"root" json:"-" yaml:"root"`
}

type Subscribe struct {
	Filename       string `mapstructure:"filename" json:"filename" yaml:"filename"`
	UpdateInterval uint16 `mapstructure:"update_interval" json:"update_interval" yaml:"update_interval"`
	PageUrl        string `mapstructure:"page_url" json:"page_url" yaml:"page_url"`
}

type Database struct {
	Type          string `mapstructure:"type" json:"type" yaml:"type"`
	SqlitePath    string `mapstructure:"sqlite_path" json:"sqlite_path" yaml:"sqlite_path"`
	MysqlHost     string `mapstructure:"mysql_host" json:"mysql_host" yaml:"mysql_host"`
	MysqlPort     string `mapstructure:"mysql_port" json:"mysql_port" yaml:"mysql_port"`
	MysqlDbname   string `mapstructure:"mysql_dbname" json:"mysql_dbname" yaml:"mysql_dbname"`
	MysqlUsername string `mapstructure:"mysql_username" json:"mysql_username" yaml:"mysql_username"`
	MysqlPassword string `mapstructure:"mysql_password" json:"mysql_password" yaml:"mysql_password"`
}

type Redis struct {
	Addr                   string `mapstructure:"addr" json:"addr" yaml:"addr"`
	Password               string `mapstructure:"password" json:"password" yaml:"password"`
	DB                     int64  `mapstructure:"db" json:"db" yaml:"db"`
	LockExpiry             int64  `mapstructure:"lock_expiry" json:"lock_expiry" yaml:"lock_expiry"`
	ConversationHistoryTTL int64  `mapstructure:"conversation_history_ttl" json:"conversation_history_ttl" yaml:"conversation_history_ttl"`
	HistoryLockExpiry      int64  `mapstructure:"history_lock_expiry" json:"history_lock_expiry" yaml:"history_lock_expiry"`
}

type Telegram struct {
	Token string `mapstructure:"token" json:"token" yaml:"token"`
	Id    int64  `mapstructure:"id" json:"id" yaml:"id"`
}

func (p *Proxies) SetProxyDefault() {
	// 如果 Servername 仍为空，且 Server 看起来像域名（排查非 IPv4/IPv6 地址），则兜底
	if p.Servername == "" && p.Server != "" && net.ParseIP(strings.Trim(p.Server, "[]")) == nil {
		p.Servername = p.Server
	}

	if p.Name == "" {
		var name strings.Builder
		if !p.Root {
			name.WriteString("外网信息复杂_理智分辨真假_")
		}
		name.WriteString(fmt.Sprintf("%s_%s", p.Network, p.Server))
		if p.RealityOpts.PublicKey != "" {
			name.WriteString("_reality")
		}
		if p.Flow != "" {
			name.WriteString(fmt.Sprintf("_%s", p.Flow))
		}
		if p.XhttpOpts.DownloadSettings != nil && p.XhttpOpts.DownloadSettings.Server != "" {
			name.WriteString("_downSet")
		}
		p.Name = name.String()
	}
	p.ClientFingerprint = cmp.Or(p.ClientFingerprint, "chrome")
	if len(p.Alpn) == 0 {
		// 默认 ALPN，针对不同协议可能需要调整，这里保持默认
		p.Alpn = []string{"h2", "http/1.1"}
	}

	// 强制开启 TLS/UDP
	p.Tls = true
	p.Udp = true
}
