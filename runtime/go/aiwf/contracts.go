package aiwf

import (
	"context"
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

// ModelCall описывает запрос к LLM.
type ModelCall struct {
	Model          string
	SystemPrompt   string
	UserPrompt     string
	MaxTokens      int
	Temperature    float64
	Stream         bool
	Payload        any // Входные данные (уже типизированные)
	ThreadID       string
	ThreadMetadata map[string]any

	// Метаданные типов для провайдера
	InputTypeName  string // Имя входного типа
	OutputTypeName string // Имя выходного типа
	TypeMetadata   any    // Опциональные метаданные типа (например, TypeDef для провайдера)
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

// ThreadState описывает состояние треда в провайдере.
type ThreadState struct {
	ID       string
	Metadata map[string]any
}

// ThreadBinding описывает политику работы с тредом.
type ThreadBinding struct {
	Name     string
	Provider string
	Strategy string
	Metadata map[string]any
}

// ThreadManager управляет жизненным циклом тредов между шагами.
type ThreadManager interface {
	Start(ctx context.Context, assistant string, binding ThreadBinding) (*ThreadState, error)
	Continue(ctx context.Context, state *ThreadState, feedback string) error
	Close(ctx context.Context, state *ThreadState) error
}

// TypeProvider предоставляет метаданные типов для провайдеров.
// SDK реализует этот интерфейс для экспорта информации о типах.
type TypeProvider interface {
	GetTypeMetadata(typeName string) (any, error)
	GetInputTypeFor(agentName string) (string, any, error)
	GetOutputTypeFor(agentName string) (string, any, error)
}

// Agent описывает базовый интерфейс агента.
// Конкретные агенты в SDK будут иметь типизированные методы.
type Agent interface {
	Name() string
	Model() string
	SystemPrompt() string
}
