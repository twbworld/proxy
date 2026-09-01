package sqlite

import (
	"fmt"
	"slices"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"github.com/twbworld/proxy/model/db"
)

type Client struct {
	log     *logrus.Logger
	version string
	db      *sqlx.DB
}

// NewClient 创建并初始化一个 SQLite 客户端实例
func NewClient(dsn string, log *logrus.Logger) (*Client, error) {
	// 连接 SQLite 数据库
	dbConn, err := sqlx.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败[siodjf]: %w", err)
	}

	// 连接池基础配置
	dbConn.SetMaxOpenConns(16)
	dbConn.SetMaxIdleConns(8)
	dbConn.SetConnMaxLifetime(time.Minute * 5)

	if err = dbConn.Ping(); err != nil {
		return nil, fmt.Errorf("数据库Ping失败[siodjf]: %w", err)
	}

	// SQLite 高性能与并发参数调优
	if _, err = dbConn.Exec("PRAGMA journal_mode = WAL;"); err != nil {
		return nil, fmt.Errorf("设置WAL模式失败: %w", err)
	}
	if _, err = dbConn.Exec("PRAGMA busy_timeout = 10000;"); err != nil {
		return nil, fmt.Errorf("设置超时等待失败: %w", err)
	}
	if _, err = dbConn.Exec("PRAGMA synchronous = NORMAL;"); err != nil {
		return nil, fmt.Errorf("设置同步模式失败: %w", err)
	}

	// 查询 SQLite 版本
	var version string
	if err := dbConn.Get(&version, `SELECT sqlite_version()`); err != nil && log != nil {
		log.Warnf("查询sqlite版本失败: %v", err)
	}

	return &Client{
		log:     log,
		version: version,
		db:      dbConn,
	}, nil
}

// DB 获取原生 sqlx.DB 连接对象
func (c *Client) DB() *sqlx.DB {
	return c.db
}

// CanLock SQLite 不支持 FOR UPDATE 行级锁，返回 false
func (c *Client) CanLock() bool {
	return false
}

// Version 获取 SQLite 版本信息
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

// CreateTable 执行 SQLite 特有的表、索引创建及初始数据写入
func (c *Client) CreateTable() error {
	var existingTables []string
	// 查询 sqlite_master 获取已存在的表
	err := c.db.Select(&existingTables, "SELECT name _id FROM sqlite_master WHERE type ='table'")
	if err != nil {
		return fmt.Errorf("查询数据库表失败: %w", err)
	}

	// 统一维护 SQLite 专属的 DDL 与索引语句
	sqls := map[string][]string{
		db.Users{}.TableName(): {
			`CREATE TABLE "%[1]s" (
				"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
				"username" text(64) NOT NULL DEFAULT '',
				"password" text(56) NOT NULL DEFAULT '',
				"quota" integer NOT NULL DEFAULT 0,
				"download" integer NOT NULL DEFAULT 0,
				"upload" integer NOT NULL DEFAULT 0,
				"expiryDate" text(10) NOT NULL DEFAULT '',
				"created_at" integer NOT NULL DEFAULT 0,
				"updated_at" integer NOT NULL DEFAULT 0
		  	);`,
			`CREATE INDEX "username" ON "%[1]s" ( "username" ASC );`,
		},
		db.SystemInfo{}.TableName(): {
			`CREATE TABLE "%[1]s" (
				"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
				"key" text(255) NOT NULL DEFAULT '',
				"value" text(255) NOT NULL DEFAULT '',
				"created_at" integer NOT NULL DEFAULT 0,
				"updated_at" integer NOT NULL DEFAULT 0
			);`,
			`CREATE UNIQUE INDEX "idx_key" ON "%[1]s" ( "key" ASC );`,
		},
	}

	// 事务执行建表与初始数据注入
	tx, err := c.db.Beginx()
	if err != nil {
		return fmt.Errorf("开启建表事务失败: %w", err)
	}
	panicked := true
	defer func() {
		if panicked || err != nil {
			tx.Rollback()
		}
	}()

	for tableName, sqlList := range sqls {
		if !slices.Contains(existingTables, tableName) {
			// 执行建表和索引
			for _, sqlTpl := range sqlList {
				if _, e := tx.Exec(fmt.Sprintf(sqlTpl, tableName)); e != nil {
					err = fmt.Errorf("执行建表SQL错误[ghjbcvgs]: %s\n%w", sqlTpl, e)
					return err
				}
			}

			if c.log != nil {
				c.log.Infof("创建%s表[dkyjh]", tableName)
			}

			// 插入默认种子数据
			if e := c.insertInitData(tableName, tx); e != nil {
				err = fmt.Errorf("插入初始数据失败: %s\n%w", tableName, e)
				return err
			}
		}
	}

	err = tx.Commit()
	panicked = false
	return err
}

// insertInitData 插入 SQLite 的默认初始数据
func (c *Client) insertInitData(tableName string, tx *sqlx.Tx) error {
	var sqlStrs []string

	switch tableName {
	case db.Users{}.TableName():
		sqlStrs = []string{
			fmt.Sprintf("INSERT INTO `%s`(`username`, `password`, `quota`) VALUES('test', '90a3ed9e32b2aaf4c61c410eb925426119e1a9dc53d4286ade99a809', -1)", tableName),
		}
	}

	for _, sqlStr := range sqlStrs {
		if c.log != nil {
			c.log.Infof("创建数据[dfskkjh]: %s", sqlStr)
		}
		if _, e := tx.Exec(sqlStr); e != nil {
			return fmt.Errorf("执行插入失败[gh90iggs]: %s\n%w", sqlStr, e)
		}
	}
	return nil
}
