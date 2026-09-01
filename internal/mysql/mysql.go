package mysql

import (
	"fmt"
	"slices"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/twbworld/proxy/model/db"
)

type Client struct {
	log     *logrus.Logger
	version string
	db      *sqlx.DB
}

// NewClient 创建一个新的 MySQL 客户端实例
func NewClient(dsn string, log *logrus.Logger) (*Client, error) {
	dbConn, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败[rwbhe3]: %s\n%w", dsn, err)
	}

	dbConn.SetMaxOpenConns(16)
	dbConn.SetMaxIdleConns(8)
	dbConn.SetConnMaxLifetime(time.Minute * 5)

	if err = dbConn.Ping(); err != nil {
		return nil, fmt.Errorf("数据库连接检测失败: %s\n%w", dsn, err)
	}

	var version string
	if err := dbConn.Get(&version, `SELECT version()`); err != nil {
		log.Warnf("查询mysql版本失败: %v", err)
	}

	return &Client{
		log:     log,
		version: version,
		db:      dbConn,
	}, nil
}

func (c *Client) DB() *sqlx.DB {
	return c.db
}

func (c *Client) CanLock() bool {
	return true
}

func (c *Client) Version() string {
	return c.version
}

// Close 安全关闭数据库连接
func (c *Client) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

func (c *Client) CreateTable() error {
	var u []string
	err := c.db.Select(&u, "SHOW TABLES")
	if err != nil {
		return fmt.Errorf("查询表失败: %w", err)
	}

	sqls := map[string][]string{
		db.Users{}.TableName(): {
			"CREATE TABLE `%[1]s` ( `id` int unsigned NOT NULL AUTO_INCREMENT, `username` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '用户名', `password` char(56) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '密码', `quota` bigint NOT NULL DEFAULT '0' COMMENT '流量限制, 单位byte,1G=1073741824byte;-1:不限', `download` bigint unsigned NOT NULL DEFAULT '0' COMMENT '下行流量', `upload` bigint unsigned NOT NULL DEFAULT '0' COMMENT '上行流量', `expiryDate` char(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT '' COMMENT '限期; 年-月-日', `created_at` bigint NOT NULL DEFAULT 0 COMMENT '创建时间', `updated_at` bigint NOT NULL DEFAULT 0 COMMENT '更新时间', PRIMARY KEY (`id`), KEY `username` (`username`) ) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='用户表';",
		},
		db.SystemInfo{}.TableName(): {
			"CREATE TABLE `%[1]s` ( `id` int unsigned NOT NULL AUTO_INCREMENT, `key` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '', `value` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '', `created_at` bigint NOT NULL DEFAULT 0 COMMENT '创建时间', `updated_at` bigint NOT NULL DEFAULT 0 COMMENT '更新时间', PRIMARY KEY (`id`), UNIQUE KEY `idx_key` (`key`) USING BTREE ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统配置表';",
		},
	}

	tx, err := c.db.Beginx()
	if err != nil {
		return fmt.Errorf("开启事务失败: %w", err)
	}

	panicked := true
	defer func() {
		if panicked || err != nil {
			tx.Rollback()
		}
	}()

	for k, vList := range sqls {
		if !slices.Contains(u, k) {
			for _, sqlTpl := range vList {
				if _, e := tx.Exec(fmt.Sprintf(sqlTpl, k)); e != nil {
					err = fmt.Errorf("执行建表SQL失败: %s\n%w", k, e)
					return err
				}
			}
			c.log.Infof("创建%s表[dfsjh]", k)

			if err = c.insertInitData(k, tx); err != nil {
				err = fmt.Errorf("插入数据失败[fnko9]: %s\n%w", k, err)
				return err
			}
		}
	}

	err = tx.Commit()
	panicked = false
	if err != nil {
		return fmt.Errorf("事务执行失败: %w", err)
	}
	return nil
}

// insertInitData 插入 MySQL 的默认初始数据
func (c *Client) insertInitData(tableName string, tx *sqlx.Tx) error {
	var sqlStrs []string

	switch tableName {
	case db.Users{}.TableName():
		sqlStrs = []string{
			fmt.Sprintf("INSERT INTO `%s`(`username`, `password`, `quota`) VALUES('test', '90a3ed9e32b2aaf4c61c410eb925426119e1a9dc53d4286ade99a809', -1)", tableName),
		}
	}

	for _, sqlStr := range sqlStrs {
		c.log.Infof("创建数据[dfskkjh]: %s", sqlStr)
		if _, e := tx.Exec(sqlStr); e != nil {
			return fmt.Errorf("执行插入失败[gh90iggs]: %s\n%w", sqlStr, e)
		}
	}
	return nil
}
