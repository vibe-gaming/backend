package gigachat

import (
	"encoding/json"
)

// TokenResponse представляет ответ с токеном
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresAt   int64  `json:"expires_at"`
}

// ErrorResponse представляет ответ с ошибкой
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Message представляет сообщение в чате
type Message struct {
	Role         string        `json:"role"`
	Content      string        `json:"content"`
	Attachments  []string      `json:"attachments,omitempty"`
	FunctionCall *FunctionCall `json:"function_call,omitempty"`
}

// FunctionCall представляет вызов функции
type FunctionCall struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// ChatRequest представляет запрос к чату
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// ChatResponse представляет ответ от чата
type ChatResponse struct {
	Choices []Choice `json:"choices"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Usage   Usage    `json:"usage"`
	Object  string   `json:"object"`
}

// Choice представляет один из вариантов ответа
type Choice struct {
	Message      Message `json:"message"`
	Index        int     `json:"index"`
	FinishReason string  `json:"finish_reason"`
}

// Usage представляет статистику использования токенов
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Scope константы для различных типов доступа
const (
	ScopePersonal = "GIGACHAT_API_PERS" // для физических лиц
	ScopeB2B      = "GIGACHAT_API_B2B"  // для ИП и юридических лиц (платные пакеты)
	ScopeCorp     = "GIGACHAT_API_CORP" // для ИП и юридических лиц (pay-as-you-go)
)

// Model константы для доступных моделей
const (
	ModelGigaChat     = "GigaChat"
	ModelGigaChatPro  = "GigaChat-Pro"
	ModelGigaChatMax  = "GigaChat-Max"
	ModelGigaChat2    = "GigaChat-2"
	ModelGigaChatPro2 = "GigaChat-2-Pro"
	ModelGigaChatMax2 = "GigaChat-2-Max"
)

// Role константы для ролей сообщений
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleFunction  = "function"
)
