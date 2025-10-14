package core

import "fmt"

// IR описывает нормализованный набор ассистентов и воркфлоу.
type IR struct {
	Assistants map[string]IRAssistant
	Workflows  map[string]IRWorkflow
}

// IRAssistant содержит сведения для генерации SDK.
type IRAssistant struct {
    Name             string
    Model            string
    SystemPrompt     string
    Use              string
    InputSchemaPath  string
    OutputSchemaPath string
    DependsOn        []string
    InputSchemaRef   string
    OutputSchemaRef  string
}

// IRWorkflow описывает workflow и его шаги.
type IRWorkflow struct {
	Name        string
	Description string
	Steps       []IRStep
}

// IRStep описывает шаг внутри workflow.
type IRStep struct {
	Name         string
	Assistant    string
	Needs        []string
	Scatter      *ScatterSpec
	InputBinding map[string]any
}

// BuildIR преобразует Spec в IR и выполняет дополнительную валидацию.
func BuildIR(spec *Spec) (*IR, error) {
	if spec == nil {
		return nil, fmt.Errorf("core: spec is nil")
	}

	ir := &IR{
		Assistants: make(map[string]IRAssistant, len(spec.Assistants)),
		Workflows:  make(map[string]IRWorkflow, len(spec.Workflows)),
	}

	merr := &MultiError{}

	for name, as := range spec.Assistants {
		if as.Resolved.OutputSchemaPath == "" {
			merr.Append(&ValidationError{Field: fmt.Sprintf("assistants.%s.output_schema_ref", name), Msg: "schema path не вычислен"})
			continue
		}

        assistant := IRAssistant{
            Name:             name,
            Model:            as.Model,
            SystemPrompt:     as.SystemPrompt,
            Use:              as.Use,
            InputSchemaPath:  as.Resolved.InputSchemaPath,
            OutputSchemaPath: as.Resolved.OutputSchemaPath,
            DependsOn:        cloneSlice(as.DependsOn),
            InputSchemaRef:   as.InputSchemaRef,
            OutputSchemaRef:  as.OutputSchemaRef,
        }
		ir.Assistants[name] = assistant
	}

	usedAssistants := make(map[string]bool)

	for wfName, wfSpec := range spec.Workflows {
		seenSteps := make(map[string]int)
		steps := make([]IRStep, 0, len(wfSpec.DAG))
		if len(wfSpec.DAG) == 0 {
			merr.AppendWarning(&ValidationWarning{Field: fmt.Sprintf("workflows.%s", wfName), Msg: "workflow не содержит шагов"})
		}
		for idx, step := range wfSpec.DAG {
			invalid := false
			if step.Step == "" {
				merr.Append(&ValidationError{Field: fmt.Sprintf("workflows.%s.dag[%d].step", wfName, idx), Msg: "должен быть заполнен"})
				continue
			}
			if step.Assistant == "" {
				merr.Append(&ValidationError{Field: fmt.Sprintf("workflows.%s.dag[%d].assistant", wfName, idx), Msg: "должен быть заполнен"})
				continue
			}
			if _, exists := ir.Assistants[step.Assistant]; !exists {
				merr.Append(&ValidationError{Field: fmt.Sprintf("workflows.%s.dag[%d].assistant", wfName, idx), Msg: "ассистент не найден"})
				continue
			}
			usedAssistants[step.Assistant] = true
			if prevIdx, dup := seenSteps[step.Step]; dup {
				merr.Append(&ValidationError{Field: fmt.Sprintf("workflows.%s.dag[%d].step", wfName, idx), Msg: fmt.Sprintf("дубликат шага, уже объявлен в dag[%d]", prevIdx)})
				continue
			}
			seenSteps[step.Step] = idx

			for _, need := range step.Needs {
				if _, ok := seenSteps[need]; !ok {
					merr.Append(&ValidationError{Field: fmt.Sprintf("workflows.%s.dag[%d].needs", wfName, idx), Msg: "ссылается на шаг, который ещё не определён"})
					invalid = true
				}
			}

			if step.Scatter != nil {
				if step.Scatter.From == "" {
					merr.Append(&ValidationError{Field: fmt.Sprintf("workflows.%s.dag[%d].scatter.from", wfName, idx), Msg: "должен быть заполнен"})
					invalid = true
				}
				if step.Scatter.As == "" {
					merr.Append(&ValidationError{Field: fmt.Sprintf("workflows.%s.dag[%d].scatter.as", wfName, idx), Msg: "должен быть заполнен"})
					invalid = true
				}
				if step.Scatter.Concurrency < 0 {
					merr.Append(&ValidationError{Field: fmt.Sprintf("workflows.%s.dag[%d].scatter.concurrency", wfName, idx), Msg: "не может быть отрицательным"})
					invalid = true
				}
			}

			if invalid {
				continue
			}

			irStep := IRStep{
				Name:         step.Step,
				Assistant:    step.Assistant,
				Needs:        cloneSlice(step.Needs),
				Scatter:      cloneScatter(step.Scatter),
				InputBinding: cloneMap(step.InputBinding),
			}
			steps = append(steps, irStep)
		}
		ir.Workflows[wfName] = IRWorkflow{
			Name:        wfName,
			Description: wfSpec.Description,
			Steps:       steps,
		}
	}

	for name := range spec.Assistants {
		if !usedAssistants[name] {
			merr.AppendWarning(&ValidationWarning{Field: fmt.Sprintf("assistants.%s", name), Msg: "ассистент не используется ни в одном workflow"})
		}
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

func cloneScatter(in *ScatterSpec) *ScatterSpec {
	if in == nil {
		return nil
	}
	copy := *in
	return &copy
}
