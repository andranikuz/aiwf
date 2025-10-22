package core

// Spec описывает parsed YAML.
type Spec struct {
	Version    string                   `yaml:"version"`
	Imports    []ImportSpec             `yaml:"imports"`
	Types      map[string]interface{}   `yaml:"types"`
	Threads    map[string]ThreadSpec    `yaml:"threads"`
	Assistants map[string]AssistantSpec `yaml:"assistants"`
	Resolved   SpecResolution           `yaml:"-"`
}

// ImportSpec описывает подключение YAML-файла с типами.
type ImportSpec struct {
	As   string `yaml:"as"`
	Path string `yaml:"path"`
}

// SpecResolution содержит вспомогательные структуры, полученные при загрузке.
type SpecResolution struct {
	TypeRegistry *TypeRegistry
}

// AssistantSpec описывает агента в YAML.
type AssistantSpec struct {
	Use          string   `yaml:"use"`
	Model        string   `yaml:"model"`
	SystemPrompt string   `yaml:"system_prompt"`
	InputType    string   `yaml:"input_type"`
	OutputType   string   `yaml:"output_type"`
	DependsOn    []string `yaml:"depends_on"`
	Thread       *ThreadBindingSpec `yaml:"thread"`
	Dialog       *DialogSpec        `yaml:"dialog"`
	Resolved     AssistantResolution `yaml:"-"`
}

// AssistantResolution содержит разрешённые типы.
type AssistantResolution struct {
	InputType  *TypeDef
	OutputType *TypeDef
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
