package backendgo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/andranikuz/aiwf/generator/core"
)

func TestGenerateService(t *testing.T) {
	ir := &core.IR{
		Assistants: map[string]core.IRAssistant{
			"writer": {
				Name:            "writer",
				Model:           "gpt-4",
				SystemPrompt:    "You are a writing assistant",
				InputSchemaRef:  "writer_input.json",
				OutputSchemaRef: "writer_output.json",
				InputSchemaPath:  filepath.Join("testdata", "writer_input.json"),
				OutputSchemaPath: filepath.Join("testdata", "writer_output.json"),
			},
			"critic": {
				Name:            "critic",
				Model:           "gpt-4-turbo",
				SystemPrompt:    "Provide critique",
				InputSchemaRef:  "critic_input.json",
				OutputSchemaRef: "critic_output.json",
				InputSchemaPath:  filepath.Join("testdata", "critic_input.json"),
				OutputSchemaPath: filepath.Join("testdata", "critic_output.json"),
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
	}

	files, err := Generate(ir, Options{Package: "generated", OutputDir: "sdk"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

    goldens := map[string]string{
        "sdk/service.go":     "service.golden",
        "sdk/agents.go":      "agents.golden",
        "sdk/workflows.go":   "workflows.golden",
        "sdk/contracts.go":   "contracts.golden",
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
