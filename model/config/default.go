package config

// viper要用到mapstructure/yaml
type Config struct {
	Debug       bool     `json:"debug" mapstructure:"debug" yaml:"debug"`
	ProjectName string   `json:"project_name" mapstructure:"project_name" yaml:"project_name"`
	GinAddr     string   `json:"gin_addr" mapstructure:"gin_addr" yaml:"gin_addr"`
	Domain      string   `json:"domain" mapstructure:"domain" yaml:"domain"`
	ClashPath   string   `json:"clash_path" mapstructure:"clash_path" yaml:"clash_path"`
	StaticDir   string   `json:"static_dir" mapstructure:"static_dir" yaml:"static_dir"`
	GinLogPath  string   `json:"gin_log_path" mapstructure:"gin_log_path" yaml:"gin_log_path"`
	RunLogPath  string   `json:"run_log_path" mapstructure:"run_log_path" yaml:"run_log_path"`
	Tz          string   `json:"tz" mapstructure:"tz" yaml:"tz"`
	Proxy       []Proxy  `json:"proxy" mapstructure:"proxy" yaml:"proxy"`
	Database    Database `json:"database" mapstructure:"database" yaml:"database"`
	Telegram    Telegram `json:"telegram" mapstructure:"telegram" yaml:"telegram"`
	Cors        []string `json:"cors" mapstructure:"cors" yaml:"cors"`
}
