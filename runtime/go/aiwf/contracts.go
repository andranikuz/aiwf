package aiwf

import (
	"context"
	"encoding/json"
	"time"
)

// Tokens описывает затраты токенов, отражая раздел Contracts в sdk.md.
type Tokens struct {
	Prompt     int
	Completion int
	Total      int
}

// Trace фиксирует наблюдаемость выполнения шага, совпадая с ожиданиями SDK.
type Trace struct {
	StepName   string
	Usage      Tokens
	Attempts   int
	Duration   time.Duration
	ArtifactID string
}

// ModelCall описывает запрос к LLM c привязкой к JSON Schema.
type ModelCall struct {
	Model           string
	InputSchemaRef  string
	OutputSchemaRef string
	OutputSchema    json.RawMessage
	SystemPrompt    string
	UserPrompt      string
	MaxTokens       int
	Temperature     float64
	Stream          bool
	Payload         any
}

// StreamChunk описывает инкрементальные ответы модели при потоковой генерации.
type StreamChunk struct {
	Data       []byte
	Done       bool
	Partial    any
	Timestamps map[string]any
}

// ModelClient оборачивает вызовы модели, возвращая строго типизированные результаты.
type ModelClient interface {
	CallJSONSchema(ctx context.Context, call ModelCall) ([]byte, Tokens, error)
	CallJSONSchemaStream(ctx context.Context, call ModelCall) (<-chan StreamChunk, Tokens, error)
}

// Workflow описывает типизированный раннер воркфлоу.
type Workflow[I any, O any] interface {
	Run(ctx context.Context, input I) (O, *Trace, error)
	RunStep(ctx context.Context, step string, payload any) ([]byte, *Trace, error)
}

// RetryPolicy принимает решение о retry, синхронизируясь с ожиданиями рантайма.
type RetryPolicy interface {
	ShouldRetry(err error, attempt int) (retry bool, backoff time.Duration)
}

// ArtifactStore сохраняет артефакты шага (промежуточные JSON, промпты).
type ArtifactStore interface {
	Put(ctx context.Context, key string, data []byte) error
	Get(ctx context.Context, key string) ([]byte, bool, error)
	Key(workflow, step, item, inputHash string) string
}
