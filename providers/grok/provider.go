package grok

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/andranikuz/aiwf/runtime/go/aiwf"
)

const (
	defaultBaseURL = "https://api.x.ai/v1"
)

// ClientConfig определяет параметры доступа к Grok API (xAI).
type ClientConfig struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
	Timeout    time.Duration
}

// Client реализует aiwf.ModelClient для Grok (xAI).
type Client struct {
	baseURL   string
	apiKey    string
	http      *http.Client
	threadMgr aiwf.ThreadManager
}

// NewClient создаёт клиента для Grok.
func NewClient(cfg ClientConfig) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("grok: api key is required")
	}

	base := cfg.BaseURL
	if base == "" {
		base = defaultBaseURL
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 60 * time.Second
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeout}
	} else if httpClient.Timeout == 0 {
		httpClient.Timeout = timeout
	}

	return &Client{
		baseURL:   base,
		apiKey:    cfg.APIKey,
		http:      httpClient,
		threadMgr: nil,
	}, nil
}

// WithThreadManager устанавливает менеджер тредов
func (c *Client) WithThreadManager(tm aiwf.ThreadManager) *Client {
	c.threadMgr = tm
	return c
}

// CallJSONSchema выполняет запрос к Grok API.
func (c *Client) CallJSONSchema(ctx context.Context, call aiwf.ModelCall) ([]byte, aiwf.Tokens, error) {
	req, err := c.newChatRequest(ctx, call)
	if err != nil {
		return nil, aiwf.Tokens{}, fmt.Errorf("grok: failed to create request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, aiwf.Tokens{}, fmt.Errorf("grok: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		buf, _ := io.ReadAll(resp.Body)
		return nil, aiwf.Tokens{}, fmt.Errorf("grok: unexpected status %d: %s", resp.StatusCode, string(buf))
	}

	var parsed ChatCompletion
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, aiwf.Tokens{}, fmt.Errorf("grok: failed to decode response: %w", err)
	}

	if len(parsed.Choices) == 0 {
		return nil, aiwf.Tokens{}, errors.New("grok: empty response choices")
	}

	content := parsed.Choices[0].Message.Content
	usage := aiwf.Tokens{
		Prompt:     parsed.Usage.PromptTokens,
		Completion: parsed.Usage.CompletionTokens,
		Total:      parsed.Usage.TotalTokens,
	}

	return []byte(content), usage, nil
}

// CallJSONSchemaStream не реализовано для Grok
func (c *Client) CallJSONSchemaStream(ctx context.Context, call aiwf.ModelCall) (<-chan aiwf.StreamChunk, aiwf.Tokens, error) {
	return nil, aiwf.Tokens{}, errors.New("grok: streaming not implemented")
}

// newChatRequest создаёт HTTP запрос для Chat API.
func (c *Client) newChatRequest(ctx context.Context, call aiwf.ModelCall) (*http.Request, error) {
	messages := []Message{
		{
			Role:    "system",
			Content: call.SystemPrompt,
		},
	}

	// Используем Payload как входные данные (уже типизированные)
	userMessage := call.UserPrompt
	if call.Payload != nil {
		inputJSON, err := json.Marshal(call.Payload)
		if err != nil {
			return nil, err
		}
		userMessage = string(inputJSON)
	}

	messages = append(messages, Message{
		Role:    "user",
		Content: userMessage,
	})

	payload := ChatRequest{
		Model:       call.Model,
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   2000,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := c.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// Message представляет сообщение в диалоге
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest - структура запроса к Grok Chat API
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`
}

// ChatCompletion - структура ответа от Grok API
type ChatCompletion struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice - выбор из ответа
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage - информация об использованных токенах
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}