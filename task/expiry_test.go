package task

import (
	"io"
	"testing"
	"time"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"github.com/twbworld/proxy/dao"
	"github.com/twbworld/proxy/global"
)

// mockNotifier 用于捕获发出的通知信息，验证通知逻辑
type mockNotifier struct {
	msgs []string
}

func (m *mockNotifier) Send(text string) error {
	m.msgs = append(m.msgs, text)
	return nil
}
func (m *mockNotifier) Clear() error                   { return nil }
func (m *mockNotifier) SendRaw(msg tg.Chattable) error { return nil }

func TestManager_Expiry(t *testing.T) {
	// 抑制常规日志，防止污染测试终端
	global.Log = logrus.New()
	global.Log.SetOutput(io.Discard)
	global.Tz = time.UTC

	// 采用内存 SQLite 进行单元测试，速度快且与实际业务的 SQL 兼容
	dbConn, err := sqlx.Open("sqlite3", "file:test_expiry?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("开启内存数据库失败: %v", err)
	}
	t.Cleanup(func() { _ = dbConn.Close() })

	_, err = dbConn.Exec(`CREATE TABLE "users" (
		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		"username" text(64) NOT NULL DEFAULT '',
		"password" text(56) NOT NULL DEFAULT '',
		"quota" integer NOT NULL DEFAULT 0,
		"download" integer NOT NULL DEFAULT 0,
		"upload" integer NOT NULL DEFAULT 0,
		"expiryDate" text(10) NOT NULL DEFAULT '',
		"created_at" integer NOT NULL DEFAULT 0,
		"updated_at" integer NOT NULL DEFAULT 0
	)`)
	if err != nil {
		t.Fatalf("建表失败: %v", err)
	}

	// 劫持并替换全局依赖
	dao.DB = dbConn
	dao.CanLock = false
	mockNoti := &mockNotifier{}
	global.Notifier = mockNoti

	manager := &Manager{}

	t.Run("没有用户数据的场景", func(t *testing.T) {
		_, _ = dbConn.Exec("DELETE FROM users") // 确保干净环境
		if err := manager.Expiry(); err != nil {
			t.Errorf("期望返回 nil，实际出错: %v", err)
		}
	})

	t.Run("处理各种过期情况的场景", func(t *testing.T) {
		_, _ = dbConn.Exec("DELETE FROM users")

		now := time.Now().In(time.UTC)
		expiredDate := now.AddDate(0, 0, -2).Format(time.DateOnly)
		validDate := now.AddDate(0, 0, 5).Format(time.DateOnly)

		// 注：因原业务代码使用 t1.After(ti) && t2.Before(ti) 计算-7和-5存在悖论 (dead code)
		// 这里仍按原代码取值范围插入一个边缘日期做覆盖度测试
		edgeDate := now.AddDate(0, 0, -6).Format(time.DateOnly)

		query := `INSERT INTO users (username, quota, expiryDate) VALUES (?, 1000, ?)`
		for _, v := range []struct {
			name string
			date string
		}{
			{"expired_user", expiredDate},
			{"valid_user", validDate},
			{"edge_user", edgeDate},
			{"no_date_user", ""},
		} {
			// 测试场景下使用 MustExec 减少冗余的 error 断言
			dbConn.MustExec(query, v.name, v.date)
		}

		if err := manager.Expiry(); err != nil {
			t.Errorf("执行 Expiry 任务失败: %v", err)
		}

		// 利用结构体切片批量查询和映射验证，替代原有分散的多次单行 dbConn.Get 查询
		var results []struct {
			Username string `db:"username"`
			Quota    int    `db:"quota"`
		}
		_ = dbConn.Select(&results, "SELECT username, quota FROM users")

		expected := map[string]int{
			"expired_user": 0,
			"edge_user":    0,
			"valid_user":   1000,
			"no_date_user": 1000,
		}

		for _, u := range results {
			if want, ok := expected[u.Username]; ok && u.Quota != want {
				t.Errorf("期望用户(%s)配额为 %d，实际为 %d", u.Username, want, u.Quota)
			}
		}
	})
}
