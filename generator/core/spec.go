package core

// Spec описывает parsed YAML.
type Spec struct {
	Version        string                   `yaml:"version"`
	SchemaRegistry SchemaRegistrySpec       `yaml:"schema_registry"`
	Imports        []ImportSpec             `yaml:"imports"`
	Assistants     map[string]AssistantSpec `yaml:"assistants"`
	Workflows      map[string]WorkflowSpec  `yaml:"workflows"`
	Resolved       SpecResolution           `yaml:"-"`
}

// SchemaRegistrySpec описывает расположение JSON Schema.
type SchemaRegistrySpec struct {
	Root string `yaml:"root"`
}

// ImportSpec описывает подключение YAML-файла с типами.
type ImportSpec struct {
	As   string `yaml:"as"`
	Path string `yaml:"path"`
}

// SpecResolution содержит вспомогательные структуры, полученные при загрузке.
type SpecResolution struct {
	TypeRegistry map[string]*SchemaDocument
}

// SchemaDocument описывает загруженную схему (из JSON или YAML-типов).
type SchemaDocument struct {
	ID     string
	Name   string
	Source string
	Alias  string
	Data   []byte
}

// AssistantSpec описывает агента в YAML.
type AssistantSpec struct {
	Use             string   `yaml:"use"`
	Model           string   `yaml:"model"`
	SystemPrompt    string   `yaml:"system_prompt"`
	InputSchemaRef  string   `yaml:"input_schema_ref"`
	OutputSchemaRef string   `yaml:"output_schema_ref"`
	DependsOn       []string `yaml:"depends_on"`
	// Resolved пути заполняются загрузчиком.
	Resolved AssistantResolution `yaml:"-"`
}

// AssistantResolution содержит абсолютные пути к схемам.
type AssistantResolution struct {
	InputSchemaPath  string
	OutputSchemaPath string
	InputSchema      *SchemaDocument
	OutputSchema     *SchemaDocument
}

// WorkflowSpec описывает workflow.
type WorkflowSpec struct {
	Description string        `yaml:"description"`
	DAG         []WorkflowDAG `yaml:"dag"`
}

// WorkflowDAG описывает шаг воркфлоу.
type WorkflowDAG struct {
	Step         string         `yaml:"step"`
	Assistant    string         `yaml:"assistant"`
	Needs        []string       `yaml:"needs"`
	Scatter      *ScatterSpec   `yaml:"scatter"`
	InputBinding map[string]any `yaml:"input_binding"`
}

// ScatterSpec описывает fan-out шаг.
type ScatterSpec struct {
	From        string `yaml:"from"`
	As          string `yaml:"as"`
	Concurrency int    `yaml:"concurrency"`
}

// ValidationError описывает ошибку загрузки.
type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Msg
}

// ValidationWarning описывает предупреждение.
type ValidationWarning struct {
	Field string
	Msg   string
}
