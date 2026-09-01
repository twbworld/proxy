package task

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/model/config"
)

func TestCleanUpLogs(t *testing.T) {
	// 初始化必要的全局依赖
	global.Tz = time.Local
	global.Log = logrus.New()
	global.Log.SetOutput(io.Discard) // 单元测试中丢弃日志输出

	// 利用 t.TempDir() 自动创建并清理临时测试目录，替代老旧的 os.MkdirTemp + defer os.RemoveAll
	tempDir := t.TempDir()
	now := time.Now().In(global.Tz)

	// 准备测试用例数据
	testCases := []struct {
		name     string
		expected bool // 期望是否保留
	}{
		{"run.log", true},                                                 // 无日期后缀，不应被清理
		{"run.log.invalid-date", true},                                    // 日期格式错误，不应被清理
		{"run.log." + now.Format(time.DateOnly), true},                    // 当天日志，保留
		{"run.log." + now.AddDate(0, 0, -6).Format(time.DateOnly), true},  // 6天前，保留
		{"run.log." + now.AddDate(0, 0, -7).Format(time.DateOnly), true},  // 7天前 (刚好等于 cutoffDate)，保留
		{"run.log." + now.AddDate(0, 0, -8).Format(time.DateOnly), false}, // 8天前 (Before cutoffDate)，需删除
	}

	// 批量创建测试日志文件
	for _, tc := range testCases {
		path := filepath.Join(tempDir, tc.name)
		if err := os.WriteFile(path, []byte("mock log data"), 0644); err != nil {
			t.Fatalf("创建测试文件失败: %v", err)
		}
	}

	// 注入全局配置
	cfg := &config.Config{
		LogRetentionDays: 7,
		RunLogPath:       filepath.Join(tempDir, "run.log"), // 让程序在 tempDir 下执行清理
	}
	global.InitConfig(cfg)

	// 执行目标函数
	m := NewManager()
	if err := m.CleanUpLogs(); err != nil {
		t.Fatalf("CleanUpLogs 运行失败: %v", err)
	}

	// 验证文件清理结果
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(tempDir, tc.name)
			_, err := os.Stat(path)
			exists := !os.IsNotExist(err)

			if exists != tc.expected {
				t.Errorf("文件存在状态异常! 期望存在: %v, 实际存在: %v", tc.expected, exists)
			}
		})
	}

	// 边界测试：当 RetentionDays 为 0 时，禁用清理功能
	t.Run("RetentionDisabled", func(t *testing.T) {
		cfg.LogRetentionDays = 0
		global.InitConfig(cfg)

		// 补齐一个过期文件
		oldLogName := "run.log." + now.AddDate(0, 0, -10).Format(time.DateOnly)
		oldLogPath := filepath.Join(tempDir, oldLogName)
		_ = os.WriteFile(oldLogPath, []byte("mock"), 0644)

		if err := m.CleanUpLogs(); err != nil {
			t.Fatalf("CleanUpLogs 运行失败: %v", err)
		}

		// 期望文件未被删除
		if _, err := os.Stat(oldLogPath); os.IsNotExist(err) {
			t.Error("当 LogRetentionDays 为 0 时，日志不应该被清理")
		}
	})
}
