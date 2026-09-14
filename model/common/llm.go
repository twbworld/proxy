package common

import (
	"github.com/sashabaranov/go-openai"
)

// LLMMessage 定义了发送给LLM的聊天消息格式以及存入 Redis 的历史记录格式
type LLMMessage struct {
	Role       string            `json:"role"`
	Content    string            `json:"content"`
	ToolCalls  []openai.ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string            `json:"tool_call_id,omitempty"`
	Name       string            `json:"name,omitempty"`
}

// --- 下列为 Function Calling 专用的参数结构 ---

type GetUserParams struct {
	Username string `json:"username"`
}

type CreateUserParams struct {
	Username   string  `json:"username"`
	ExpiryDate *string `json:"expiry_date"`
	Quota      *int    `json:"quota"`
}

type UpdateQuotaParams struct {
	Username string `json:"username"`
	Quota    int    `json:"quota"`
}

type UpdateExpiryParams struct {
	Username   string `json:"username"`
	ExpiryDate string `json:"expiry_date"`
}

type DeleteUserParams struct {
	Username string `json:"username"`
}
