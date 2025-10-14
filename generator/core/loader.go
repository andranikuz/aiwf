package core

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

// LoadSpec читает YAML и возвращает нормализованный Spec.
func LoadSpec(path string) (*Spec, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("core: read yaml: %w", err)
	}

	spec := &Spec{}
	if err := yaml.Unmarshal(raw, spec); err != nil {
		return nil, fmt.Errorf("core: parse yaml: %w", err)
	}

	if err := validateSpec(path, spec); err != nil {
		return nil, err
	}

	if err := resolveAssistantSchemas(path, spec); err != nil {
		return nil, err
	}

	if err := validateWorkflows(spec); err != nil {
		return nil, err
	}

	return spec, nil
}

func validateSpec(p string, spec *Spec) error {
	if spec.SchemaRegistry.Root == "" {
		return &ValidationError{Field: "schema_registry.root", Msg: "должен быть заполнен"}
	}
	if len(spec.Assistants) == 0 {
		return &ValidationError{Field: "assistants", Msg: "не найдены"}
	}
	if len(spec.Workflows) == 0 {
		return &ValidationError{Field: "workflows", Msg: "не найдены"}
	}
	return nil
}

func resolveAssistantSchemas(specPath string, spec *Spec) error {
	base := filepath.Dir(specPath)
	registryRoot := resolvePath(base, spec.SchemaRegistry.Root)

    merr := &MultiError{}

    for name, assistant := range spec.Assistants {
        if assistant.OutputSchemaRef == "" {
            merr.Append(&ValidationError{Field: fmt.Sprintf("assistants.%s.output_schema_ref", name), Msg: "должен быть заполнен"})
            continue
        }

        outputPath := filepath.Join(registryRoot, assistant.OutputSchemaRef)
        if err := ensureJSONSchema(outputPath); err != nil {
            merr.Append(&ValidationError{Field: fmt.Sprintf("assistants.%s.output_schema_ref", name), Msg: err.Error()})
            continue
        }

        var inputPath string
        if assistant.InputSchemaRef != "" {
            inputPath = filepath.Join(registryRoot, assistant.InputSchemaRef)
            if err := ensureJSONSchema(inputPath); err != nil {
                merr.Append(&ValidationError{Field: fmt.Sprintf("assistants.%s.input_schema_ref", name), Msg: err.Error()})
            }
        }

        res := assistant.Resolved
        res.OutputSchemaPath = outputPath
		res.InputSchemaPath = inputPath
		assistant.Resolved = res
		spec.Assistants[name] = assistant
	}

    if merr.HasErrors() {
        return merr
    }

    return nil
}

func validateWorkflows(spec *Spec) error {
	for wfName, wf := range spec.Workflows {
		for idx, step := range wf.DAG {
			if step.Assistant == "" {
				return &ValidationError{Field: fmt.Sprintf("workflows.%s.dag[%d].assistant", wfName, idx), Msg: "должен быть заполнен"}
			}
			if _, ok := spec.Assistants[step.Assistant]; !ok {
				return &ValidationError{Field: fmt.Sprintf("workflows.%s.dag[%d].assistant", wfName, idx), Msg: "ассистент не найден"}
			}
		}
	}
	return nil
}

func resolvePath(base, target string) string {
	if filepath.IsAbs(target) {
		return filepath.Clean(target)
	}
	return filepath.Clean(filepath.Join(base, target))
}

func ensureJSONSchema(path string) error {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("schema не найдена: %s", path)
	}
	if err != nil {
		return err
	}

	loader := gojsonschema.NewBytesLoader(data)
	if _, err := gojsonschema.NewSchema(loader); err != nil {
		return fmt.Errorf("schema некорректна (%s): %v", path, err)
	}
	return nil
}
