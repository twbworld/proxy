package config


type AppConfig struct{
	ProjectName string `json:"projectName" mapstructure:"projectName" info:"本项目名称"`
	GinAddr string `json:"ginAddr" mapstructure:"ginAddr" info:"gin监听的地址"`
	EnvPath string `json:"envPath" mapstructure:"envPath" info:"敏感配置文件的路径"`
	ClashPath string `json:"clashPath" mapstructure:"clashPath" info:"clash默认配置文件"`
	UsersPath string `json:"usersPath" mapstructure:"usersPath" info:"用户配置"`
	GinLogPath string `json:"ginLogPath" mapstructure:"ginLogPath" info:"gin日志文件"`
	RunLogPath string `json:"runLogPath" mapstructure:"runLogPath" info:"运行日志文件"`
}
