package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/andranikuz/aiwf/runtime/go/aiwf"
)

const (
	defaultBaseURL = "https://api.openai.com/v1"
	responsesPath  = "/responses"
)

// ClientConfig определяет параметры доступа к OpenAI.
type ClientConfig struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
	Timeout    time.Duration
}

// Client реализует aiwf.ModelClient для OpenAI Responses API.
type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

// NewClient создаёт клиента с базовыми значениями.
func NewClient(cfg ClientConfig) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("openai: api key is required")
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
		baseURL: base,
		apiKey:  cfg.APIKey,
		http:    httpClient,
	}, nil
}

// CallJSONSchema выполняет синхронный запрос к Responses API.
func (c *Client) CallJSONSchema(ctx context.Context, call aiwf.ModelCall) ([]byte, aiwf.Tokens, error) {
	req, err := c.newRequest(ctx, call)
	if err != nil {
		return nil, aiwf.Tokens{}, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, aiwf.Tokens{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		buf, _ := io.ReadAll(resp.Body)
		log.Printf("openai: status %d body=%s", resp.StatusCode, string(buf))
		return nil, aiwf.Tokens{}, fmt.Errorf("openai: unexpected status %d", resp.StatusCode)
	}

	var parsed responsePayload
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, aiwf.Tokens{}, err
	}

	structuredText, err := extractStructuredText(parsed.Output)
	if err != nil {
		return nil, aiwf.Tokens{}, err
	}
	log.Printf("openai: output json=%s", structuredText)

	var payload any
	if err := json.Unmarshal([]byte(structuredText), &payload); err != nil {
		return nil, aiwf.Tokens{}, fmt.Errorf("openai: parse structured json: %w", err)
	}

	if err := validateSchema(call.OutputSchemaRef, payload); err != nil {
		return nil, aiwf.Tokens{}, err
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, aiwf.Tokens{}, err
	}

	usage := aiwf.Tokens{
		Prompt:     parsed.Usage.PromptTokens,
		Completion: parsed.Usage.CompletionTokens,
		Total:      parsed.Usage.TotalTokens,
	}
	if parsed.Usage.InputTokens != 0 || parsed.Usage.OutputTokens != 0 {
		usage.Prompt = parsed.Usage.InputTokens
		usage.Completion = parsed.Usage.OutputTokens
		if usage.Total == 0 {
			usage.Total = usage.Prompt + usage.Completion
		}
	}

	return raw, usage, nil
}

// CallJSONSchemaStream возвращает заглушку каналов для будущей поддержки стриминга.
func (c *Client) CallJSONSchemaStream(ctx context.Context, call aiwf.ModelCall) (<-chan aiwf.StreamChunk, aiwf.Tokens, error) {
	ch := make(chan aiwf.StreamChunk)
	close(ch)
	return ch, aiwf.Tokens{}, errors.New("openai: streaming not implemented yet")
}

func (c *Client) newRequest(ctx context.Context, call aiwf.ModelCall) (*http.Request, error) {
	inputMessages, err := buildInputMessages(call)
	if err != nil {
		return nil, err
	}

	format, err := buildJSONSchemaFormat(call)
	if err != nil {
		return nil, err
	}

	payload := requestPayload{
		Model:           call.Model,
		Input:           inputMessages,
		MaxOutputTokens: call.MaxTokens,
		Temperature:     call.Temperature,
		Text:            format,
	}
	if meta := buildMetadata(call); len(meta) > 0 {
		payload.Metadata = meta
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	log.Printf("openai: req=%s", string(body))

	url := c.baseURL + responsesPath
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// validateSchema пока что выполняет заглушку, заменяем настоящей валидацией после интеграции с runtime validate.
func validateSchema(schemaRef string, data any) error {
	if schemaRef == "" {
		return errors.New("openai: schema reference is required")
	}
	return nil
}

type requestPayload struct {
	Model           string         `json:"model"`
	Input           any            `json:"input"`
	MaxOutputTokens int            `json:"max_output_tokens,omitempty"`
	Temperature     float64        `json:"temperature,omitempty"`
	Text            textSection    `json:"text"`
	Metadata        map[string]any `json:"metadata,omitempty"`
}

type textSection struct {
	Format textFormat `json:"format"`
}

type textFormat struct {
	Type       string         `json:"type"`
	Name       string         `json:"name"`
	JSONSchema textJSONSchema `json:"json_schema"`
}

type textJSONSchema struct {
	Name   string          `json:"name"`
	Strict bool            `json:"strict"`
	Schema json.RawMessage `json:"schema"`
}

type inputMessage struct {
	Role    string         `json:"role,omitempty"`
	Content []contentBlock `json:"content,omitempty"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type responsePayload struct {
	Output []responseMessage `json:"output"`
	Usage  usagePayload      `json:"usage"`
}

type responseMessage struct {
	Content []responseContent `json:"content"`
}

type responseContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type usagePayload struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
	InputTokens      int `json:"input_tokens"`
	OutputTokens     int `json:"output_tokens"`
}

func buildInputMessages(call aiwf.ModelCall) ([]inputMessage, error) {
	var messages []inputMessage

	if call.SystemPrompt != "" {
		messages = append(messages, inputMessage{
			Role: "system",
			Content: []contentBlock{
				{Type: "input_text", Text: call.SystemPrompt},
			},
		})
	}

	var userParts []string
	if call.UserPrompt != "" {
		userParts = append(userParts, call.UserPrompt)
	}
	if call.Payload != nil {
		data, err := json.Marshal(call.Payload)
		if err != nil {
			return nil, fmt.Errorf("openai: marshal payload: %w", err)
		}
		userParts = append(userParts, string(data))
	}
	userText := strings.TrimSpace(strings.Join(userParts, "\n\n"))
	if userText != "" {
		messages = append(messages, inputMessage{
			Role: "user",
			Content: []contentBlock{
				{Type: "input_text", Text: userText},
			},
		})
	}

	if len(messages) == 0 {
		return nil, errors.New("openai: empty input")
	}

	return messages, nil
}

func buildJSONSchemaFormat(call aiwf.ModelCall) (textSection, error) {
	if len(call.OutputSchema) == 0 {
		return textSection{}, errors.New("openai: output schema is required")
	}

	name := schemaFormatName(call.OutputSchemaRef)

	return textSection{
		Format: textFormat{
			Type: "json_schema",
			JSONSchema: textJSONSchema{
				Name:   name,
				Strict: true,
				Schema: call.OutputSchema,
			},
		},
	}, nil
}

func schemaFormatName(ref string) string {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "aiwf_output"
	}

	candidate := ref
	if strings.Contains(candidate, "://") {
		if idx := strings.LastIndex(candidate, "/"); idx >= 0 && idx < len(candidate)-1 {
			candidate = candidate[idx+1:]
		}
	} else if strings.Contains(candidate, "/") {
		candidate = filepath.Base(candidate)
	}

	candidate = strings.TrimSuffix(candidate, filepath.Ext(candidate))
	if candidate == "" {
		candidate = "aiwf_output"
	}

	sanitized := make([]rune, 0, len(candidate))
	for _, r := range candidate {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			sanitized = append(sanitized, r)
			continue
		}
		if r == '_' || r == '-' {
			sanitized = append(sanitized, r)
			continue
		}
		sanitized = append(sanitized, '_')
	}

	name := strings.TrimLeftFunc(string(sanitized), func(r rune) bool { return r == '_' || r == '-' })
	if name == "" {
		name = "aiwf_output"
	}
	return name
}

func buildMetadata(call aiwf.ModelCall) map[string]any {
	if call.ThreadID == "" && len(call.ThreadMetadata) == 0 {
		return nil
	}
	meta := make(map[string]any)
	if call.ThreadID != "" {
		meta["thread_id"] = call.ThreadID
	}
	if len(call.ThreadMetadata) > 0 {
		for k, v := range call.ThreadMetadata {
			meta[k] = v
		}
	}
	return meta
}

func extractStructuredText(messages []responseMessage) (string, error) {
	if len(messages) == 0 {
		return "", errors.New("openai: empty output")
	}

	var sb strings.Builder
	for _, message := range messages {
		for _, block := range message.Content {
			text := strings.TrimSpace(block.Text)
			if text == "" {
				continue
			}
			if sb.Len() > 0 {
				sb.WriteByte('\n')
			}
			sb.WriteString(text)
		}
	}

	if sb.Len() == 0 {
		return "", errors.New("openai: empty output content")
	}

	return sb.String(), nil
}
