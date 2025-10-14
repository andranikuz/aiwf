package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/andranikuz/aiwf/providers/internal/retry"
	"github.com/andranikuz/aiwf/runtime/go/aiwf"
)

const (
	defaultBaseURL = "https://api.anthropic.com/v1"
	messagesPath   = "/messages"
)

// ClientConfig задаёт параметры клиента Anthropic.
type ClientConfig struct {
	BaseURL    string
	APIKey     string
	Version    string
	HTTPClient *http.Client
	Retry      retry.Strategy
}

// Client реализует aiwf.ModelClient для Anthropic Messages API.
type Client struct {
	baseURL string
	apiKey  string
	version string
	http    *http.Client
	retry   retry.Strategy
}

// NewClient создаёт клиента и валидирует обязательные поля.
func NewClient(cfg ClientConfig) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("anthropic: api key is required")
	}

	base := cfg.BaseURL
	if base == "" {
		base = defaultBaseURL
	}

	version := cfg.Version
	if version == "" {
		version = "2023-06-01"
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	strat := cfg.Retry
	if strat.MaxAttempts == 0 {
		strat = retry.DefaultStrategy()
	}

	return &Client{
		baseURL: base,
		apiKey:  cfg.APIKey,
		version: version,
		http:    httpClient,
		retry:   strat,
	}, nil
}

// CallJSONSchema выполняет запрос Messages API c JSON Schema.
func (c *Client) CallJSONSchema(ctx context.Context, call aiwf.ModelCall) ([]byte, aiwf.Tokens, error) {
	if call.OutputSchemaRef == "" {
		return nil, aiwf.Tokens{}, errors.New("anthropic: output schema ref is required")
	}

	messages := make([]message, 0, 2)
	if call.SystemPrompt != "" {
		messages = append(messages, message{
			Role:    "system",
			Content: []contentBlock{{Type: "text", Text: call.SystemPrompt}},
		})
	}
	messages = append(messages, message{
		Role:    "user",
		Content: []contentBlock{{Type: "text", Text: call.UserPrompt}},
	})

	payload := requestPayload{
		Model:           call.Model,
		Messages:        messages,
		MaxOutputTokens: call.MaxTokens,
		Temperature:     call.Temperature,
		ResponseFormat: responseFormat{
			Type: "json_schema",
			Schema: schemaDescriptor{
				Name: call.OutputSchemaRef,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, aiwf.Tokens{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+messagesPath, bytes.NewReader(body))
	if err != nil {
		return nil, aiwf.Tokens{}, err
	}

	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", c.version)
	req.Header.Set("content-type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, aiwf.Tokens{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, aiwf.Tokens{}, fmt.Errorf("anthropic: status %d", resp.StatusCode)
	}

	var parsed responsePayload
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, aiwf.Tokens{}, err
	}

	if len(parsed.Content) == 0 {
		return nil, aiwf.Tokens{}, errors.New("anthropic: empty content")
	}

	first := parsed.Content[0]
	if first.Type != "tool_result" || first.ToolResult == nil {
		return nil, aiwf.Tokens{}, errors.New("anthropic: missing json tool result")
	}

	if err := validateSchema(call.OutputSchemaRef, first.ToolResult); err != nil {
		return nil, aiwf.Tokens{}, err
	}

	raw, err := json.Marshal(first.ToolResult)
	if err != nil {
		return nil, aiwf.Tokens{}, err
	}

	usage := aiwf.Tokens{
		Prompt:     parsed.Usage.InputTokens,
		Completion: parsed.Usage.OutputTokens,
		Total:      parsed.Usage.TotalTokens,
	}

	return raw, usage, nil
}

// CallJSONSchemaStream пока возвращает заглушку.
func (c *Client) CallJSONSchemaStream(ctx context.Context, call aiwf.ModelCall) (<-chan aiwf.StreamChunk, aiwf.Tokens, error) {
	ch := make(chan aiwf.StreamChunk)
	close(ch)
	return ch, aiwf.Tokens{}, errors.New("anthropic: streaming not implemented")
}

func validateSchema(schemaRef string, data any) error {
	if schemaRef == "" {
		return errors.New("anthropic: schema reference is required")
	}
	return nil
}

type requestPayload struct {
	Model           string         `json:"model"`
	Messages        []message      `json:"messages"`
	MaxOutputTokens int            `json:"max_output_tokens,omitempty"`
	Temperature     float64        `json:"temperature,omitempty"`
	ResponseFormat  responseFormat `json:"response_format"`
}

type message struct {
	Role    string         `json:"role"`
	Content []contentBlock `json:"content"`
}

type contentBlock struct {
	Type       string `json:"type"`
	Text       string `json:"text,omitempty"`
	ToolResult any    `json:"tool_result,omitempty"`
}

type responseFormat struct {
	Type   string           `json:"type"`
	Schema schemaDescriptor `json:"json_schema"`
}

type schemaDescriptor struct {
	Name string `json:"name"`
}

type responsePayload struct {
	Content []contentBlock `json:"content"`
	Usage   usagePayload   `json:"usage"`
}

type usagePayload struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}
