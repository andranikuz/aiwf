package core

// Spec описывает parsed YAML.
type Spec struct {
	Version        string                   `yaml:"version"`
	SchemaRegistry SchemaRegistrySpec       `yaml:"schema_registry"`
	Imports        []ImportSpec             `yaml:"imports"`
	Threads        map[string]ThreadSpec    `yaml:"threads"`
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
	Thread          *ThreadBindingSpec `yaml:"thread"`
	Dialog          *DialogSpec        `yaml:"dialog"`
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

// ThreadSpec описывает политику работы с тредами.
type ThreadSpec struct {
	Provider      string         `yaml:"provider"`
	Strategy      string         `yaml:"strategy"`
	Create        bool           `yaml:"create"`
	CloseOnFinish bool           `yaml:"close_on_finish"`
	TTLHours      int            `yaml:"ttl_hours"`
	Metadata      map[string]any `yaml:"metadata"`
}

// ThreadBindingSpec привязывает ассистента/шаг к политике треда.
type ThreadBindingSpec struct {
	Use      string `yaml:"use"`
	Strategy string `yaml:"strategy"`
}

// DialogSpec описывает диалоговые настройки.
type DialogSpec struct {
	MaxRounds int `yaml:"max_rounds"`
}

// ApprovalSpec описывает расширенные правила проверки.
type ApprovalSpec struct {
	Review               map[string]any `yaml:"review"`
	RequireForRetry      bool           `yaml:"require_for_retry"`
	MaxRetries           int            `yaml:"max_retries"`
	OnApprove            map[string]any `yaml:"on_approve"`
	OnReject             map[string]any `yaml:"on_reject"`
	FeedbackTemplate     string         `yaml:"feedback_template"`
	AutoContinueOnApprove *bool         `yaml:"auto_continue_on_approve"`
}

// NextStepSpec описывает переход к следующему шагу.
type NextStepSpec struct {
	Step             string         `yaml:"step"`
	InputBinding     map[string]any `yaml:"input_binding"`
	InputContractRef string         `yaml:"input_contract_ref"`
}

// WorkflowSpec описывает workflow.
type WorkflowSpec struct {
	Description string        `yaml:"description"`
	DAG         []WorkflowDAG `yaml:"dag"`
	Thread      *ThreadBindingSpec `yaml:"thread"`
}

// WorkflowDAG описывает шаг воркфлоу.
type WorkflowDAG struct {
	Step         string         `yaml:"step"`
	Assistant    string         `yaml:"assistant"`
	Needs        []string       `yaml:"needs"`
	Scatter      *ScatterSpec   `yaml:"scatter"`
	InputBinding map[string]any `yaml:"input_binding"`
	Thread       *ThreadBindingSpec `yaml:"thread"`
	Dialog       *DialogSpec        `yaml:"dialog"`
	Approval     *ApprovalSpec      `yaml:"approval"`
	Next         *NextStepSpec      `yaml:"next"`
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
