package validate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andranikuz/aiwf/generator/core"
)

func TestValidateSuccess(t *testing.T) {
	dir := t.TempDir()
	registry := filepath.Join(dir, "registry")
	if err := os.MkdirAll(registry, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	mustWrite(t, filepath.Join(registry, "output.json"), "{}")

	yamlPath := filepath.Join(dir, "spec.yaml")
	mustWrite(t, yamlPath, `
version: 0.2
schema_registry:
  root: registry
assistants:
  writer:
    model: gpt-4
    output_schema_ref: output.json
workflows:
  draft:
    description: test
    dag:
      - step: run
        assistant: writer
`)

	cmd := NewCommand()
	buf := newBuffer()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--file", yamlPath})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	if got := buf.String(); got == "" {
		t.Fatalf("expected output, got empty")
	}
}

func TestValidateMissingFile(t *testing.T) {
	cmd := NewCommand()
	buf := newBuffer()
	cmd.SetErr(buf)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(buf.String(), "✗") {
		t.Fatalf("expected formatted error, got %s", buf.String())
	}
}

func TestValidateAggregatesErrors(t *testing.T) {
	dir := t.TempDir()
	registry := filepath.Join(dir, "registry")
	if err := os.MkdirAll(registry, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	mustWrite(t, filepath.Join(registry, "schema.json"), "{}")

	yamlPath := filepath.Join(dir, "spec.yaml")
	mustWrite(t, yamlPath, `
version: 0.2
schema_registry:
  root: registry
assistants:
  writer:
    model: gpt-4
    output_schema_ref: schema.json
workflows:
  novel:
    description: test
    dag:
      - step: draft
        assistant: writer
      - step: draft
        assistant: writer
      - step: review
        assistant: writer
        scatter:
          as: reviewer
          concurrency: -1
`)

	spec, err := core.LoadSpec(yamlPath)
	if err != nil {
		t.Fatalf("LoadSpec: %v", err)
	}
	_, _ = core.BuildIR(spec)

	cmd := NewCommand()
	out := newBuffer()
	errBuf := newBuffer()
	cmd.SetOut(out)
	cmd.SetErr(errBuf)
	cmd.SetArgs([]string{"--file", yamlPath})

	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected validation error")
	}

	lines := strings.Split(strings.TrimSpace(errBuf.String()), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected multiple errors, got %s", errBuf.String())
	}
	if !containsSubstring(lines, "dag[1].step") {
		t.Fatalf("missing duplicate step error: %v", lines)
	}
	if !containsSubstring(lines, "dag[2].scatter.from") {
		t.Fatalf("missing scatter from error: %v", lines)
	}
	if !containsSubstring(lines, "dag[2].scatter.concurrency") {
		t.Fatalf("missing scatter concurrency error: %v", lines)
	}
}

func TestValidateWarnings(t *testing.T) {
	dir := t.TempDir()
	registry := filepath.Join(dir, "registry")
	if err := os.MkdirAll(registry, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	mustWrite(t, filepath.Join(registry, "schema.json"), "{}")

	yamlPath := filepath.Join(dir, "spec.yaml")
	mustWrite(t, yamlPath, `
version: 0.2
schema_registry:
  root: registry
assistants:
  writer:
    model: gpt-4
    output_schema_ref: schema.json
workflows:
  empty:
    description: empty workflow
`)

	cmd := NewCommand()
	outBuf := newBuffer()
	errBuf := newBuffer()
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)
	cmd.SetArgs([]string{"--file", yamlPath})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected success with warnings, got %v", err)
	}

	if !strings.Contains(errBuf.String(), "⚠ workflows.empty") {
		t.Fatalf("expected warning output, got %s", errBuf.String())
	}
	if !strings.Contains(outBuf.String(), "✓ YAML валиден") {
		t.Fatalf("expected success message, got %s", outBuf.String())
	}
}

func mustWrite(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

type buffer struct {
	data []byte
}

func newBuffer() *buffer { return &buffer{} }

func (b *buffer) Write(p []byte) (int, error) {
	b.data = append(b.data, p...)
	return len(p), nil
}

func (b *buffer) String() string { return string(b.data) }

func containsSubstring(lines []string, needle string) bool {
	for _, line := range lines {
		if strings.Contains(line, needle) {
			return true
		}
	}
	return false
}
