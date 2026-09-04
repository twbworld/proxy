package redis

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

var ErrNil = redis.Nil

type Client struct {
	rdb *redis.Client
}

// NewClient 创建一个新的Redis客户端实例
func NewClient(ctx context.Context, addr, password string, db int, log *logrus.Logger) (*Client, error) {
	if addr == "" {
		log.Warnln("redis服务未启动: 配置为空")
		return nil, fmt.Errorf("redis配置为空")
	}

	if _, _, err := net.SplitHostPort(addr); err != nil {
		return nil, fmt.Errorf("配置错误: %w", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
		PoolSize: 10, // 连接池大小
	})

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if _, err := rdb.Ping(pingCtx).Result(); err != nil {
		if netErr, ok := errors.AsType[*net.OpError](err); ok {
			return nil, fmt.Errorf("连接Redis网络异常: %v", netErr.Error())
		}
		return nil, fmt.Errorf("连接Redis失败: %w", err)
	}

	return &Client{rdb: rdb}, nil
}

func (c *Client) Close() error {
	return c.rdb.Close()
}

func (c *Client) Get(ctx context.Context, key string) *redis.StringCmd {
	return c.rdb.Get(ctx, key)
}

func (c *Client) Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd {
	return c.rdb.Set(ctx, key, value, expiration)
}

func (c *Client) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	return c.rdb.Del(ctx, keys...)
}

func (c *Client) HGetAll(ctx context.Context, key string) *redis.StringStringMapCmd {
	return c.rdb.HGetAll(ctx, key)
}

func (c *Client) HSet(ctx context.Context, key string, values ...any) *redis.IntCmd {
	return c.rdb.HSet(ctx, key, values...)
}

func (c *Client) HDel(ctx context.Context, key string, fields ...string) *redis.IntCmd {
	return c.rdb.HDel(ctx, key, fields...)
}

func (c *Client) SetNX(ctx context.Context, key string, value any, expiration time.Duration) *redis.BoolCmd {
	return c.rdb.SetNX(ctx, key, value, expiration)
}

func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	return c.rdb.Expire(ctx, key, expiration)
}

func (c *Client) Ping(ctx context.Context) *redis.StatusCmd {
	return c.rdb.Ping(ctx)
}
