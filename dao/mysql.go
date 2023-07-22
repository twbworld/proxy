package dao

import (
	"fmt"
	"github.com/twbworld/proxy/global"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var DB *sqlx.DB

func InitMysql() (err error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", global.Config.TrojanGoConfig.Mysql.Username, global.Config.TrojanGoConfig.Mysql.Password, global.Config.TrojanGoConfig.Mysql.Host, global.Config.TrojanGoConfig.Mysql.Port, global.Config.TrojanGoConfig.Mysql.Dbname)

	global.Log.Info("连接服务: ", dsn)

	//也可以使用MustConnect连接不成功就panic
	DB, err = sqlx.Connect("mysql", dsn)
	if err != nil {
		global.Log.Fatal("数据库连接失败 ", err)
	}

	DB.SetMaxOpenConns(20)
	DB.SetMaxIdleConns(10)

	return
}
