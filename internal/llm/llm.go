package llm

import (
	"cmp"
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
	"github.com/twbworld/proxy/model/common"
	"github.com/twbworld/proxy/model/config"
	"github.com/twbworld/proxy/model/enum"
)

type CtxKey string

const DisableThinkingKey CtxKey = "disable_thinking"

type ChatRequest struct {
	Size            enum.LLMSize
	SystemPrompt    enum.SystemPrompt
	Messages        []common.LLMMessage
	Tools           []openai.Tool
	SchemaName      enum.SchemaName
	Schema          any
	Strict          bool
	Temperature     *float32
	DisableThinking bool
}

type Service interface {
	Chat(ctx context.Context, req *ChatRequest) (*openai.ChatCompletionMessage, error)
	ChatAuto(ctx context.Context, req *ChatRequest, toolHandler func(context.Context, string, string) (string, error)) ([]common.LLMMessage, error)
}

type client struct {
	llmClients map[enum.LLMSize]*openai.Client
	llmConfigs []config.LLM
}

func NewClient(log *logrus.Logger, clients map[enum.LLMSize]*openai.Client, configs []config.LLM) Service {
	return &client{
		llmClients: clients,
		llmConfigs: configs,
	}
}

func (c *client) Chat(ctx context.Context, req *ChatRequest) (*openai.ChatCompletionMessage, error) {
	var llmConfig *config.LLM
	for i := range c.llmConfigs {
		if enum.LLMSize(c.llmConfigs[i].Size) == req.Size {
			llmConfig = &c.llmConfigs[i]
			break
		}
	}
	if llmConfig == nil && len(c.llmConfigs) > 0 {
		llmConfig = &c.llmConfigs[0]
	}
	if llmConfig == nil || llmConfig.Model == "" {
		return nil, errors.New("未找到LLM配置")
	}

	if req.DisableThinking {
		ctx = context.WithValue(ctx, DisableThinkingKey, true)
	}

	messages := make([]openai.ChatCompletionMessage, 0, len(req.Messages)+1)
	if req.SystemPrompt != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: string(req.SystemPrompt),
		})
	}
	for _, msg := range req.Messages {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:       msg.Role,
			Content:    msg.Content,
			ToolCalls:  msg.ToolCalls,
			ToolCallID: msg.ToolCallID,
			Name:       msg.Name,
		})
	}

	openaiReq := openai.ChatCompletionRequest{
		Model:    llmConfig.Model,
		Messages: messages,
		Tools:    req.Tools,
	}

	if temp := cmp.Or(req.Temperature, llmConfig.Temperature); temp != nil && *temp != 0 {
		openaiReq.Temperature = *temp
	}

	if req.SchemaName != "" && req.Schema != nil {
		var schemaMarshaler json.Marshaler
		switch s := req.Schema.(type) {
		case json.Marshaler:
			schemaMarshaler = s
		case []byte:
			schemaMarshaler = jsontext.Value(s)
		case string:
			schemaMarshaler = jsontext.Value(s)
		default:
			b, err := json.Marshal(req.Schema)
			if err != nil {
				return nil, fmt.Errorf("序列化 JSON Schema 失败: %w", err)
			}
			schemaMarshaler = jsontext.Value(b)
		}

		openaiReq.ResponseFormat = &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				Name:   string(req.SchemaName),
				Schema: schemaMarshaler,
				Strict: req.Strict,
			},
		}
	}

	llmClient, ok := c.llmClients[req.Size]
	if !ok {
		return nil, errors.New("未找到指定大小的LLM实例")
	}

	// aa, _ := json.Marshal(openaiReq)
	// fmt.Printf("==========请求LLM: %s", string(aa))

	resp, err := llmClient.CreateChatCompletion(ctx, openaiReq)
	if err != nil {
		return nil, err
	}
	if len(resp.Choices) == 0 {
		return nil, errors.New("LLM返回空结果")
	}

	// bb, _ := json.Marshal(resp.Choices)
	// fmt.Printf("==========LLM返回: %s", string(bb))

	msg := resp.Choices[0].Message
	if msg.Content != "" {
		msg.Content = filterContent(msg.Content)
	}

	return &msg, nil
}

func (c *client) ChatAuto(ctx context.Context, req *ChatRequest, toolHandler func(context.Context, string, string) (string, error)) ([]common.LLMMessage, error) {
	for range 5 {
		msg, err := c.Chat(ctx, req)
		if err != nil {
			return req.Messages, err
		}

		if len(msg.ToolCalls) == 0 {
			if msg.Content != "" {
				req.Messages = append(req.Messages, common.LLMMessage{
					Role:    openai.ChatMessageRoleAssistant,
					Content: msg.Content,
				})
			}
			return req.Messages, nil
		}

		req.Messages = append(req.Messages, common.LLMMessage{
			Role:      openai.ChatMessageRoleAssistant,
			Content:   msg.Content,
			ToolCalls: msg.ToolCalls,
		})

		for _, toolCall := range msg.ToolCalls {
			result, err := toolHandler(ctx, toolCall.Function.Name, toolCall.Function.Arguments)
			toolMsg := common.LLMMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    result,
				ToolCallID: toolCall.ID,
				Name:       toolCall.Function.Name,
			}
			if err != nil {
				toolMsg.Content = fmt.Sprintf("执行出错：%v", err)
			}
			req.Messages = append(req.Messages, toolMsg)
		}
	}

	return req.Messages, errors.New("超出LLM最大调用次数")
}

func filterContent(rawAnswer string) string {
	startTag, endTag := "<think>", "</think>"
	sIdx := strings.Index(rawAnswer, startTag)
	eIdx := strings.Index(rawAnswer, endTag)

	if sIdx != -1 && eIdx != -1 && eIdx > sIdx {
		return strings.TrimSpace(rawAnswer[:sIdx] + rawAnswer[eIdx+len(endTag):])
	}
	if parts := strings.SplitN(rawAnswer, startTag, 2); len(parts) > 1 {
		return strings.TrimSpace(parts[1])
	}
	return strings.TrimSpace(rawAnswer)
}
