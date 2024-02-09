package config

type AppConfig struct {
	ProjectName        string `default:"VPN会员系统" json:"projectName" info:"本项目名称"`
	GinAddr            string `default:":80" json:"ginAddr" info:"gin监听的地址"`
	EnvPath            string `default:"config/.env" json:"envPath" info:"敏感配置文件的路径"`
	ClashPath          string `default:"config/clash.ini" json:"clashPath" info:"clash默认配置文件"`
	GinLogPath         string `default:"log/gin.log" json:"ginLogPath" info:"gin日志文件"`
	RunLogPath         string `default:"log/run.log" json:"runLogPath" info:"运行日志文件"`
}
