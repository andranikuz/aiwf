package core

import "fmt"

// IR описывает нормализованный набор ассистентов.
type IR struct {
    Assistants map[string]IRAssistant
    Threads    map[string]ThreadSpec
    Types      *TypeRegistry
}

// IRAssistant содержит сведения для генерации SDK.
type IRAssistant struct {
    Name           string
    Model          string
    SystemPrompt   string
    Use            string
    InputTypeName  string
    OutputTypeName string
    MaxTokens      int
    Temperature    float64
    InputType      *TypeDef
    OutputType     *TypeDef
    DependsOn      []string
    Thread         *ThreadBindingSpec
    Dialog         *DialogSpec
}


// BuildIR преобразует Spec в IR и выполняет дополнительную валидацию.
func BuildIR(spec *Spec) (*IR, error) {
	if spec == nil {
		return nil, fmt.Errorf("core: spec is nil")
	}

	// Resolve types first
	if err := ResolveSpec(spec); err != nil {
		return nil, fmt.Errorf("failed to resolve spec: %w", err)
	}

    ir := &IR{
        Assistants: make(map[string]IRAssistant, len(spec.Assistants)),
        Threads:    make(map[string]ThreadSpec, len(spec.Threads)),
        Types:      spec.Resolved.TypeRegistry,
    }

	merr := &MultiError{}

	for name, as := range spec.Assistants {
		// Если output_type не указан, используем string по умолчанию
		outputTypeName := as.OutputType
		if outputTypeName == "" {
			outputTypeName = "string"
		}

        assistant := IRAssistant{
            Name:           name,
            Model:          as.Model,
            SystemPrompt:   as.SystemPrompt,
            Use:            as.Use,
            InputTypeName:  as.InputType,
            OutputTypeName: outputTypeName,
            MaxTokens:      as.MaxTokens,
            Temperature:    as.Temperature,
            InputType:      as.Resolved.InputType,
            OutputType:     as.Resolved.OutputType,
            DependsOn:      cloneSlice(as.DependsOn),
            Thread:         cloneThreadBinding(as.Thread),
            Dialog:         cloneDialog(as.Dialog),
        }
        ir.Assistants[name] = assistant
	}


    for name, thread := range spec.Threads {
        ir.Threads[name] = thread
    }



	if merr.HasErrors() {
		return nil, merr
	}

	if merr.HasWarnings() {
		return ir, merr
	}

    return ir, nil
}

func cloneSlice(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}


func cloneMap(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneThreadBinding(in *ThreadBindingSpec) *ThreadBindingSpec {
    if in == nil {
        return nil
    }
    copy := *in
    return &copy
}

func cloneDialog(in *DialogSpec) *DialogSpec {
    if in == nil {
        return nil
    }
    copy := *in
    return &copy
}
