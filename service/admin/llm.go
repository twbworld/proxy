package admin

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sashabaranov/go-openai"
	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/internal/llm"
	"github.com/twbworld/proxy/model/common"
	"github.com/twbworld/proxy/model/enum"
)

type LLMService interface {
	Handle(ctx context.Context, chatID int64, userMessage string) (string, error)
}

type llmService struct {
	tools []openai.Tool
}

func NewLLMService() LLMService {
	return &llmService{
		tools: getToolDefinitions(),
	}
}

func (s *llmService) Handle(ctx context.Context, chatID int64, userMessage string) (string, error) {
	if global.LlmService == nil || global.RedisClient == nil {
		return "智能助理功能未启用，请使用 /start 快捷菜单", nil
	}

	pingCtx, pingCancel := context.WithTimeout(ctx, 2*time.Second)
	defer pingCancel()
	if err := global.RedisClient.Ping(pingCtx).Err(); err != nil {
		global.Log.Warnf("Redis不可用，LLM服务降级：%v", err)
		return "智能助理当前不可用，请稍后重试", nil
	}

	historyKey := fmt.Sprintf("llm:history:%d", chatID)
	messages, err := s.loadHistory(ctx, historyKey)
	if err != nil {
		global.Log.Errorf("加载对话历史失败：%v", err)
		return "系统繁忙，请稍后重试", nil
	}

	messages = append(messages, common.LLMMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: userMessage,
	})

	req := &llm.ChatRequest{
		Size:            enum.ModelSmall,
		SystemPrompt:    enum.SystemPromptChat,
		Messages:        messages,
		Tools:           s.tools,
		DisableThinking: true,
	}

	// 利用闭包注入业务逻辑
	newMessages, err := global.LlmService.ChatAuto(ctx, req, func(c context.Context, name, args string) (string, error) {
		global.Log.Debug("=====LLM调用工具: ", name, "|参数: ", args)
		return executeTool(c, name, args)
	})
	if err != nil {
		global.Log.Errorf("LLM自动执行链路异常：%v", err)
		return "智能助理处理遇到问题，请稍后重试", nil
	}

	_ = s.saveHistory(ctx, historyKey, newMessages)

	lastMsg := newMessages[len(newMessages)-1]
	if lastMsg.Role == openai.ChatMessageRoleAssistant && lastMsg.Content != "" {
		return lastMsg.Content, nil
	}
	return "抱歉，我无法理解您的请求", nil
}

// loadHistory 从 Redis 加载对话历史
func (s *llmService) loadHistory(ctx context.Context, key string) ([]common.LLMMessage, error) {
	// 使用分布式锁防止并发写入
	lockKey := key + ":lock"
	acquired, err := global.RedisClient.SetNX(ctx, lockKey, "1", time.Duration(global.Config().Redis.HistoryLockExpiry)*time.Second).Result()
	if err != nil {
		return nil, fmt.Errorf("获取锁失败：%w", err)
	}
	if !acquired {
		return nil, errors.New("获取会话锁超时")
	}
	defer global.RedisClient.Del(ctx, lockKey)

	data, err := global.RedisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return []common.LLMMessage{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("获取历史失败：%w", err)
	}

	var messages []common.LLMMessage
	if err := json.Unmarshal([]byte(data), &messages); err != nil {
		return []common.LLMMessage{}, nil
	}

	// 限制最多保留最近 5 轮完整对话
	const maxTurns = 3
	return trimHistory(messages, maxTurns), nil
}

// trimHistory 倒序寻找第 maxTurns 个 user 消息边界，完整保留最近的上下文轮次
func trimHistory(messages []common.LLMMessage, maxTurns int) []common.LLMMessage {
	if maxTurns <= 0 || len(messages) == 0 {
		return messages
	}

	userCount := 0
	cutIdx := 0

	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == openai.ChatMessageRoleUser {
			userCount++
			if userCount == maxTurns {
				cutIdx = i
				break
			}
		}
	}

	if userCount >= maxTurns {
		// 复制切片，避免底层大数组引用驻留内存
		return slices.Clone(messages[cutIdx:])
	}

	return messages
}

// saveHistory 保存到 Redis
func (s *llmService) saveHistory(ctx context.Context, key string, messages []common.LLMMessage) error {
	data, err := json.Marshal(messages)
	if err != nil {
		return err
	}
	ttl := time.Duration(global.Config().Redis.ConversationHistoryTTL) * time.Second
	return global.RedisClient.Set(ctx, key, data, ttl).Err()
}
