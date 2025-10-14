package local

import (
	"context"

	"github.com/andranikuz/aiwf/providers/openai"
	"github.com/andranikuz/aiwf/runtime/go/aiwf"
)

// ClientConfig описывает параметры локального OpenAI-совместимого эндпоинта.
type ClientConfig struct {
	Endpoint string
	APIKey   string
}

// Client оборачивает openai.Client для повторного использования логики JSON Schema.
type Client struct {
	upstream *openai.Client
}

// NewClient создаёт обёртку над OpenAI-клиентом с пользовательским endpoint.
func NewClient(cfg ClientConfig) (*Client, error) {
	upstream, err := openai.NewClient(openai.ClientConfig{
		BaseURL: cfg.Endpoint,
		APIKey:  fallbackKey(cfg.APIKey),
	})
	if err != nil {
		return nil, err
	}
	return &Client{upstream: upstream}, nil
}

func fallbackKey(key string) string {
	if key == "" {
		return "local-override"
	}
	return key
}

// CallJSONSchema делегирует вызов локальному OpenAI-совместимому серверу.
func (c *Client) CallJSONSchema(ctx context.Context, call aiwf.ModelCall) ([]byte, aiwf.Tokens, error) {
	return c.upstream.CallJSONSchema(ctx, call)
}

// CallJSONSchemaStream делегирует потоковый вызов (пока тоже заглушка).
func (c *Client) CallJSONSchemaStream(ctx context.Context, call aiwf.ModelCall) (<-chan aiwf.StreamChunk, aiwf.Tokens, error) {
	return c.upstream.CallJSONSchemaStream(ctx, call)
}
