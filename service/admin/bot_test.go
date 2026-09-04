package admin

import (
	"bytes"
	"encoding/json/v2"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/twbworld/proxy/dao"
	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/internal/sqlite"
	"github.com/twbworld/proxy/model/config"
)

// mockNotifier 模拟 Telegram 消息发送服务
type mockNotifier struct {
	sentRaw []tg.Chattable
	sent    []string
}

func (m *mockNotifier) Send(text string) error {
	m.sent = append(m.sent, text)
	return nil
}

func (m *mockNotifier) Clear() error { return nil }

func (m *mockNotifier) SendRaw(msg tg.Chattable) error {
	m.sentRaw = append(m.sentRaw, msg)
	return nil
}

// mockAsync 模拟异步协程服务，将其退化为同步阻塞调用，以确保数据的一致性及可断言性
type mockAsync struct{}

func (m *mockAsync) GoSafe(fn func()) {
	fn() // 单元测试中同步执行
}

// setupTestEnvironment 构建基于内存 SQLite 的纯净单元测试环境
func setupTestEnvironment(t *testing.T) {
	t.Helper()

	global.Log = logrus.New()
	global.Log.SetLevel(logrus.FatalLevel) // 屏蔽非必要日志
	global.Tz = time.UTC

	global.InitConfig(&config.Config{Domain: "https://www.example.com"})

	// 启用内存模式 SQLite
	client, err := sqlite.NewClient(":memory:", global.Log)
	if err != nil {
		t.Fatalf("初始化测试数据库失败: %v", err)
	}

	// 挂载全局实例
	dao.DB = client.DB()
	dao.CanLock = client.CanLock()

	// 自动执行 DDL 及 seed 初始化
	if err := client.CreateTable(); err != nil {
		t.Fatalf("创建数据表失败: %v", err)
	}

	t.Cleanup(func() {
		_ = client.Close()
	})

	gin.SetMode(gin.TestMode)
}

func TestBotService_Webhook(t *testing.T) {
	setupTestEnvironment(t)

	notifier := &mockNotifier{}
	async := &mockAsync{}
	svc := NewBotService(async, notifier)

	tests := []struct {
		name    string
		update  tg.Update
		setup   func(t *testing.T)
		verify  func(t *testing.T, n *mockNotifier)
		wantErr bool
	}{
		{
			name: "新对话处理 - 解析 /start 命令",
			update: tg.Update{
				Message: &tg.Message{
					Chat: &tg.Chat{ID: 1001},
					Text: "/start",
					Entities: []tg.MessageEntity{
						{Type: "bot_command", Offset: 0, Length: 6},
					},
					From: &tg.User{FirstName: "张", LastName: "三"},
				},
			},
			verify: func(t *testing.T, n *mockNotifier) {
				if len(n.sentRaw) == 0 {
					t.Fatal("期望调用 SendRaw 发送内联菜单消息")
				}
				msg, ok := n.sentRaw[0].(tg.MessageConfig)
				if !ok {
					t.Fatalf("返回的消息类型不匹配，期望 tg.MessageConfig，得到 %T", n.sentRaw[0])
				}
				if msg.Text != "\nHello 张三" {
					t.Errorf("回复文本不符: %s", msg.Text)
				}
			},
		},
		{
			name: "自然语言对话 - 触发 LLM 助理",
			setup: func(t *testing.T) {
				// 前置条件：设定状态机为空
				err := dao.Tx(func(tx *sqlx.Tx) error {
					return dao.App.SystemInfoDb.SaveSysVal("1003_step", "", tx)
				})
				if err != nil {
					t.Fatalf("前置数据构建失败: %v", err)
				}
			},
			update: tg.Update{
				Message: &tg.Message{
					Chat: &tg.Chat{ID: 1003},
					Text: "帮我查一下用户信息",
				},
			},
			verify: func(t *testing.T, n *mockNotifier) {
				// LLM 助理在 mock 环境下因为 RedisClient 为空，会返回降级消息
				if len(n.sentRaw) == 0 {
					t.Fatal("期望调用 SendRaw 发送 LLM 处理结果")
				}
				msg, ok := n.sentRaw[0].(tg.MessageConfig)
				if !ok {
					t.Fatalf("返回的消息类型不匹配，得到 %T", n.sentRaw[0])
				}
				if msg.ParseMode != "Markdown" {
					t.Errorf("预期开启 Markdown 渲染, 实际: %s", msg.ParseMode)
				}
				if msg.Text == "" {
					t.Errorf("期望返回提示文本，得到空文本")
				}
			},
		},
		{
			name: "内联菜单回调 - 触发 [查询用户]",
			update: tg.Update{
				CallbackQuery: &tg.CallbackQuery{
					Data: "user_select",
					Message: &tg.Message{
						Chat:      &tg.Chat{ID: 1001},
						MessageID: 1,
					},
				},
			},
			verify: func(t *testing.T, n *mockNotifier) {
				if len(n.sentRaw) == 0 {
					t.Fatal("期望调用 SendRaw 修改原有消息")
				}
				msg, ok := n.sentRaw[0].(tg.EditMessageTextConfig)
				if !ok {
					t.Fatalf("返回的消息类型不匹配，期望 tg.EditMessageTextConfig，得到 %T", n.sentRaw[0])
				}
				if msg.Text != "选择*查询*的用户" {
					t.Errorf("回复文本不符: %s", msg.Text)
				}
			},
		},
		{
			name: "内联菜单回调 - 触发 [新增用户]",
			update: tg.Update{
				CallbackQuery: &tg.CallbackQuery{
					Data: "user_insert",
					Message: &tg.Message{
						Chat:      &tg.Chat{ID: 1001},
						MessageID: 2,
					},
				},
			},
			verify: func(t *testing.T, n *mockNotifier) {
				if len(n.sentRaw) == 0 {
					t.Fatal("期望调用 SendRaw 下发输入指引")
				}
				msg, ok := n.sentRaw[0].(tg.MessageConfig)
				if !ok {
					t.Fatalf("返回的消息类型不匹配，期望 tg.MessageConfig，得到 %T", n.sentRaw[0])
				}
				expected := "请输入用户名称, 4\\-64个字符以内的英文/数字/符号\n例:`210606_abc`"
				if msg.Text != expected {
					t.Errorf("回复文本不符, 得到: %s", msg.Text)
				}
			},
		},
		{
			name: "对话第五步 - 捕获用户键入的注册账户名 (状态机处理)",
			setup: func(t *testing.T) {
				// 前置条件：设定状态机当前处于新增用户步骤中
				err := dao.Tx(func(tx *sqlx.Tx) error {
					return dao.App.SystemInfoDb.SaveSysVal("1002_step", "user_insert", tx)
				})
				if err != nil {
					t.Fatalf("前置数据构建失败: %v", err)
				}
			},
			update: tg.Update{
				Message: &tg.Message{
					Chat: &tg.Chat{ID: 1002},
					Text: "new_user_123", // 用户键入注册用户名
				},
			},
			verify: func(t *testing.T, n *mockNotifier) {
				if len(n.sentRaw) == 0 {
					t.Fatal("期望调用 SendRaw 发送操作结果")
				}
				msg, ok := n.sentRaw[0].(tg.MessageConfig)
				if !ok {
					t.Fatalf("返回的消息类型不匹配，期望 tg.MessageConfig，得到 %T", n.sentRaw[0])
				}
				if msg.Text == "" {
					t.Errorf("期望返回新增成功的渲染文本，得到空文本")
				}
			},
		},
		{
			name: "异常链路检测 - 空回调数据拦截",
			update: tg.Update{
				CallbackQuery: &tg.CallbackQuery{
					Data: "", // 恶意或错误传参
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 清理上一轮测试的执行态
			notifier.sentRaw = nil
			notifier.sent = nil

			if tt.setup != nil {
				tt.setup(t)
			}

			body, err := json.Marshal(tt.update)
			if err != nil {
				t.Fatalf("结构体序列化失败: %v", err)
			}

			// 模拟 Gin 请求生命周期
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			// 优先使用 NewRequestWithContext，并使用零拷贝的 bytes.NewReader
			ctx.Request, _ = http.NewRequestWithContext(ctx, http.MethodPost, "/", bytes.NewReader(body))

			err = svc.Webhook(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Webhook() 报错 = %v, 期望触发错误 = %v", err, tt.wantErr)
			}

			if tt.verify != nil {
				tt.verify(t, notifier)
			}
		})
	}
}

// 路径: service/admin/bot_test.go（末尾追加）
func TestGenerateLinksText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		domain    string
		username  string
		wantClash string
		wantV2ray string
	}{
		{
			name:      "带有 www 前缀域名",
			domain:    "https://www.example.com",
			username:  "test_user",
			wantClash: "https://clash.example.com/test_user.html",
			wantV2ray: "https://www.example.com/test_user.html",
		},
		{
			name:      "纯根域名",
			domain:    "https://example.com",
			username:  "test_user",
			wantClash: "https://clash.example.com/test_user.html",
			wantV2ray: "https://example.com/test_user.html",
		},
		{
			name:      "空域名容错",
			domain:    "",
			username:  "test_user",
			wantClash: "",
			wantV2ray: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			global.InitConfig(&config.Config{Domain: tt.domain})
			got := generateLinksText(tt.username, false)
			if tt.domain == "" {
				if got != "" {
					t.Errorf("空域名应返回空文本，实际得到: %s", got)
				}
				return
			}
			if !strings.Contains(got, tt.wantClash) || !strings.Contains(got, tt.wantV2ray) {
				t.Errorf("链接生成异常, clash期望: %s, v2ray期望: %s, 实际得到: %s", tt.wantClash, tt.wantV2ray, got)
			}
		})
	}
}
