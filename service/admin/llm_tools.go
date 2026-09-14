package admin

import (
	"context"
	"database/sql"
	"encoding/json/v2"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sashabaranov/go-openai"
	"github.com/twbworld/proxy/dao"
	"github.com/twbworld/proxy/model/common"
	"github.com/twbworld/proxy/model/db"
)

func getToolDefinitions() []openai.Tool {
	return []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "get_user",
				Description: "查询指定用户的详细信息",
				Strict:      true,
				Parameters: map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"properties": map[string]any{
						"username": map[string]string{"type": "string", "description": "用户名"},
					},
					"required": []string{"username"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "create_user",
				Description: "新增一个用户。只要有用户名即可创建；到期时间和流量配额均为可选，切勿追问用户",
				Strict:      true,
				Parameters: map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"properties": map[string]any{
						"username": map[string]string{
							"type":        "string",
							"description": "用户名，4-64 个字符的英文/数字/下划线/中划线",
						},
						"expiry_date": map[string]any{
							"type":        []string{"string", "null"},
							"description": "到期时间，格式 YYYY-MM-DD。若未指定必须传 null",
						},
						"quota": map[string]any{
							"type":        []string{"integer", "null"},
							"description": "流量配额（GB），-1 表示不限流量。若未指定必须传 null",
						},
					},
					"required": []string{"username", "expiry_date", "quota"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "update_quota",
				Description: "修改用户的流量配额",
				Strict:      true,
				Parameters: map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"properties": map[string]any{
						"username": map[string]string{"type": "string", "description": "用户名"},
						"quota":    map[string]string{"type": "integer", "description": "新的流量配额（GB），-1表示不限"},
					},
					"required": []string{"username", "quota"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "update_expiry",
				Description: "修改用户的到期时间",
				Strict:      true,
				Parameters: map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"properties": map[string]any{
						"username":    map[string]string{"type": "string", "description": "用户名"},
						"expiry_date": map[string]string{"type": "string", "description": "新的到期时间，格式 YYYY-MM-DD"},
					},
					"required": []string{"username", "expiry_date"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "delete_user",
				Description: "删除指定的用户",
				Strict:      true,
				Parameters: map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"properties": map[string]any{
						"username": map[string]string{"type": "string", "description": "要删除的用户名"},
					},
					"required": []string{"username"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "list_users",
				Description: "列出所有用户的简要信息",
				Strict:      true,
				Parameters: map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"properties":           map[string]any{},
					"required":             []string{},
				},
			},
		},
	}
}

// executeTool 根据工具名称和参数执行对应的业务逻辑
func executeTool(ctx context.Context, toolName string, argsJSON string) (string, error) {
	switch toolName {
	case "get_user":
		return toolGetUser(ctx, argsJSON)
	case "create_user":
		return toolCreateUser(ctx, argsJSON)
	case "update_quota":
		return toolUpdateQuota(ctx, argsJSON)
	case "update_expiry":
		return toolUpdateExpiry(ctx, argsJSON)
	case "delete_user":
		return toolDeleteUser(ctx, argsJSON)
	case "list_users":
		return toolListUsers(ctx, argsJSON)
	default:
		return "", fmt.Errorf("未知工具：%s", toolName)
	}
}

// toolGetUser 查询用户信息
func toolGetUser(ctx context.Context, argsJSON string) (string, error) {
	var params common.GetUserParams
	if err := json.Unmarshal([]byte(argsJSON), &params); err != nil {
		return "", fmt.Errorf("参数解析失败：%w", err)
	}
	if params.Username == "" {
		return "", errors.New("用户名不能为空")
	}

	user := new(db.Users)
	if err := dao.App.GetByCol(user, "username", params.Username); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Sprintf("用户 `%s` 不存在", params.Username), nil
		}
		return "", fmt.Errorf("查询失败：%w", err)
	}

	quotaStr := "不限"
	if user.Quota >= 0 {
		quotaStr = fmt.Sprintf("%.2fG", float64(user.Quota)/dao.QuotaMax)
	}
	return fmt.Sprintf("用户信息: \n用户名: %s\n配额: %s\n到期时间: %s\n已用上行: %d\n已用下行: %d",
		user.Username, quotaStr, user.ExpiryDate, user.Upload, user.Download), nil
}

// toolCreateUser 创建新用户
func toolCreateUser(ctx context.Context, argsJSON string) (string, error) {
	var params common.CreateUserParams
	if err := json.Unmarshal([]byte(argsJSON), &params); err != nil {
		return "", fmt.Errorf("参数解析失败：%w", err)
	}
	if params.Username == "" {
		return "", errors.New("用户名不能为空")
	}
	if !userNameRegex.MatchString(params.Username) {
		return "", errors.New("用户名格式不合法")
	}

	existingUser := new(db.Users)
	if dao.App.GetByCol(existingUser, "username", params.Username) == nil {
		return fmt.Sprintf("用户 `%s` 已存在", params.Username), nil
	}

	expiryDate := time.Now().In(time.UTC).AddDate(0, 1, 0).Format(time.DateOnly)
	if params.ExpiryDate != nil && *params.ExpiryDate != "" {
		expiryDate = *params.ExpiryDate
		if _, err := time.Parse(time.DateOnly, expiryDate); err != nil {
			return "", fmt.Errorf("日期格式错误：%w", err)
		}
	}

	quotaBytes := 0
	if params.Quota != nil {
		if *params.Quota == -1 {
			quotaBytes = -1
		} else if *params.Quota >= 0 {
			quotaBytes = int(float64(*params.Quota) * dao.QuotaMax)
		}
	}

	err := dao.Tx(func(tx *sqlx.Tx) error {
		return dao.App.UsersDb.InsertEmptyUsersWithCustomParams(params.Username, expiryDate, quotaBytes, tx)
	})
	if err != nil {
		return "", err
	}

	quotaStr := "不限"
	if quotaBytes >= 0 {
		quotaStr = fmt.Sprintf("%.2fG", float64(quotaBytes)/dao.QuotaMax)
	}
	return fmt.Sprintf("✅ 用户 `%s` 创建成功！\n• 流量配额：%s\n• 到期时间：%s%s",
		params.Username, quotaStr, expiryDate, generateLinksText(params.Username, false)), nil
}

// toolUpdateQuota 更新用户流量
func toolUpdateQuota(ctx context.Context, argsJSON string) (string, error) {
	var params common.UpdateQuotaParams
	if err := json.Unmarshal([]byte(argsJSON), &params); err != nil {
		return "", fmt.Errorf("参数解析失败：%w", err)
	}
	if params.Username == "" {
		return "", errors.New("用户名不能为空")
	}

	user := new(db.Users)
	if err := dao.App.GetByCol(user, "username", params.Username); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Sprintf("用户 `%s` 不存在", params.Username), nil
		}
		return "", fmt.Errorf("查询用户失败：%w", err)
	}

	quotaBytes := params.Quota
	if quotaBytes != -1 {
		quotaBytes = int(float64(quotaBytes) * dao.QuotaMax)
	}

	err := dao.Tx(func(tx *sqlx.Tx) error {
		user.Quota = quotaBytes
		return dao.App.UsersDb.UpdateUsers(user, tx)
	})
	if err != nil {
		return "", fmt.Errorf("更新失败：%w", err)
	}
	quotaStr := "不限"
	if params.Quota >= 0 {
		quotaStr = fmt.Sprintf("%.2fG", float64(params.Quota))
	}
	return fmt.Sprintf("✅ 用户 `%s` 流量配额已更新为：%s", params.Username, quotaStr), nil
}

// toolUpdateExpiry 更新用户到期时间
func toolUpdateExpiry(ctx context.Context, argsJSON string) (string, error) {
	var params common.UpdateExpiryParams
	if err := json.Unmarshal([]byte(argsJSON), &params); err != nil {
		return "", fmt.Errorf("参数解析失败：%w", err)
	}
	if params.Username == "" || params.ExpiryDate == "" {
		return "", errors.New("用户名或到期时间不能为空")
	}
	if _, err := time.Parse(time.DateOnly, params.ExpiryDate); err != nil {
		return "", fmt.Errorf("日期格式错误：%w", err)
	}

	user := new(db.Users)
	if err := dao.App.GetByCol(user, "username", params.Username); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Sprintf("用户 `%s` 不存在", params.Username), nil
		}
		return "", fmt.Errorf("查询用户失败：%w", err)
	}

	err := dao.Tx(func(tx *sqlx.Tx) error {
		user.ExpiryDate = params.ExpiryDate
		return dao.App.UsersDb.UpdateUsers(user, tx)
	})
	if err != nil {
		return "", fmt.Errorf("更新失败：%w", err)
	}
	return fmt.Sprintf("✅ 用户 `%s` 到期时间已更新为：%s", params.Username, params.ExpiryDate), nil
}

// toolDeleteUser 删除用户
func toolDeleteUser(ctx context.Context, argsJSON string) (string, error) {
	var params common.DeleteUserParams
	if err := json.Unmarshal([]byte(argsJSON), &params); err != nil {
		return "", fmt.Errorf("参数解析失败：%w", err)
	}
	if params.Username == "" {
		return "", errors.New("用户名不能为空")
	}

	user := new(db.Users)
	if err := dao.App.GetByCol(user, "username", params.Username); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Sprintf("用户 `%s` 不存在", params.Username), nil
		}
		return "", fmt.Errorf("查询用户失败：%w", err)
	}

	err := dao.Tx(func(tx *sqlx.Tx) error {
		return dao.App.UsersDb.DelUsersHandle(user.Id, tx)
	})
	if err != nil {
		return "", fmt.Errorf("删除失败：%w", err)
	}
	return fmt.Sprintf("✅ 用户 `%s` 已成功删除", params.Username), nil
}

// toolListUsers 列出所有用户
func toolListUsers(ctx context.Context, argsJSON string) (string, error) {
	var users []db.Users
	if err := dao.App.GetUserNames(&users); err != nil {
		return "", fmt.Errorf("查询失败：%w", err)
	}
	if len(users) == 0 {
		return "当前没有任何用户", nil
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "共有 %d 个用户：\n", len(users))
	for i, u := range users {
		fmt.Fprintf(&sb, "%d. `%s` (ID: %d)\n", i+1, u.Username, u.Id)
	}
	return sb.String(), nil
}
