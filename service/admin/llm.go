package admin

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/twbworld/proxy/global"
	modelCommon "github.com/twbworld/proxy/model/common"
	"github.com/twbworld/proxy/model/config"
	"github.com/twbworld/proxy/service/common"
)

// LLMService LLM 服务接口
type LLMService interface {
	Handle(ctx context.Context, chatID int64, userMessage string) (string, error)
}

type llmService struct {
	async            common.AsyncService
	notifier         global.NotifierService
	config           *config.LLM
	tools            []modelCommon.ToolDefinition
	systemPromptFunc func() string // 懒加载 system prompt
}

// System prompt 的懒加载缓存
var getSystemPrompt = sync.OnceValue(func() string {
	return strings.TrimSpace(`
你是一位专业的 Telegram Bot 智能助理，负责帮助用户管理代理用户数据库。
你的任务是通过自然语言理解用户的意图，并调用相应的工具来完成操作。

可用操作包括：
1. 查询用户信息 - 获取指定用户的流量配额、已用流量、到期时间等详情
2. 新增用户 - 创建新用户。只需提供用户名即可创建，到期时间和流量配额均为可选参数：若用户未指定到期时间或配额，切勿追问用户，直接调用 create_user 工具并将其设为 null
3. 更新流量 - 修改用户的流量配额（-1 表示不限）
4. 更新到期时间 - 修改用户的到期日期
5. 删除用户 - 移除指定用户
6. 列出所有用户 - 显示所有用户的简要列表

回答规则：
- 使用友好、专业的中文回复
- 操作成功后给出明确的成功提示和结果摘要
- 操作失败时清晰说明原因并提供可能的解决方案
- 对于不确定的信息，先查询再回答，严禁编造数据
- 保持回复简洁明了，避免冗长

请根据用户的具体请求，选择合适的工具调用。
`)
})

func NewLLMService(async common.AsyncService, notifier global.NotifierService) LLMService {
	cfg := global.Config().LLM
	return &llmService{
		async:            async,
		notifier:         notifier,
		config:           &cfg,
		tools:            getToolDefinitions(),
		systemPromptFunc: getSystemPrompt,
	}
}

// Handle 处理用户的自然语言消息
func (s *llmService) Handle(ctx context.Context, chatID int64, userMessage string) (string, error) {
	if !s.config.Enabled || global.RedisClient == nil {
		return "智能助理功能未启用，请使用 /start 快捷菜单", nil
	}

	// 验证 Redis 连接
	pingCtx, pingCancel := context.WithTimeout(ctx, 2*time.Second)
	defer pingCancel()
	if err := global.RedisClient.Ping(pingCtx).Err(); err != nil {
		global.Log.Warnf("Redis 不可用，LLM 服务降级：%v", err)
		return "智能助理当前不可用，请使用 /start 快捷菜单", nil
	}

	// 构建对话历史 key
	historyKey := fmt.Sprintf("llm:history:%d", chatID)

	// 获取或初始化对话上下文
	messages, err := s.loadHistory(ctx, historyKey)
	if err != nil {
		global.Log.Errorf("加载对话历史失败：%v", err)
		return "系统繁忙，请稍后重试", nil
	}

	// 添加用户消息到历史
	messages = append(messages, modelCommon.ChatMessage{
		Role:    "user",
		Content: userMessage,
	})

	// 调用 LLM API
	resp, err := s.callLLM(ctx, messages)
	if err != nil {
		global.Log.Errorf("LLM API 调用失败：%v", err)
		return "智能助理响应超时，请稍后重试", nil
	}

	// 检查是否有错误
	if resp.Error != nil {
		global.Log.Errorf("LLM API 返回错误：%s", resp.Error.Message)
		return "智能助理遇到错误，请稍后重试", nil
	}

	// 获取助手回复
	if len(resp.Choices) == 0 {
		return "未收到有效响应，请稍后重试", nil
	}

	assistantMsg := resp.Choices[0].Message

	// 如果有 tool_calls，执行工具并回传结果
	if len(assistantMsg.ToolCalls) > 0 {
		resultMsg, err := s.executeToolsAndRetry(ctx, historyKey, messages, assistantMsg)
		if err != nil {
			global.Log.Errorf("工具执行失败：%v", err)
			return "执行操作时遇到问题，请稍后重试", nil
		}
		return resultMsg, nil
	}

	// 直接返回助手的文本回复
	finalReply := assistantMsg.Content
	if finalReply == "" {
		return "抱歉，我无法理解您的请求，请尝试用更明确的描述", nil
	}

	// 保存更新后的对话历史
	if err := s.saveHistory(ctx, historyKey, append(messages, assistantMsg)); err != nil {
		global.Log.Warnf("保存对话历史失败：%v", err)
	}

	return finalReply, nil
}

// loadHistory 从 Redis 加载对话历史
func (s *llmService) loadHistory(ctx context.Context, key string) ([]modelCommon.ChatMessage, error) {
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
		// 首次对话，只包含 system prompt
		return []modelCommon.ChatMessage{
			{Role: "system", Content: s.systemPromptFunc()},
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("获取历史失败：%w", err)
	}

	var messages []modelCommon.ChatMessage
	if err := json.Unmarshal([]byte(data), &messages); err != nil {
		// 解析失败，重置为初始状态
		return []modelCommon.ChatMessage{
			{Role: "system", Content: s.systemPromptFunc()},
		}, nil
	}

	return messages, nil
}

// saveHistory 保存到 Redis
func (s *llmService) saveHistory(ctx context.Context, key string, messages []modelCommon.ChatMessage) error {
	data, err := json.Marshal(messages)
	if err != nil {
		return fmt.Errorf("序列化历史失败：%w", err)
	}

	ttl := time.Duration(global.Config().Redis.ConversationHistoryTTL) * time.Second
	return global.RedisClient.Set(ctx, key, data, ttl).Err()
}

// callLLM 调用 LLM API
func (s *llmService) callLLM(ctx context.Context, messages []modelCommon.ChatMessage) (*modelCommon.ChatCompletionResponse, error) {
	reqBody := modelCommon.ChatCompletionRequest{
		Model:      s.config.Model,
		Messages:   messages,
		Tools:      s.tools,
		MaxTokens:  s.config.MaxTokens,
		TimeoutSec: s.config.TimeoutSec,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("请求序列化失败：%w", err)
	}

	// 构建 HTTP 请求
	apiURL := strings.TrimSuffix(s.config.BaseURL, "/") + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败：%w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.config.APIKey)

	// 发送请求
	client := &http.Client{
		Timeout: time.Duration(s.config.TimeoutSec) * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP 请求失败：%w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败：%w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API 返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	var completionResp modelCommon.ChatCompletionResponse
	if err := json.Unmarshal(body, &completionResp); err != nil {
		return nil, fmt.Errorf("解析响应失败：%w", err)
	}

	return &completionResp, nil
}

// executeToolsAndRetry 执行工具调用并回传结果给 LLM
func (s *llmService) executeToolsAndRetry(ctx context.Context, historyKey string, originalMessages []modelCommon.ChatMessage, assistantMsg modelCommon.ChatMessage) (string, error) {
	// 添加助手消息到历史
	messages := append(originalMessages, assistantMsg)

	// 执行每个工具调用
	for _, toolCall := range assistantMsg.ToolCalls {
		result, err := executeTool(ctx, toolCall.Function.Name, toolCall.Function.Arguments)
		global.Log.Debug("=====LLM调用工具: ", toolCall.Function.Name, "|参数: ", toolCall.Function.Arguments)

		toolResultMsg := modelCommon.ChatMessage{
			Role:       "tool",
			Content:    result,
			ToolCallID: toolCall.ID,
			Name:       toolCall.Function.Name,
		}

		if err != nil {
			toolResultMsg.Content = fmt.Sprintf("执行出错：%v", err)
			global.Log.Errorf("工具 `%s` 执行失败：%v", toolCall.Function.Name, err)
		}

		messages = append(messages, toolResultMsg)
	}

	// 将工具结果回传给 LLM，获取最终回复
	finalResp, err := s.callLLM(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("二次调用 LLM 失败：%w", err)
	}

	if len(finalResp.Choices) == 0 || finalResp.Choices[0].Message.Content == "" {
		return "操作已完成，但无法生成友好的回复", nil
	}

	finalReply := finalResp.Choices[0].Message.Content

	global.Log.Debug("=====LLM最终回复: ", finalReply)

	// 保存完整对话历史
	allMessages := append(messages, finalResp.Choices[0].Message)
	if err := s.saveHistory(ctx, historyKey, allMessages); err != nil {
		global.Log.Warnf("保存对话历史失败：%v", err)
	}

	return finalReply, nil
}
