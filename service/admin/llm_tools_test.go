package admin

import (
	"context"
	"strings"
	"testing"
)

func TestLLMTools(t *testing.T) {
	// 复用 bot_test.go 中预置的纯净内存数据库环境
	setupTestEnvironment(t)
	ctx := context.Background()

	// 测试 1: 创建用户
	t.Run("CreateUser", func(t *testing.T) {
		// 模拟 LLM 传来的严格 JSON 参数
		args := `{"username": "llm_tester", "expiry_date": "2030-01-01", "quota": 10}`
		res, err := executeTool(ctx, "create_user", args)
		if err != nil {
			t.Fatalf("创建用户工具执行失败: %v", err)
		}
		if !strings.Contains(res, "创建成功") {
			t.Errorf("预期包含'创建成功', 实际返回: %s", res)
		}
		// 校验是否按照刚才的逻辑生成了订阅链接
		if !strings.Contains(res, "Clash: ") || !strings.Contains(res, "V2ray: ") {
			t.Errorf("预期包含订阅链接, 实际返回: %s", res)
		}
	})

	// 测试 2: 创建重复用户时的业务拦截
	t.Run("CreateDuplicateUser", func(t *testing.T) {
		args := `{"username": "llm_tester", "expiry_date": null, "quota": 10}`
		res, err := executeTool(ctx, "create_user", args)
		if err != nil {
			t.Fatalf("预期不报代码错误，由业务层拦截: %v", err)
		}
		if !strings.Contains(res, "已存在") {
			t.Errorf("预期拦截重复用户, 实际返回: %s", res)
		}
	})

	// 测试 3: 查询用户
	t.Run("GetUser", func(t *testing.T) {
		args := `{"username": "llm_tester"}`
		res, err := executeTool(ctx, "get_user", args)
		if err != nil {
			t.Fatalf("查询用户失败: %v", err)
		}
		if !strings.Contains(res, "llm_tester") || !strings.Contains(res, "10.00G") {
			t.Errorf("返回内容缺失用户名或配额转换有误: %s", res)
		}
	})

	// 测试 4: 更新流量配额
	t.Run("UpdateQuota", func(t *testing.T) {
		// 修改为 -1 表示不限
		args := `{"username": "llm_tester", "quota": -1}`
		res, err := executeTool(ctx, "update_quota", args)
		if err != nil {
			t.Fatalf("更新流量失败: %v", err)
		}
		if !strings.Contains(res, "不限") {
			t.Errorf("预期流量更新为不限, 实际返回: %s", res)
		}
	})

	// 测试 5: 更新到期时间
	t.Run("UpdateExpiry", func(t *testing.T) {
		args := `{"username": "llm_tester", "expiry_date": "2035-12-31"}`
		res, err := executeTool(ctx, "update_expiry", args)
		if err != nil {
			t.Fatalf("更新到期时间失败: %v", err)
		}
		if !strings.Contains(res, "2035-12-31") {
			t.Errorf("预期时间已被更新, 实际返回: %s", res)
		}
	})

	// 测试 6: 列表展示
	t.Run("ListUsers", func(t *testing.T) {
		res, err := executeTool(ctx, "list_users", `{}`)
		if err != nil {
			t.Fatalf("查询列表失败: %v", err)
		}
		if !strings.Contains(res, "llm_tester") {
			t.Errorf("预期列表中包含目标用户, 实际: %s", res)
		}
	})

	// 测试 7: 删除用户
	t.Run("DeleteUser", func(t *testing.T) {
		args := `{"username": "llm_tester"}`
		res, err := executeTool(ctx, "delete_user", args)
		if err != nil {
			t.Fatalf("删除失败: %v", err)
		}
		if !strings.Contains(res, "成功删除") {
			t.Errorf("删除提示错误: %s", res)
		}

		// 校验确实被删除了
		res2, _ := executeTool(ctx, "get_user", args)
		if !strings.Contains(res2, "不存在") {
			t.Errorf("预期查询不到用户, 实际返回: %s", res2)
		}
	})
}
