package dao

import (
	"fmt"
	"log"
	"github.com/twbworld/proxy/global"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var DB *sqlx.DB

func InitMysql() (err error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local", global.Config.Env.Mysql.Username, global.Config.Env.Mysql.Password, global.Config.Env.Mysql.Host, global.Config.Env.Mysql.Dbname)

	//也可以使用MustConnect连接不成功就panic
	DB, err = sqlx.Connect("mysql", dsn)
	if err != nil {
		log.Fatalln("connect DB failed, err: ", err)
	}

	DB.SetMaxOpenConns(20)
	DB.SetMaxIdleConns(10)

	return
}
