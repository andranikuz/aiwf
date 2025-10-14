package sdk

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateSDK(t *testing.T) {
	dir := t.TempDir()
	registry := filepath.Join(dir, "registry")
	if err := os.MkdirAll(registry, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	mustWrite(t, filepath.Join(registry, "schema.json"), `{
  "type": "object",
  "properties": {
    "prompt": {"type": "string"},
    "maxTokens": {"type": "integer"}
  }
}`)

	yamlPath := filepath.Join(dir, "spec.yaml")
	mustWrite(t, yamlPath, sampleYAML(registry))

	cmd := NewCommand()
	outDir := filepath.Join(dir, "sdk")
	cmd.SetArgs([]string{"--file", yamlPath, "--out", outDir, "--package", "generated"})
	cmd.SetOut(newBuffer())
	cmd.SetErr(newBuffer())

	if err := cmd.Execute(); err != nil {
		t.Fatalf("sdk command: %v", err)
	}

    for _, name := range []string{"service.go", "agents.go", "workflows.go", "contracts.go"} {
        content, err := os.ReadFile(filepath.Join(outDir, name))
        if err != nil {
            t.Fatalf("read %s: %v", name, err)
        }
        if !strings.Contains(string(content), "package generated") {
            t.Fatalf("unexpected package in %s: %s", name, string(content))
        }
    }

    contracts, err := os.ReadFile(filepath.Join(outDir, "contracts.go"))
    if err != nil {
        t.Fatalf("read contracts: %v", err)
    }
    if !strings.Contains(string(contracts), "type WriterOutput struct") {
        t.Fatalf("expected struct-based contracts, got %s", string(contracts))
    }
}

func sampleYAML(registry string) string {
	return `
version: 0.2
schema_registry:
  root: ` + registry + `
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
`
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
