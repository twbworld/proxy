package config

type Mysql struct{
	Host string `default:"127.0.0.1" json:"server_addr" mapstructure:"server_addr" env:"MYSQL_HOST" info:"地址"`
	Port string `default:"3306" json:"server_port" mapstructure:"server_port" env:"MYSQL_PORT" info:"端口"`
	Dbname string `default:"trojan" json:"database" mapstructure:"database" env:"MYSQL_DBNAME" info:"数据库名"`
	Username string `default:"root" json:"username" mapstructure:"username" env:"MYSQL_USERNAME" info:"用户名"`
	Password string `default:"" json:"password" mapstructure:"password" env:"MYSQL_PASSWORD" info:"密码"`
}

type TrojanGoConfig struct{
	Mysql Mysql `json:"mysql" mapstructure:"mysql" info:"数据库配置"`
}
