package aiwf

import (
	"context"
	"encoding/json"
	"fmt"
)

// AgentConfig содержит конфигурацию агента
type AgentConfig struct {
	Name           string
	Model          string
	SystemPrompt   string
	InputTypeName  string
	OutputTypeName string
	MaxTokens      int
	Temperature    float64
}

// AgentBase базовая реализация агента
type AgentBase struct {
	Config AgentConfig
	Client ModelClient
	Types  TypeProvider
}

// Name возвращает имя агента
func (a *AgentBase) Name() string {
	return a.Config.Name
}

// Model возвращает модель
func (a *AgentBase) Model() string {
	return a.Config.Model
}

// SystemPrompt возвращает системный промпт
func (a *AgentBase) SystemPrompt() string {
	return a.Config.SystemPrompt
}

// CallModel вызывает модель с типизированными данными
func (a *AgentBase) CallModel(ctx context.Context, input any, thread *ThreadState) (json.RawMessage, *Trace, error) {
	// Получаем метаданные типов если есть TypeProvider
	var typeMetadata any
	if a.Types != nil && a.Config.OutputTypeName != "" {
		meta, _ := a.Types.GetTypeMetadata(a.Config.OutputTypeName)
		typeMetadata = meta
	}

	call := ModelCall{
		Model:          a.Config.Model,
		SystemPrompt:   a.Config.SystemPrompt,
		Payload:        input,
		MaxTokens:      a.Config.MaxTokens,
		Temperature:    a.Config.Temperature,
		InputTypeName:  a.Config.InputTypeName,
		OutputTypeName: a.Config.OutputTypeName,
		TypeMetadata:   typeMetadata,
	}

	// Добавляем информацию о треде если есть
	if thread != nil {
		call.ThreadID = thread.ID
		call.ThreadMetadata = thread.Metadata
	}

	// Вызываем модель
	result, tokens, err := a.Client.CallJSONSchema(ctx, call)
	if err != nil {
		return nil, nil, fmt.Errorf("model call failed: %w", err)
	}

	trace := &Trace{
		StepName: a.Config.Name,
		Usage:    tokens,
		Attempts: 1,
	}

	return result, trace, nil
}

// WorkflowContext контекст выполнения воркфлоу
type WorkflowContext struct {
	Traces         []*Trace
	Artifacts      map[string]any
	ThreadState    *ThreadState
	ArtifactStore  ArtifactStore
}

// AddTrace добавляет трейс в контекст
func (w *WorkflowContext) AddTrace(trace *Trace) {
	w.Traces = append(w.Traces, trace)
}

// SaveArtifact сохраняет промежуточный результат
func (w *WorkflowContext) SaveArtifact(key string, data any) error {
	if w.Artifacts == nil {
		w.Artifacts = make(map[string]any)
	}
	w.Artifacts[key] = data

	// Если есть ArtifactStore, сохраняем туда тоже
	if w.ArtifactStore != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return err
		}
		return w.ArtifactStore.Put(context.Background(), key, jsonData)
	}

	return nil
}

// GetArtifact получает сохранённый артефакт
func (w *WorkflowContext) GetArtifact(key string) (any, bool) {
	if w.Artifacts == nil {
		return nil, false
	}
	data, ok := w.Artifacts[key]
	return data, ok
}