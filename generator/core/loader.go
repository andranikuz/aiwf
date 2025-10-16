package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

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

	if err := normalizeThreads(spec); err != nil {
		return nil, err
	}

	if err := validateThreadBindings(spec); err != nil {
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
	if len(spec.Assistants) == 0 {
		return &ValidationError{Field: "assistants", Msg: "не найдены"}
	}
	if spec.Workflows == nil {
		spec.Workflows = make(map[string]WorkflowSpec)
	}
	if spec.SchemaRegistry.Root == "" && len(spec.Imports) == 0 {
		return &ValidationError{Field: "schema_registry.root", Msg: "должен быть заполнен или используйте imports"}
	}
	return nil
}

func resolveAssistantSchemas(specPath string, spec *Spec) error {
	base := filepath.Dir(specPath)
	registryRoot := resolvePath(base, spec.SchemaRegistry.Root)

	typeRegistry, err := buildTypeRegistry(base, spec.Imports)
	if err != nil {
		return err
	}
	spec.Resolved.TypeRegistry = typeRegistry

	merr := &MultiError{}

	for name, assistant := range spec.Assistants {
		if assistant.OutputSchemaRef == "" {
			merr.Append(&ValidationError{Field: fmt.Sprintf("assistants.%s.output_schema_ref", name), Msg: "должен быть заполнен"})
			continue
		}

		outputDoc, err := fetchSchemaDocument(base, registryRoot, assistant.OutputSchemaRef, typeRegistry)
		if err != nil {
			merr.Append(&ValidationError{Field: fmt.Sprintf("assistants.%s.output_schema_ref", name), Msg: err.Error()})
			continue
		}
		if err := validateSchemaDocument(outputDoc, typeRegistry); err != nil {
			merr.Append(&ValidationError{Field: fmt.Sprintf("assistants.%s.output_schema_ref", name), Msg: err.Error()})
			continue
		}

		var inputDoc *SchemaDocument
		if assistant.InputSchemaRef != "" {
			doc, err := fetchSchemaDocument(base, registryRoot, assistant.InputSchemaRef, typeRegistry)
			if err != nil {
				merr.Append(&ValidationError{Field: fmt.Sprintf("assistants.%s.input_schema_ref", name), Msg: err.Error()})
			} else if err := validateSchemaDocument(doc, typeRegistry); err != nil {
				merr.Append(&ValidationError{Field: fmt.Sprintf("assistants.%s.input_schema_ref", name), Msg: err.Error()})
			} else {
				inputDoc = doc
			}
		}

		res := assistant.Resolved
		res.OutputSchema = outputDoc
		res.OutputSchemaPath = outputDoc.Source
		if inputDoc != nil {
			res.InputSchema = inputDoc
			res.InputSchemaPath = inputDoc.Source
		}
		assistant.Resolved = res
		spec.Assistants[name] = assistant
	}

	if merr.HasErrors() {
		return merr
	}

	return nil
}

func normalizeThreads(spec *Spec) error {
	if spec.Threads == nil {
		spec.Threads = make(map[string]ThreadSpec)
		return nil
	}

	validStrategies := map[string]bool{
		"append":             true,
		"reset_before_step": true,
	}

	for name, thread := range spec.Threads {
		trimmedName := strings.TrimSpace(name)
		if trimmedName == "" {
			return &ValidationError{Field: "threads", Msg: "имя политики не может быть пустым"}
		}
		if thread.Provider == "" {
			return &ValidationError{Field: fmt.Sprintf("threads.%s.provider", name), Msg: "должен быть заполнен"}
		}
		if thread.Strategy == "" {
			thread.Strategy = "append"
		}
		if !validStrategies[thread.Strategy] {
			return &ValidationError{Field: fmt.Sprintf("threads.%s.strategy", name), Msg: "неподдерживаемое значение"}
		}
		if thread.TTLHours < 0 {
			return &ValidationError{Field: fmt.Sprintf("threads.%s.ttl_hours", name), Msg: "не может быть отрицательным"}
		}
		spec.Threads[name] = thread
	}
	return nil
}

func validateThreadBindings(spec *Spec) error {
	for name, as := range spec.Assistants {
		if as.Thread != nil {
			if err := ensureThreadBinding(spec, fmt.Sprintf("assistants.%s.thread", name), as.Thread); err != nil {
				return err
			}
		}
		if as.Dialog != nil && as.Dialog.MaxRounds < 0 {
			return &ValidationError{Field: fmt.Sprintf("assistants.%s.dialog.max_rounds", name), Msg: "не может быть отрицательным"}
		}
	}

	for wfName, wf := range spec.Workflows {
		if wf.Thread != nil {
			if err := ensureThreadBinding(spec, fmt.Sprintf("workflows.%s.thread", wfName), wf.Thread); err != nil {
				return err
			}
		}
		for idx, step := range wf.DAG {
			if step.Thread != nil {
				if err := ensureThreadBinding(spec, fmt.Sprintf("workflows.%s.dag[%d].thread", wfName, idx), step.Thread); err != nil {
					return err
				}
			}
			if step.Dialog != nil && step.Dialog.MaxRounds < 0 {
				return &ValidationError{Field: fmt.Sprintf("workflows.%s.dag[%d].dialog.max_rounds", wfName, idx), Msg: "не может быть отрицательным"}
			}
		}
	}
	return nil
}

func ensureThreadBinding(spec *Spec, field string, binding *ThreadBindingSpec) error {
	if binding == nil {
		return nil
	}
	if binding.Use != "" {
		if _, ok := spec.Threads[binding.Use]; !ok {
			return &ValidationError{Field: field + ".use", Msg: "политика не найдена"}
		}
	}
	if binding.Strategy != "" {
		switch binding.Strategy {
		case "append", "reset_before_step":
		default:
			return &ValidationError{Field: field + ".strategy", Msg: "неподдерживаемое значение"}
		}
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

func buildTypeRegistry(base string, imports []ImportSpec) (map[string]*SchemaDocument, error) {
	if len(imports) == 0 {
		return map[string]*SchemaDocument{}, nil
	}

	registry := make(map[string]*SchemaDocument)
	merr := &MultiError{}

	for idx, imp := range imports {
		fieldPrefix := fmt.Sprintf("imports[%d]", idx)
		if imp.Path == "" {
			merr.Append(&ValidationError{Field: fieldPrefix + ".path", Msg: "должен быть заполнен"})
			continue
		}
		if imp.As == "" {
			merr.Append(&ValidationError{Field: fieldPrefix + ".as", Msg: "должен быть заполнен"})
		}

		fullPath := resolvePath(base, imp.Path)
		data, err := os.ReadFile(fullPath)
		if errors.Is(err, fs.ErrNotExist) {
			merr.Append(&ValidationError{Field: fieldPrefix, Msg: fmt.Sprintf("файл не найден: %s", fullPath)})
			continue
		}
		if err != nil {
			merr.Append(&ValidationError{Field: fieldPrefix, Msg: err.Error()})
			continue
		}

		var parsed struct {
			Types map[string]map[string]any `yaml:"types"`
		}
		if err := yaml.Unmarshal(data, &parsed); err != nil {
			merr.Append(&ValidationError{Field: fieldPrefix, Msg: fmt.Sprintf("ошибка чтения типов: %v", err)})
			continue
		}

		for typeName, schema := range parsed.Types {
			if schema == nil {
				merr.Append(&ValidationError{Field: fmt.Sprintf("%s.types.%s", fieldPrefix, typeName), Msg: "пустое описание schema"})
				continue
			}
			idRaw, ok := schema["$id"].(string)
			if !ok || strings.TrimSpace(idRaw) == "" {
				merr.Append(&ValidationError{Field: fmt.Sprintf("%s.types.%s.$id", fieldPrefix, typeName), Msg: "должен быть заполнен"})
				continue
			}
			jsonBytes, err := json.Marshal(schema)
			if err != nil {
				merr.Append(&ValidationError{Field: fmt.Sprintf("%s.types.%s", fieldPrefix, typeName), Msg: fmt.Sprintf("не удалось сериализовать schema: %v", err)})
				continue
			}
			if existing, exists := registry[idRaw]; exists {
				alias := existing.Alias
				if alias == "" {
					alias = existing.Source
				}
				merr.Append(&ValidationError{Field: fmt.Sprintf("%s.types.%s.$id", fieldPrefix, typeName), Msg: fmt.Sprintf("дубликат $id, уже определён в %s", alias)})
				continue
			}
			registry[idRaw] = &SchemaDocument{
				ID:     idRaw,
				Name:   typeName,
				Source: fullPath,
				Alias:  imp.As,
				Data:   jsonBytes,
			}
		}
	}

	if merr.HasErrors() {
		return nil, merr
	}

	return registry, nil
}

func fetchSchemaDocument(base, registryRoot, ref string, registry map[string]*SchemaDocument) (*SchemaDocument, error) {
	if strings.HasPrefix(ref, "aiwf://") {
		doc, ok := registry[ref]
		if !ok {
			return nil, fmt.Errorf("schema не найдена: %s", ref)
		}
		return doc, nil
	}

	var target string
	if registryRoot != "" {
		target = resolvePath(registryRoot, ref)
	} else {
		target = resolvePath(base, ref)
	}

	data, err := os.ReadFile(target)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("schema не найдена: %s", target)
	}
	if err != nil {
		return nil, err
	}

	return &SchemaDocument{
		ID:     "",
		Name:   filepath.Base(ref),
		Source: target,
		Data:   data,
	}, nil
}

func validateSchemaDocument(doc *SchemaDocument, registry map[string]*SchemaDocument) error {
	if doc == nil {
		return errors.New("schema не найдена")
	}
	if len(doc.Data) == 0 {
		return fmt.Errorf("schema пуста (%s)", doc.Source)
	}

	schemaLoader := gojsonschema.NewSchemaLoader()
	schemaLoader.AutoDetect = true

	seen := make(map[string]bool)
	for id, typedoc := range registry {
		if typedoc == nil || len(typedoc.Data) == 0 {
			continue
		}
		if id == "" || seen[id] {
			continue
		}
		if err := schemaLoader.AddSchema(id, gojsonschema.NewBytesLoader(typedoc.Data)); err != nil {
			return fmt.Errorf("schema registry %s: %v", id, err)
		}
		seen[id] = true
	}

	var rootLoader gojsonschema.JSONLoader
	if doc.ID != "" {
		if !seen[doc.ID] {
			if err := schemaLoader.AddSchema(doc.ID, gojsonschema.NewBytesLoader(doc.Data)); err != nil {
				return fmt.Errorf("schema некорректна (%s): %v", doc.ID, err)
			}
			seen[doc.ID] = true
		}
		rootLoader = gojsonschema.NewReferenceLoader(doc.ID)
	} else {
		rootLoader = gojsonschema.NewBytesLoader(doc.Data)
	}

	if _, err := schemaLoader.Compile(rootLoader); err != nil {
		source := doc.Source
		if source == "" {
			source = doc.ID
		}
		return fmt.Errorf("schema некорректна (%s): %v", source, err)
	}

	return nil
}
