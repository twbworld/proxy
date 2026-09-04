package initialize

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/twbworld/proxy/dao"
	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/internal/mysql"
	"github.com/twbworld/proxy/internal/redis"
	"github.com/twbworld/proxy/internal/sqlite"
	"github.com/twbworld/proxy/internal/telegram"
	"github.com/twbworld/proxy/model/enum"
)

func (i *Initializer) InitTz() error {
	Location, err := time.LoadLocation(global.Config().Tz)
	if err != nil {
		return fmt.Errorf("时区配置失败[siortuj]: %w", err)
	}
	global.Tz = Location
	return nil
}

// initNotify 根据配置初始化对应的通知系统 (如 Telegram / Email / 钉钉 等)
func (i *Initializer) initNotify() error {
	client, err := telegram.NewClient(telegram.Config{
		Token:       global.Config().Telegram.Token,
		Domain:      global.Config().Domain,
		ProjectName: global.Config().ProjectName,
		TelegramId:  global.Config().Telegram.Id,
		Debug:       global.Config().Debug,
	}, global.Log)
	if err != nil {
		global.Notifier = telegram.NewNoopClient(global.Log)
		return fmt.Errorf("通知服务降级为空操作模式: %w", err)
	}

	global.Notifier = client
	return nil
}

// notifyClose 统一关闭和释放通知服务资源
func (i *Initializer) notifyClose() error {
	return global.Notifier.Clear()
}

// initRedis 初始化Redis客户端
func (i *Initializer) initRedis(ctx context.Context) error {
	client, err := redis.NewClient(
		ctx,
		global.Config().Redis.Addr,
		global.Config().Redis.Password,
		int(global.Config().Redis.DB),
		global.Log,
	)
	if err != nil {
		return fmt.Errorf("连接Redis异常: %w", err)
	}
	global.RedisClient = client
	return nil
}

// redisClose 关闭Redis客户端连接
func (i *Initializer) redisClose() error {
	if global.RedisClient != nil {
		return global.RedisClient.Close()
	}
	return nil
}

// initDB 初始化数据库客户端
func (i *Initializer) initDB() error {
	var (
		driver  global.DBService
		addrMsg string
		err     error
	)

	switch global.Config().Database.Type {
	case string(enum.MYSQL):
		hostPort := net.JoinHostPort(global.Config().Database.MysqlHost, global.Config().Database.MysqlPort)
		dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			global.Config().Database.MysqlUsername,
			global.Config().Database.MysqlPassword,
			hostPort,
			global.Config().Database.MysqlDbname,
		)
		addrMsg = fmt.Sprintf("%s地址: @tcp(%s)/%s", global.Config().Database.Type, hostPort, global.Config().Database.MysqlDbname)
		driver, err = mysql.NewClient(dsn, global.Log)

	case string(enum.SQLITE):
		fallthrough
	default:
		addrMsg = global.Config().Database.SqlitePath
		driver, err = sqlite.NewClient(global.Config().Database.SqlitePath, global.Log)
	}

	if err != nil {
		return err
	}

	// 统一绑定全局 DAO
	dao.DB = driver.DB()
	dao.CanLock = driver.CanLock()

	global.Log.Infof("%s版本: %s; 地址: %s", global.Config().Database.Type, driver.Version(), addrMsg)

	// 统一初始化数据表
	if err = driver.CreateTable(); err != nil {
		return fmt.Errorf("创建数据表失败: %w", err)
	}

	return nil
}

func (i *Initializer) dbClose() error {
	if dao.DB != nil {
		return dao.DB.Close()
	}
	return nil
}
