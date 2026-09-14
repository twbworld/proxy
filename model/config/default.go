package config

import "encoding/json/v2"

// viper 要用到 mapstructure/yaml
type Config struct {
	Debug            bool      `mapstructure:"debug" json:"debug" yaml:"debug"`
	ProjectName      string    `mapstructure:"project_name" json:"project_name" yaml:"project_name"`
	GinAddr          string    `mapstructure:"gin_addr" json:"gin_addr" yaml:"gin_addr"`
	Domain           string    `mapstructure:"domain" json:"domain" yaml:"domain"`
	ClashPath        string    `mapstructure:"clash_path" json:"clash_path" yaml:"clash_path"`
	StaticDir        string    `mapstructure:"static_dir" json:"static_dir" yaml:"static_dir"`
	GinLogPath       string    `mapstructure:"gin_log_path" json:"gin_log_path" yaml:"gin_log_path"`
	RunLogPath       string    `mapstructure:"run_log_path" json:"run_log_path" yaml:"run_log_path"`
	LogRetentionDays uint      `mapstructure:"log_retention_days" json:"log_retention_days" yaml:"log_retention_days"`
	Tz               string    `mapstructure:"tz" json:"tz" yaml:"tz"`
	Cors             []string  `mapstructure:"cors" json:"cors" yaml:"cors"`
	Subscribe        Subscribe `mapstructure:"subscribe" json:"subscribe" yaml:"subscribe"`
	Database         Database  `mapstructure:"database" json:"database" yaml:"database"`
	Redis            Redis     `mapstructure:"redis" json:"redis" yaml:"redis"`
	Telegram         Telegram  `mapstructure:"telegram" json:"telegram" yaml:"telegram"`
	LLM              []LLM     `mapstructure:"llm" json:"llm" yaml:"llm"`
	Proxies          []Proxies `mapstructure:"proxies" json:"proxies" yaml:"proxies"`
}

// DeepCopy 使用 JSON 序列化和反序列化实现 Config 对象的深度拷贝
func (c *Config) DeepCopy() *Config {
	newConfig := new(Config)
	data, _ := json.Marshal(c)
	_ = json.Unmarshal(data, newConfig, json.RejectUnknownMembers(true))
	return newConfig
}
