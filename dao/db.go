package dao

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/utils"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sqlx.DB

type db struct{}
type mysql struct{}
type sqlite struct{}
type class interface {
	createTable() error
}

func (d db) setDB(t string) (class, error) {
	switch t {
	case "sqlite":
		s := sqlite{}
		err := s.connect()
		return s, err
	case "mysql":
		m := mysql{}
		err := m.connect()
		return m, err
	default:
		return nil, errors.New("参数错误[okjdoi]: " + t)
	}
}

func (s sqlite) createTable() (err error) {
	var u []string
	err = DB.Select(&u, "SELECT name _id FROM sqlite_master WHERE type ='table'")
	if err != nil {
		return errors.New("错误[giosjfio]: " + err.Error())
	}

	tx, err := DB.Beginx()
	if err != nil {
		return errors.New("开启事务失败[iufghs]: " + err.Error())
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			global.Log.Println("错误[huiosdjghioa]", p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
			if err != nil {
				err = errors.New("错误[fgdjgo]: " + err.Error())
			}
		}
	}()

	if utils.InSlice(&u, "users") < 0 {
		global.Log.Infoln("开始创建users表[soidfjo]")

		_, err = tx.Exec(`CREATE TABLE "users" (
			"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
			"username" text(64) NOT NULL DEFAULT '',
			"password" text(56) NOT NULL DEFAULT '',
			"passwordShow" text(255) NOT NULL DEFAULT '',
			"quota" integer NOT NULL DEFAULT 0,
			"download" integer NOT NULL DEFAULT 0,
			"upload" integer NOT NULL DEFAULT 0,
			"useDays" integer NOT NULL DEFAULT 0,
			"expiryDate" text(10) NOT NULL DEFAULT ''
		  );`)
		if err != nil {
			return errors.New("错误[ugfdhsn]: " + err.Error())
		}
		_, err = tx.Exec(`CREATE INDEX "password" ON "users" ( "password" ASC );`)
		if err != nil {
			return errors.New("错误[jfgfds]: " + err.Error())
		}

		sql := "INSERT INTO `users`(`username`, `password`, `passwordShow`, `quota`) VALUES(:username, :password, :passwordShow, :quota)"
		_, err = tx.NamedExec(sql, map[string]interface{}{
			"username":     "test",
			"password":     "90a3ed9e32b2aaf4c61c410eb925426119e1a9dc53d4286ade99a809",
			"passwordShow": "OTBhM2VkOWUzMmIyYWFmNGM2MWM0MTBlYjkyNTQyNjExOWUxYTlkYzUzZDQyODZhZGU5OWE4MDk=",
			"quota":        "-1",
		})
		if err != nil {
			return errors.New("错误[tyudsfsadm]: " + err.Error())
		}
	}

	if utils.InSlice(&u, "system_info") < 0 {
		global.Log.Infoln("开始创建system_info表[ihjfgmsd]")

		_, err = tx.Exec(`CREATE TABLE "system_info" (
			"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
			"key" text(255) NOT NULL DEFAULT '',
			"value" text(255) NOT NULL DEFAULT '',
			"update_time" text NOT NULL DEFAULT ''
		);`)
		if err != nil {
			return errors.New("错误[jdgfsrsag]: " + err.Error())
		}

		_, err = tx.Exec(`CREATE UNIQUE INDEX "idx_key" ON "system_info" ( "key" ASC );`)
		if err != nil {
			return errors.New("错误[hjdrsrtgf]: " + err.Error())
		}
	}

	return
}

func (m mysql) createTable() (err error) {
	var u []string
	DB.Select(&u, "SHOW TABLES")
	if err != nil {
		return errors.New("错误[hsfds]: " + err.Error())
	}

	tx, err := DB.Beginx()
	if err != nil {
		return errors.New("开启事务失败[khhdsfg]: " + err.Error())
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			global.Log.Println("错误[eoirtujge]", p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
			if err != nil {
				err = errors.New("错误[kjdgsz]: " + err.Error())
			}
		}
	}()

	if utils.InSlice(&u, "users") < 0 {
		global.Log.Infoln("开始创建users表[gdhfd]")

		_, err = tx.Exec("CREATE TABLE `users` ( `id` int unsigned NOT NULL AUTO_INCREMENT, `username` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '用户名', `password` char(56) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '密码', `passwordShow` varchar(255) NOT NULL, `quota` bigint NOT NULL DEFAULT '0' COMMENT '流量限制, 单位byte,1G=1073741824byte;-1:不限', `download` bigint unsigned NOT NULL DEFAULT '0' COMMENT '下行流量', `upload` bigint unsigned NOT NULL DEFAULT '0' COMMENT '上行流量', `useDays` int DEFAULT '0', `expiryDate` char(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT '' COMMENT '限期; 年-月-日', PRIMARY KEY (`id`), KEY `password` (`password`) ) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='用户表';")
		if err != nil {
			return errors.New("错误[kghjffsd]: " + err.Error())
		}

		sql := "INSERT INTO `users`(`username`, `password`, `passwordShow`, `quota`) VALUES(:username, :password, :passwordShow, :quota)"
		_, err = tx.NamedExec(sql, map[string]interface{}{
			"username":     "test",
			"password":     "90a3ed9e32b2aaf4c61c410eb925426119e1a9dc53d4286ade99a809",
			"passwordShow": "OTBhM2VkOWUzMmIyYWFmNGM2MWM0MTBlYjkyNTQyNjExOWUxYTlkYzUzZDQyODZhZGU5OWE4MDk=",
			"quota":        "-1",
		})
		if err != nil {
			return errors.New("错误[jghgsd]: " + err.Error())
		}
	}

	if utils.InSlice(&u, "system_info") < 0 {
		global.Log.Infoln("开始创建system_info表[df9sijh]")

		_, err = tx.Exec("CREATE TABLE `system_info` ( `id` int unsigned NOT NULL AUTO_INCREMENT, `key` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '', `value` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '', `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, PRIMARY KEY (`id`), UNIQUE KEY `idx_key` (`key`) USING BTREE ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统配置表';")

		if err != nil {
			return errors.New("错误[ghjdfgs]: " + err.Error())
		}
	}
	return
}

func (s sqlite) connect() (err error) {
	global.Log.Infoln("连接sqlite服务[poeriw]: ", global.Config.Env.Db.SqlitePath)

	DB, err = sqlx.Open("sqlite3", global.Config.Env.Db.SqlitePath)
	if err != nil {
		return errors.New("数据库连接失败[jgadsfgas]: " + err.Error())
	}
	err = DB.Ping() //没有数据库会创建
	if err != nil {
		return errors.New("数据库连接失败[khdsfgs]: " + err.Error())
	}
	DB.SetMaxOpenConns(20)
	DB.SetMaxIdleConns(10)

	return
}
func (m mysql) connect() (err error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", global.Config.Env.Db.MysqlUsername, global.Config.Env.Db.MysqlPassword, global.Config.Env.Db.MysqlHost, global.Config.Env.Db.MysqlPort, global.Config.Env.Db.MysqlDbname)

	global.Log.Infoln("连接mysql服务[giodjg]: ", dsn)

	//也可以使用MustConnect连接不成功就panic
	DB, err = sqlx.Connect("mysql", dsn)
	if err != nil {
		return errors.New("数据库连接失败[ujefaf]: " + err.Error())
	}

	DB.SetMaxOpenConns(20)
	DB.SetMaxIdleConns(10)

	return
}

func Close() (err error) {
	return DB.Close()
}

func Init() {
	dbRes, err := new(db).setDB(global.Config.Env.Db.Type)
	if err != nil {
		panic(err)
	}
	if dbRes == nil {
		b, _ := json.Marshal(global.Config.Env.Db)
		panic("缺少数据库信息[igoujsd]: " + string(b))
	}
	err = dbRes.createTable()
	if err != nil {
		panic(err)
	}
}
