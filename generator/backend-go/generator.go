package backendgo

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/andranikuz/aiwf/generator/core"
)

// Options описывает параметры генерации Go SDK.
type Options struct {
	Package   string
	OutputDir string
}

// Generate создает Go SDK на основе IR.
func Generate(ir *core.IR, opts Options) (map[string][]byte, error) {
	if ir == nil {
		return nil, fmt.Errorf("backend-go: ir is nil")
	}
	if opts.Package == "" {
		opts.Package = "aiwfgen"
	}
	if opts.OutputDir == "" {
		opts.OutputDir = "."
	}

	files := make(map[string][]byte)

	// Generate types.go
	if ir.Types != nil && len(ir.Types.Types) > 0 {
		typesGen := NewTypesGenerator(ir)
		typesCode, err := typesGen.Generate(opts.Package)
		if err != nil {
			return nil, fmt.Errorf("failed to generate types: %w", err)
		}
		files[filepath.Join(opts.OutputDir, "types.go")] = []byte(typesCode)
	}

	// Generate agents.go
	if len(ir.Assistants) > 0 {
		agentsGen := NewAgentsGenerator(ir)
		agentsCode, err := agentsGen.Generate(opts.Package)
		if err != nil {
			return nil, fmt.Errorf("failed to generate agents: %w", err)
		}
		files[filepath.Join(opts.OutputDir, "agents.go")] = []byte(agentsCode)
	}

	// Generate service.go
	serviceGen := NewServiceGenerator(ir)
	serviceCode, err := serviceGen.Generate(opts.Package)
	if err != nil {
		return nil, fmt.Errorf("failed to generate service: %w", err)
	}
	files[filepath.Join(opts.OutputDir, "service.go")] = []byte(serviceCode)

	// Don't generate go.mod - SDK should be part of the user's project

	return files, nil
}


// WriteFiles writes generated files to disk
func WriteFiles(files map[string][]byte) error {
	for path, content := range files {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		if err := os.WriteFile(path, content, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", path, err)
		}
	}
	return nil
}
