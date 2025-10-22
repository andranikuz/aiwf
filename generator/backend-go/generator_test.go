package backendgo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/andranikuz/aiwf/generator/core"
)

func TestGenerateService(t *testing.T) {
	writerInput := filepath.Join("testdata", "writer_input.json")
	writerOutput := filepath.Join("testdata", "writer_output.json")
	criticInput := filepath.Join("testdata", "critic_input.json")
	criticOutput := filepath.Join("testdata", "critic_output.json")

	ir := &core.IR{
		Assistants: map[string]core.IRAssistant{
			"writer": {
				Name:             "writer",
				Model:            "gpt-4",
				SystemPrompt:     "You are a writing assistant",
				InputSchemaRef:   "writer_input.json",
				OutputSchemaRef:  "writer_output.json",
				InputSchemaPath:  writerInput,
				OutputSchemaPath: writerOutput,
				InputSchemaData:  mustRead(t, writerInput),
				OutputSchemaData: mustRead(t, writerOutput),
			},
			"critic": {
				Name:             "critic",
				Model:            "gpt-4-turbo",
				SystemPrompt:     "Provide critique",
				InputSchemaRef:   "critic_input.json",
				OutputSchemaRef:  "critic_output.json",
				InputSchemaPath:  criticInput,
				OutputSchemaPath: criticOutput,
				InputSchemaData:  mustRead(t, criticInput),
				OutputSchemaData: mustRead(t, criticOutput),
			},
		},
		Workflows: map[string]core.IRWorkflow{
			"novel": {
				Name: "novel",
				Steps: []core.IRStep{
					{Name: "draft", Assistant: "writer"},
					{Name: "critique", Assistant: "critic", Needs: []string{"draft"}},
				},
			},
		},
		TypeRegistry: map[string][]byte{
			"aiwf://common/Tone": []byte(`{"type":"string","enum":["dark","hopeful","playful"]}`),
		},
	}

	files, err := Generate(ir, Options{Package: "generated", OutputDir: "sdk"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	goldens := map[string]string{
		"sdk/service.go":   "service.golden",
		"sdk/agents.go":    "agents.golden",
		"sdk/workflows.go": "workflows.golden",
		"sdk/dialog.go":    "dialog.golden",
		"sdk/contracts.go": "contracts.golden",
	}

	for path, goldenName := range goldens {
		content, ok := files[path]
		if !ok {
			t.Fatalf("expected generated file %s", path)
		}

		goldenPath := filepath.Join("testdata", goldenName)
		if os.Getenv("UPDATE_GOLDEN") == "1" {
			if err := os.WriteFile(goldenPath, content, 0o644); err != nil {
				t.Fatalf("write golden: %v", err)
			}
		}

		golden, err := os.ReadFile(goldenPath)
		if err != nil {
			t.Fatalf("read golden: %v", err)
		}
		if string(content) != string(golden) {
			t.Fatalf("snapshot mismatch for %s:\n%s", goldenName, string(content))
		}
	}
}

func TestGenerateWithoutWorkflows(t *testing.T) {
	writerOutput := filepath.Join("testdata", "writer_output.json")

	ir := &core.IR{
		Assistants: map[string]core.IRAssistant{
			"writer": {
				Name:             "writer",
				Model:            "gpt-4",
				OutputSchemaRef:  "writer_output.json",
				OutputSchemaPath: writerOutput,
				OutputSchemaData: mustRead(t, writerOutput),
			},
		},
		Workflows: map[string]core.IRWorkflow{},
	}

	files, err := Generate(ir, Options{Package: "generated", OutputDir: "sdk"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if _, ok := files[filepath.Join("sdk", "workflows.go")]; ok {
		t.Fatalf("workflows.go should not be generated when workflows are absent")
	}

	if _, ok := files[filepath.Join("sdk", "service.go")]; !ok {
		t.Fatalf("service.go not generated")
	}
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file %s: %v", path, err)
	}
	return data
}
