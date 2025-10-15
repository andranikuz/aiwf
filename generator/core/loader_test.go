package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSpecSuccess(t *testing.T) {
	dir := t.TempDir()
	registry := filepath.Join(dir, "registry")
	if err := os.MkdirAll(registry, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	writeFile(t, filepath.Join(registry, "assistant_output.json"), "{}")
	writeFile(t, filepath.Join(registry, "assistant_input.json"), "{}")

	yamlPath := filepath.Join(dir, "workflow.yaml")
	writeFile(t, yamlPath, `
version: 0.2
schema_registry:
  root: registry
assistants:
  summarizer:
    model: gpt-4.1-mini
    system_prompt: prompts/summarizer.md
    input_schema_ref: assistant_input.json
    output_schema_ref: assistant_output.json
workflows:
  story:
    description: Test workflow
    dag:
      - step: summarize
        assistant: summarizer
`)

	spec, err := LoadSpec(yamlPath)
	if err != nil {
		t.Fatalf("LoadSpec: %v", err)
	}

	assistant, ok := spec.Assistants["summarizer"]
	if !ok {
		t.Fatalf("assistant summarizer not found")
	}

	expectedOutput := filepath.Join(registry, "assistant_output.json")
	if assistant.Resolved.OutputSchemaPath != expectedOutput {
		t.Fatalf("unexpected output path: %s", assistant.Resolved.OutputSchemaPath)
	}
}

func TestLoadSpecWithImports(t *testing.T) {
	dir := t.TempDir()
	typesPath := filepath.Join(dir, "book.types.yaml")
	writeFile(t, typesPath, `types:
  PremiseOutput:
    $id: "aiwf://book/PremiseOutput"
    type: object
    properties:
      title: { type: string }
`)

	yamlPath := filepath.Join(dir, "workflow.yaml")
	writeFile(t, yamlPath, `
version: 0.2
imports:
  - as: book
    path: ./book.types.yaml
assistants:
  premise:
    model: gpt-4o-mini
    output_schema_ref: "aiwf://book/PremiseOutput"
workflows:
  wf:
    dag:
      - step: premise
        assistant: premise
`)

	spec, err := LoadSpec(yamlPath)
	if err != nil {
		t.Fatalf("LoadSpec: %v", err)
	}

	assistant := spec.Assistants["premise"]
	if assistant.Resolved.OutputSchema == nil {
		t.Fatalf("expected resolved schema")
	}
	if assistant.Resolved.OutputSchemaPath != typesPath {
		t.Fatalf("unexpected schema source: %s", assistant.Resolved.OutputSchemaPath)
	}
	if len(assistant.Resolved.OutputSchema.Data) == 0 {
		t.Fatalf("schema data is empty")
	}

	if _, ok := spec.Resolved.TypeRegistry["aiwf://book/PremiseOutput"]; !ok {
		t.Fatalf("type registry missing entry")
	}
}

func TestLoadSpecMissingSchema(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "workflow.yaml")
	writeFile(t, yamlPath, `
version: 0.2
schema_registry:
  root: registry
assistants:
  summarizer:
    model: gpt-4.1-mini
    output_schema_ref: missing.json
workflows:
  story:
    description: Test workflow
    dag:
      - step: summarize
        assistant: summarizer
`)

	_, err := LoadSpec(yamlPath)
	if err == nil {
		t.Fatalf("expected error for missing schema")
	}
}

func TestLoadSpecInvalidSchema(t *testing.T) {
	dir := t.TempDir()
	registry := filepath.Join(dir, "registry")
	if err := os.MkdirAll(registry, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	writeFile(t, filepath.Join(registry, "bad.json"), "not-json")

	yamlPath := filepath.Join(dir, "workflow.yaml")
	writeFile(t, yamlPath, `
version: 0.2
schema_registry:
  root: registry
assistants:
  bad:
    model: gpt-4.1
    output_schema_ref: bad.json
workflows:
  wf:
    description: bad schema
    dag:
      - step: s
        assistant: bad
`)

	_, err := LoadSpec(yamlPath)
	if err == nil {
		t.Fatalf("expected schema validation error")
	}
}

func TestWorkflowAssistantValidation(t *testing.T) {
	dir := t.TempDir()
	registry := filepath.Join(dir, "registry")
	if err := os.MkdirAll(registry, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	writeFile(t, filepath.Join(registry, "assistant_output.json"), "{}")

	yamlPath := filepath.Join(dir, "workflow.yaml")
	writeFile(t, yamlPath, `
version: 0.2
schema_registry:
  root: registry
assistants:
  summarizer:
    model: gpt-4.1-mini
    output_schema_ref: assistant_output.json
workflows:
  story:
    description: Test workflow
    dag:
      - step: summarize
        assistant: missing
`)

	_, err := LoadSpec(yamlPath)
	if err == nil {
		t.Fatalf("expected error for missing assistant")
	}
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}
