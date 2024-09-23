package system

import (
	"fmt"
	"time"

	"github.com/twbworld/proxy/dao"
	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/model/db"
	"github.com/twbworld/proxy/utils"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type mysql struct{}
type sqlite struct{}
type class interface {
	connect() error
	createTable() error
	insertData(string, *sqlx.Tx) error
	version() string
}

func DbStart() error {
	var dbRes class

	switch global.Config.Database.Type {
	case "mysql":
		dbRes = &mysql{}
	case "sqlite":
		dbRes = &sqlite{}
	default:
		dbRes = &sqlite{}
	}

	if err := dbRes.connect(); err != nil {
		return err
	}
	dbRes.createTable()
	return nil
}

// 关闭数据库连接
func DbClose() error {
	if dao.DB != nil {
		return dao.DB.Close()
	}
	return nil
}

// 连接SQLite数据库
func (s *sqlite) connect() error {
	var err error

	if dao.DB, err = sqlx.Open("sqlite3", global.Config.Database.SqlitePath); err != nil {
		return fmt.Errorf("数据库连接失败: %w", err)
	}
	//没有数据库会创建
	if err = dao.DB.Ping(); err != nil {
		return fmt.Errorf("数据库连接失败: %w", err)
	}

	dao.DB.SetMaxOpenConns(16)
	dao.DB.SetMaxIdleConns(8)
	dao.DB.SetConnMaxLifetime(time.Minute * 5)

	//提高并发
	if _, err = dao.DB.Exec("PRAGMA journal_mode = WAL"); err != nil {
		return fmt.Errorf("数据库设置失败: %w", err)
	}
	//超时等待
	if _, err = dao.DB.Exec("PRAGMA busy_timeout = 10000;"); err != nil {
		return fmt.Errorf("数据库设置失败: %w", err)
	}
	// 设置同步模式为 NORMAL
	if _, err = dao.DB.Exec("PRAGMA synchronous = NORMAL;"); err != nil {
		return fmt.Errorf("数据库设置失败: %w", err)
	}

	dao.CanLock = false

	global.Log.Infof("%s版本: %s; 地址: %s", global.Config.Database.Type, s.version(), global.Config.Database.SqlitePath)
	return nil
}

func (m *mysql) connect() error {
	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", global.Config.Database.MysqlUsername, global.Config.Database.MysqlPassword, global.Config.Database.MysqlHost, global.Config.Database.MysqlPort, global.Config.Database.MysqlDbname)

	//也可以使用MustConnect连接不成功就panic
	if dao.DB, err = sqlx.Connect("mysql", dsn); err != nil {
		return fmt.Errorf("数据库连接失败[rwbhe3]: %s\n%w", dsn, err)
	}

	dao.DB.SetMaxOpenConns(16)
	dao.DB.SetMaxIdleConns(8)
	dao.DB.SetConnMaxLifetime(time.Minute * 5) // 设置连接的最大生命周期

	if err = dao.DB.Ping(); err != nil {
		return fmt.Errorf("数据库连接失败: %s\n%w", dsn, err)
	}

	dao.CanLock = true
	global.Log.Infof("%s版本: %s; 地址: @tcp(%s:%s)/%s", global.Config.Database.Type, m.version(), global.Config.Database.MysqlHost, global.Config.Database.MysqlPort, global.Config.Database.MysqlDbname)
	return nil
}

func (s *sqlite) createTable() error {
	var u []string
	err := dao.DB.Select(&u, "SELECT name _id FROM sqlite_master WHERE type ='table'")
	if err != nil {
		return fmt.Errorf("查询表失败: %w", err)
	}

	sqls := map[string][]string{
		db.Users{}.TableName(): {
			`CREATE TABLE "%s" (
				"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
				"username" text(64) NOT NULL DEFAULT '',
				"password" text(56) NOT NULL DEFAULT '',
				"passwordShow" text(255) NOT NULL DEFAULT '',
				"quota" integer NOT NULL DEFAULT 0,
				"download" integer NOT NULL DEFAULT 0,
				"upload" integer NOT NULL DEFAULT 0,
				"useDays" integer NOT NULL DEFAULT 0,
				"expiryDate" text(10) NOT NULL DEFAULT ''
		  	);`,
			`CREATE INDEX "password" ON "%s" ( "password" ASC );`,
		},
		db.SystemInfo{}.TableName(): {
			`CREATE TABLE "%s" (
				"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
				"key" text(255) NOT NULL DEFAULT '',
				"value" text(255) NOT NULL DEFAULT '',
				"update_time" text NOT NULL DEFAULT ''
			);`,
			`CREATE UNIQUE INDEX "idx_key" ON "%s" ( "key" ASC );`,
		},
	}

	err = dao.Tx(func(tx *sqlx.Tx) (e error) {
		for k, v := range sqls {
			if utils.InSlice(u, k) < 0 {
				for _, val := range v {
					if _, e := tx.Exec(fmt.Sprintf(val, k)); e != nil {
						return fmt.Errorf("错误[ghjbcvgs]:  %s\n%w", val, e)
					}
				}
				if err := s.insertData(k, tx); err != nil {
					return fmt.Errorf("插入数据失败: %s\n%w", k, err)
				}
				global.Log.Infof("创建%s表[dkyjh]", k)
			}
		}
		return
	})
	if err != nil {
		return fmt.Errorf("事务执行失败: %w", err)
	}
	return nil
}

func (m *mysql) createTable() error {
	var u []string
	err := dao.DB.Select(&u, "SHOW TABLES")
	if err != nil {
		return fmt.Errorf("插入数据失败: %w", err)
	}

	sqls := map[string]string{
		db.Users{}.TableName():      "CREATE TABLE `%s` ( `id` int unsigned NOT NULL AUTO_INCREMENT, `username` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '用户名', `password` char(56) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '密码', `passwordShow` varchar(255) NOT NULL, `quota` bigint NOT NULL DEFAULT '0' COMMENT '流量限制, 单位byte,1G=1073741824byte;-1:不限', `download` bigint unsigned NOT NULL DEFAULT '0' COMMENT '下行流量', `upload` bigint unsigned NOT NULL DEFAULT '0' COMMENT '上行流量', `useDays` int DEFAULT '0', `expiryDate` char(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT '' COMMENT '限期; 年-月-日', PRIMARY KEY (`id`), KEY `password` (`password`) ) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='用户表';",
		db.SystemInfo{}.TableName(): "CREATE TABLE `%s` ( `id` int unsigned NOT NULL AUTO_INCREMENT, `key` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '', `value` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '', `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, PRIMARY KEY (`id`), UNIQUE KEY `idx_key` (`key`) USING BTREE ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统配置表';",
	}

	err = dao.Tx(func(tx *sqlx.Tx) (e error) {
		for k, v := range sqls {
			if utils.InSlice(u, k) < 0 {
				if _, e := tx.Exec(fmt.Sprintf(v, k)); e != nil {
					return fmt.Errorf("插入数据失败: %s\n%w", k, err)
				}
				global.Log.Infof("创建%s表[dfsjh]", k)
				if err := m.insertData(k, tx); err != nil {
					return fmt.Errorf("插入数据失败[fnko9]: %s\n%w", k, err)
				}
			}
		}
		return
	})
	if err != nil {
		return fmt.Errorf("事务执行失败: %w", err)
	}
	return nil
}

func (m *mysql) insertData(t string, tx *sqlx.Tx) error {
	return insert(t, tx)
}

func (m *sqlite) insertData(t string, tx *sqlx.Tx) error {
	return insert(t, tx)
}

func insert(t string, tx *sqlx.Tx) error {
	var sqls []string

	switch t {
	case db.Users{}.TableName():
		sqls = []string{
			fmt.Sprintf("INSERT INTO `%s`(`username`, `password`, `passwordShow`, `quota`) VALUES('test', '90a3ed9e32b2aaf4c61c410eb925426119e1a9dc53d4286ade99a809', 'OTBhM2VkOWUzMmIyYWFmNGM2MWM0MTBlYjkyNTQyNjExOWUxYTlkYzUzZDQyODZhZGU5OWE4MDk=', -1)", db.Users{}.TableName()),
		}
	}

	for _, v := range sqls {
		global.Log.Infof("创建数据[dfskkjh]%s", v)
		if _, e := tx.Exec(v); e != nil {
			return fmt.Errorf("错误[gh90iggs]: %s\n%w", v, e)
		}
	}
	return nil
}

func (*sqlite) version() (t string) {
	dao.DB.Get(&t, `SELECT sqlite_version()`)
	return
}

func (*mysql) version() (t string) {
	dao.DB.Get(&t, `SELECT version()`)
	return
}
