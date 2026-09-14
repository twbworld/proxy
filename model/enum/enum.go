package enum

type DbType string

const (
	MYSQL  DbType = `mysql`
	SQLITE DbType = `sqlite3`
)

type Msg string

const (
	DefaultSuccessMsg Msg = `ok`
	DefaultFailMsg    Msg = `错误`
)

type ResCode int8

const (
	SuccessCode   ResCode = 0
	ErrorCode     ResCode = 1
	AuthErrorCode ResCode = 2
)

type LLMSize string

const (
	ModelSmall  LLMSize = "small"
	ModelMedium LLMSize = "medium"
	ModelLarge  LLMSize = "large"
)

type SystemPrompt string

const (
	SystemPromptChat SystemPrompt = `你是一位专业的Telegram Bot智能助理，负责代理用户数据库管理。根据用户意图调用工具完成操作。
- 无法确定时先调用工具查询，严禁凭空编造。
- 结果排版用单星号*加粗*，反引号包裹代码，保证Markdown符号成对闭合。`
)

type SchemaName string
