package common

import (
	"fmt"

	"github.com/twbworld/proxy/model/db"
)

// FlexInt 灵活整型，专门处理 LLM 可能会返回字符串形式数字的幻觉问题 (兼容 "100" 和 100)
type FlexInt int

type ChatCompletionRequest struct {
	Model      string           `json:"model"`
	Messages   []ChatMessage    `json:"messages"`
	Tools      []ToolDefinition `json:"tools,omitempty"`
	MaxTokens  int              `json:"max_tokens,omitzero"`
	TimeoutSec int              `json:"-"`
}

type ChatMessage struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	Name       string     `json:"name,omitempty"`
}

type ToolDefinition struct {
	Type     string      `json:"type"`
	Function FunctionDef `json:"function"`
}

// FunctionDef 支持 OpenAI Structured Outputs
type FunctionDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Strict      bool           `json:"strict"` // 💡 启用严格结构化输出
	Parameters  ToolParameters `json:"parameters"`
}

type ToolParameters struct {
	Type                 string         `json:"type"`
	Properties           map[string]any `json:"properties"`
	Required             []string       `json:"required"`
	AdditionalProperties bool           `json:"additionalProperties"` // 💡 拒绝额外属性，防止幻觉
}

type ChatCompletionResponse struct {
	ID      string    `json:"id"`
	Choices []Choice  `json:"choices"`
	Error   *APIError `json:"error,omitempty"`
}

type Choice struct {
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// 工具入参结构体
type CreateUserParams struct {
	Username   string   `json:"username"`
	ExpiryDate *string  `json:"expiry_date"`
	Quota      *FlexInt `json:"quota"`
}
type UpdateQuotaParams struct {
	Username string  `json:"username"`
	Quota    FlexInt `json:"quota"`
}
type UpdateExpiryParams struct {
	Username   string `json:"username"`
	ExpiryDate string `json:"expiry_date"`
}
type DeleteUserParams struct {
	Username string `json:"username"`
}
type GetUserParams struct {
	Username string `json:"username"`
}

// FormatUserInfo 格式化用户信息
func FormatUserInfo(user *db.Users) string {
	if user == nil {
		return "用户不存在"
	}
	quotaStr := "不限"
	if user.Quota > 0 {
		quotaGB := float64(user.Quota) / 1024 / 1024 / 1024
		quotaStr = fmt.Sprintf("%.2fG", quotaGB)
	}
	usedGB := float64(user.Download+user.Upload) / 1024 / 1024 / 1024

	return fmt.Sprintf("• 用户名：%s\n• ID: %d\n• 流量配额：%s\n• 已用流量：%.2fG\n• 到期时间：%s",
		user.Username, user.Id, quotaStr, usedGB, user.ExpiryDate)
}
