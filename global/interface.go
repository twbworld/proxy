package global

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
)

// DBService 数据库抽象接口
type DBService interface {
	DB() *sqlx.DB
	CreateTable() error
	Version() string
	CanLock() bool
	Close() error
}

// RedisService Redis 抽象接口
type RedisService interface {
	Close() error
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	HGetAll(ctx context.Context, key string) *redis.StringStringMapCmd
	HSet(ctx context.Context, key string, values ...any) *redis.IntCmd
	HDel(ctx context.Context, key string, fields ...string) *redis.IntCmd
	SetNX(ctx context.Context, key string, value any, expiration time.Duration) *redis.BoolCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
	Ping(ctx context.Context) *redis.StatusCmd
}

// NotifierService 通知能力的抽象接口
type NotifierService interface {
	Send(text string) error
	Clear() error
	SendRaw(msg tg.Chattable) error //SendRaw 发送 TG 原生高级消息(含Inline键盘菜单等)
}
