package user

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/model/config"
	"github.com/twbworld/proxy/model/db"
	"github.com/twbworld/proxy/utils"
)

// initTestConfig 初始化单元测试所需的全局配置与环境
func initTestConfig(t *testing.T) {
	// 优先使用新标准库特性 t.TempDir() 自动管理测试生命周期内的临时目录与清理
	clashPath := filepath.Join(t.TempDir(), "clash.yaml")

	// 构造测试用的 Clash 模板文件
	clashTpl := "proxies:\n[proxies]\nproxy-groups:\n  - name: Auto\n    proxies: \n[proxies_name]"
	if err := os.WriteFile(clashPath, []byte(clashTpl), 0644); err != nil {
		t.Fatalf("初始化测试 clash 模板失败: %v", err)
	}

	cfg := &config.Config{
		ClashPath: clashPath,
		Proxies: []config.Proxies{
			{
				Name:    "测试节点_TCP",
				Type:    "vless",
				Server:  "127.0.0.1",
				Port:    443,
				Uuid:    "mock-uuid-1",
				Network: "tcp",
				Root:    false,
			},
			{
				Name:    "测试节点_Root",
				Type:    "vless",
				Server:  "127.0.0.2",
				Port:    443,
				Uuid:    "mock-uuid-root",
				Network: "tcp",
				Root:    true,
			},
		},
	}

	// 利用底层实现的无锁原子替换加载配置，保障并发安全
	global.InitConfig(cfg)
}

func TestIndexService_Handle(t *testing.T) {
	initTestConfig(t)
	svc := NewIndexService()

	now := time.Now().UTC()
	futureDate := now.AddDate(1, 0, 0).Format(time.DateOnly)
	pastDate := now.AddDate(-1, 0, 0).Format(time.DateOnly)

	tests := []struct {
		name        string
		protocol    string
		user        *db.Users
		wantContain []string
		isBase64    bool
	}{
		{
			name:     "过期用户_Clash兜底订阅",
			protocol: "clash",
			user:     &db.Users{ExpiryDate: pastDate},
			wantContain: []string{
				"!!! 订阅已过期 !!!",
				"!!!!!! 订阅已过期 !!!!!!",
			},
			isBase64: false,
		},
		{
			name:     "过期用户_V2ray兜底订阅",
			protocol: "v2ray",
			user:     &db.Users{ExpiryDate: pastDate},
			wantContain: []string{
				"vless://0@cn.bing.com:80?type=tcp#!!! 订阅已过期 !!!",
			},
			isBase64: true,
		},
		{
			name:     "普通有效用户_Clash订阅(过滤Root节点)",
			protocol: "clash",
			user:     &db.Users{Quota: 1024, ExpiryDate: futureDate},
			wantContain: []string{
				"测试节点_TCP",
				"127.0.0.1",
			},
			isBase64: false,
		},
		{
			name:     "无限流量有效用户_V2ray订阅(包含Root节点)",
			protocol: "v2ray",
			user:     &db.Users{Quota: -1, ExpiryDate: futureDate}, // -1 表示无限流量
			wantContain: []string{
				"mock-uuid-1@127.0.0.1:443",
				"mock-uuid-root@127.0.0.2:443",
			},
			isBase64: true,
		},
		{
			name:     "未知协议_回退为V2ray订阅",
			protocol: "unknown",
			user:     &db.Users{Quota: 1024, ExpiryDate: futureDate},
			wantContain: []string{
				"mock-uuid-1@127.0.0.1:443",
			},
			isBase64: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := svc.Handle(tc.protocol, tc.user)

			if tc.isBase64 {
				got = utils.Base64Decode(got)
			}

			// 核心逻辑断言
			for _, want := range tc.wantContain {
				if !strings.Contains(got, want) {
					t.Errorf("协议[%s] 期望结果包含:\n%s\n实际返回:\n%s", tc.protocol, want, got)
				}
			}

			// 边界测试：非无限流量的普通用户不能越权获取到 Root 节点信息
			if tc.user.Quota != -1 && !tc.isBase64 {
				if strings.Contains(got, "测试节点_Root") || strings.Contains(got, "mock-uuid-root") {
					t.Errorf("严重错误: 普通用户不应该包含 Root 节点信息")
				}
			}
		})
	}
}
