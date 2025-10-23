package anthropic

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
	defaultBaseURL   = "https://api.anthropic.com/v1"
	anthropicVersion = "2024-10-01"
)

// ClientConfig определяет параметры доступа к Anthropic API.
type ClientConfig struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
	Timeout    time.Duration
}

// Client реализует aiwf.ModelClient для Anthropic (Claude).
type Client struct {
	baseURL   string
	apiKey    string
	http      *http.Client
	threadMgr aiwf.ThreadManager
}

// NewClient создаёт клиента для Anthropic.
func NewClient(cfg ClientConfig) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("anthropic: api key is required")
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

// CallJSONSchema выполняет запрос к Anthropic API.
func (c *Client) CallJSONSchema(ctx context.Context, call aiwf.ModelCall) ([]byte, aiwf.Tokens, error) {
	req, err := c.newMessageRequest(ctx, call)
	if err != nil {
		return nil, aiwf.Tokens{}, fmt.Errorf("anthropic: failed to create request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, aiwf.Tokens{}, fmt.Errorf("anthropic: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		buf, _ := io.ReadAll(resp.Body)
		return nil, aiwf.Tokens{}, fmt.Errorf("anthropic: unexpected status %d: %s", resp.StatusCode, string(buf))
	}

	var parsed Message
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, aiwf.Tokens{}, fmt.Errorf("anthropic: failed to decode response: %w", err)
	}

	if len(parsed.Content) == 0 {
		return nil, aiwf.Tokens{}, errors.New("anthropic: empty response content")
	}

	content := ""
	for _, block := range parsed.Content {
		if block.Type == "text" {
			content += block.Text
		}
	}

	usage := aiwf.Tokens{
		Prompt:     parsed.Usage.InputTokens,
		Completion: parsed.Usage.OutputTokens,
		Total:      parsed.Usage.InputTokens + parsed.Usage.OutputTokens,
	}

	return []byte(content), usage, nil
}

// CallJSONSchemaStream не реализовано для Anthropic
func (c *Client) CallJSONSchemaStream(ctx context.Context, call aiwf.ModelCall) (<-chan aiwf.StreamChunk, aiwf.Tokens, error) {
	return nil, aiwf.Tokens{}, errors.New("anthropic: streaming not implemented")
}

// newMessageRequest создаёт HTTP запрос для Messages API.
func (c *Client) newMessageRequest(ctx context.Context, call aiwf.ModelCall) (*http.Request, error) {
	// Используем Payload как входные данные (уже типизированные)
	userContent := call.UserPrompt
	if call.Payload != nil {
		inputJSON, err := json.Marshal(call.Payload)
		if err != nil {
			return nil, err
		}
		userContent = string(inputJSON)
	}

	maxTokens := call.MaxTokens
	if maxTokens == 0 {
		maxTokens = 2000
	}

	temperature := call.Temperature
	if temperature == 0 {
		temperature = 0.7
	}

	payload := MessageRequest{
		Model: call.Model,
		Messages: []MessageParam{
			{
				Role:    "user",
				Content: userContent,
			},
		},
		System:      call.SystemPrompt,
		MaxTokens:   maxTokens,
		Temperature: temperature,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := c.baseURL + "/messages"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", anthropicVersion)
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// MessageRequest - структура запроса к Anthropic Messages API
type MessageRequest struct {
	Model       string         `json:"model"`
	Messages    []MessageParam `json:"messages"`
	System      string         `json:"system,omitempty"`
	MaxTokens   int            `json:"max_tokens"`
	Temperature float64        `json:"temperature,omitempty"`
	TopK        int            `json:"top_k,omitempty"`
	TopP        float64        `json:"top_p,omitempty"`
}

// MessageParam - параметр сообщения в запросе
type MessageParam struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Message - структура ответа от Anthropic API
type Message struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Role         string         `json:"role"`
	Content      []ContentBlock `json:"content"`
	Model        string         `json:"model"`
	StopReason   string         `json:"stop_reason"`
	StopSequence string         `json:"stop_sequence"`
	Usage        Usage          `json:"usage"`
}

// ContentBlock - блок контента в ответе
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// Usage - информация об использованных токенах
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}
