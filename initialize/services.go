package initialize

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/twbworld/proxy/dao"
	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/internal/llm"
	"github.com/twbworld/proxy/internal/mysql"
	"github.com/twbworld/proxy/internal/redis"
	"github.com/twbworld/proxy/internal/sqlite"
	"github.com/twbworld/proxy/internal/telegram"
	"github.com/twbworld/proxy/model/enum"
	"golang.org/x/sync/errgroup"
)

type llmTransport struct {
	Base http.RoundTripper
}

func (t *llmTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// 1. 卫语句检查：非目标请求直接透传，0 级嵌套
	enabled, _ := req.Context().Value(llm.DisableThinkingKey).(bool)
	if !enabled || req.Body == nil {
		return t.Base.RoundTrip(req)
	}

	// 2. 注入参数（内部做好容错，失败也不中断主请求）
	_ = injectDisableThinking(req)

	return t.Base.RoundTrip(req)
}

// injectDisableThinking 篡改请求体，注入 {"enable_thinking": false}
func injectDisableThinking(req *http.Request) error {
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}
	// 防御性保底：读取后立即恢复未修改的 Body，确保后续无论在哪一步出错退出，原请求都不受损
	req.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	var bodyMap map[string]any
	if err := json.Unmarshal(bodyBytes, &bodyMap); err != nil {
		return err
	}

	bodyMap["enable_thinking"] = false

	newBodyBytes, err := json.Marshal(bodyMap)
	if err != nil {
		return err
	}

	// 覆写篡改后的请求体
	req.Body = io.NopCloser(bytes.NewReader(newBodyBytes))
	req.ContentLength = int64(len(newBodyBytes))
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(newBodyBytes)), nil
	}

	return nil
}

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

// initLlm 初始化LLM服务
func (i *Initializer) initLlm() error {
	if len(global.Config().LLM) == 0 {
		return fmt.Errorf("未配置任何LLM")
	}

	llms := global.Config().LLM

	llmClients := make(map[enum.LLMSize]*openai.Client, len(llms))
	for _, cfg := range llms {
		config := openai.DefaultConfig(cfg.Auth)
		config.BaseURL = cfg.Url
		config.HTTPClient = &http.Client{
			Timeout:   time.Duration(cfg.Timeout) * time.Second,
			Transport: &llmTransport{Base: http.DefaultTransport},
		}
		llmClients[enum.LLMSize(cfg.Size)] = openai.NewClientWithConfig(config)
	}

	g, gCtx := errgroup.WithContext(context.Background())
	// 并发地对所有配置的LLM服务进行连接测试
	for _, cfg := range llms {
		g.Go(func() error {
			size := enum.LLMSize(cfg.Size)
			client := llmClients[size]

			// 使用 errgroup 的 context，以便在任何一个 goroutine 失败时可以取消其他的
			reqCtx, cancel := context.WithTimeout(gCtx, 5*time.Second)
			defer cancel()

			// 通过ListModels接口验证服务是否可用
			if _, err := client.ListModels(reqCtx); err != nil {
				return fmt.Errorf("无法连接到LLM服务 (size: %s, url: %s): %w", size, cfg.Url, err)
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("初始化LLM服务失败: %v", err)
	}

	global.LlmService = llm.NewClient(
		global.Log,
		llmClients,
		llms,
	)
	return nil
}
