// Package blogsdk - пример сгенерированного SDK
package blogsdk

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/andranikuz/aiwf/runtime/go/aiwf"
)

// ============ TYPES (сгенерировано из TypeDef) ============

type ExtractRequest struct {
	Text           string     `json:"text"`
	ExtractionMode ExtractMode `json:"extraction_mode"`
}

type ExtractMode string

const (
	ExtractModeEntities     ExtractMode = "entities"
	ExtractModeRelationships ExtractMode = "relationships"
	ExtractModeFull         ExtractMode = "full"
)

type Entity struct {
	Type       EntityType `json:"type"`
	Value      string     `json:"value"`
	Confidence float64    `json:"confidence"`
}

type EntityType string

const (
	EntityTypePerson       EntityType = "person"
	EntityTypeOrganization EntityType = "organization"
	EntityTypeLocation     EntityType = "location"
	EntityTypeDate         EntityType = "date"
	EntityTypeAmount       EntityType = "amount"
)

type ExtractedData struct {
	Entities      []Entity       `json:"entities"`
	Relationships []Relationship `json:"relationships"`
	Metadata      Metadata       `json:"metadata"`
}

type Relationship struct {
	Subject   string `json:"subject"`
	Predicate string `json:"predicate"`
	Object    string `json:"object"`
}

type Metadata struct {
	ExtractionTimestamp time.Time `json:"extraction_timestamp"`
	SourceLength        int       `json:"source_length"`
	DetectedLanguage    string    `json:"detected_language"`
}

// ============ VALIDATION (сгенерировано из ограничений) ============

func ValidateExtractRequest(r *ExtractRequest) error {
	if len(r.Text) < 1 || len(r.Text) > 10000 {
		return fmt.Errorf("text length must be between 1 and 10000, got %d", len(r.Text))
	}
	return nil
}

// ============ TYPE METADATA (для провайдеров) ============

// TypeDefinitions экспортирует метаданные типов
var TypeDefinitions = map[string]any{
	"ExtractRequest": map[string]any{
		"kind": "object",
		"properties": map[string]any{
			"text": map[string]any{
				"kind": "string",
				"min_length": 1,
				"max_length": 10000,
			},
			"extraction_mode": map[string]any{
				"kind": "enum",
				"values": []string{"entities", "relationships", "full"},
			},
		},
	},
	"ExtractedData": map[string]any{
		"kind": "object",
		"properties": map[string]any{
			"entities": map[string]any{
				"kind": "array",
				"items": "Entity",
			},
			"relationships": map[string]any{
				"kind": "array",
				"items": "Relationship",
			},
			"metadata": map[string]any{
				"kind": "object",
				"ref": "Metadata",
			},
		},
	},
	// ... остальные типы
}

// ============ AGENTS ============

type DataExtractorAgent struct {
	aiwf.AgentBase
}

func NewDataExtractorAgent(client aiwf.ModelClient) *DataExtractorAgent {
	return &DataExtractorAgent{
		AgentBase: aiwf.AgentBase{
			Config: aiwf.AgentConfig{
				Name:           "data_extractor",
				Model:          "gpt-4o-mini",
				SystemPrompt:   "Extract structured information from text.\nReturn only explicitly stated facts.",
				InputTypeName:  "ExtractRequest",
				OutputTypeName: "ExtractedData",
				MaxTokens:      1000,
				Temperature:    0.3,
			},
			Client: client,
		},
	}
}

// Run выполняет агента с типизированными параметрами
func (a *DataExtractorAgent) Run(ctx context.Context, input ExtractRequest) (*ExtractedData, *aiwf.Trace, error) {
	// Валидация входных данных
	if err := ValidateExtractRequest(&input); err != nil {
		return nil, nil, fmt.Errorf("validation failed: %w", err)
	}

	// Вызов модели
	result, trace, err := a.CallModel(ctx, input, nil)
	if err != nil {
		return nil, trace, err
	}

	// Парсинг результата
	var output ExtractedData
	if err := json.Unmarshal(result, &output); err != nil {
		return nil, trace, fmt.Errorf("failed to parse response: %w", err)
	}

	return &output, trace, nil
}

// RunWithThread выполняет агента с состоянием треда
func (a *DataExtractorAgent) RunWithThread(ctx context.Context, input ExtractRequest, thread *aiwf.ThreadState) (*ExtractedData, *aiwf.Trace, error) {
	if err := ValidateExtractRequest(&input); err != nil {
		return nil, nil, fmt.Errorf("validation failed: %w", err)
	}

	result, trace, err := a.CallModel(ctx, input, thread)
	if err != nil {
		return nil, trace, err
	}

	var output ExtractedData
	if err := json.Unmarshal(result, &output); err != nil {
		return nil, trace, fmt.Errorf("failed to parse response: %w", err)
	}

	return &output, trace, nil
}

// ============ SERVICE ============

type Agents struct {
	DataExtractor *DataExtractorAgent
	// ... другие агенты
}

type Service struct {
	client        aiwf.ModelClient
	threadManager aiwf.ThreadManager
	artifactStore aiwf.ArtifactStore
	Agents        *Agents
}

// NewService создаёт новый сервис
func NewService(client aiwf.ModelClient) *Service {
	return &Service{
		client: client,
		Agents: &Agents{
			DataExtractor: NewDataExtractorAgent(client),
			// ... инициализация других агентов
		},
	}
}

// WithThreadManager добавляет менеджер тредов
func (s *Service) WithThreadManager(tm aiwf.ThreadManager) *Service {
	s.threadManager = tm
	return s
}

// WithArtifactStore добавляет хранилище артефактов
func (s *Service) WithArtifactStore(store aiwf.ArtifactStore) *Service {
	s.artifactStore = store
	return s
}

// GetTypeMetadata реализует aiwf.TypeProvider
func (s *Service) GetTypeMetadata(typeName string) (any, error) {
	if meta, ok := TypeDefinitions[typeName]; ok {
		return meta, nil
	}
	return nil, fmt.Errorf("type %s not found", typeName)
}

// GetInputTypeFor реализует aiwf.TypeProvider
func (s *Service) GetInputTypeFor(agentName string) (string, any, error) {
	switch agentName {
	case "data_extractor":
		return "ExtractRequest", TypeDefinitions["ExtractRequest"], nil
	default:
		return "", nil, fmt.Errorf("agent %s not found", agentName)
	}
}

// GetOutputTypeFor реализует aiwf.TypeProvider
func (s *Service) GetOutputTypeFor(agentName string) (string, any, error) {
	switch agentName {
	case "data_extractor":
		return "ExtractedData", TypeDefinitions["ExtractedData"], nil
	default:
		return "", nil, fmt.Errorf("agent %s not found", agentName)
	}
}