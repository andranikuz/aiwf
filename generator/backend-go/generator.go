package backendgo

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/andranikuz/aiwf/generator/core"
)

// Options описывает параметры генерации Go SDK.
type Options struct {
	Package        string
	OutputDir      string
	GenerateServer bool // Генерировать ли HTTP сервер
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

	// Determine SDK directory based on whether server is being generated
	sdkDir := opts.OutputDir
	if opts.GenerateServer {
		// When generating server, place SDK files in sdk/ subdirectory
		sdkDir = filepath.Join(opts.OutputDir, opts.Package)
	}

	// Generate types.go
	if ir.Types != nil && len(ir.Types.Types) > 0 {
		typesGen := NewTypesGenerator(ir)
		typesCode, err := typesGen.Generate(opts.Package)
		if err != nil {
			return nil, fmt.Errorf("failed to generate types: %w", err)
		}
		files[filepath.Join(sdkDir, "types.go")] = []byte(typesCode)
	}

	// Generate agents.go
	if len(ir.Assistants) > 0 {
		agentsGen := NewAgentsGenerator(ir)
		agentsCode, err := agentsGen.Generate(opts.Package)
		if err != nil {
			return nil, fmt.Errorf("failed to generate agents: %w", err)
		}
		files[filepath.Join(sdkDir, "agents.go")] = []byte(agentsCode)
	}

	// Generate service.go
	serviceGen := NewServiceGenerator(ir)
	serviceCode, err := serviceGen.Generate(opts.Package)
	if err != nil {
		return nil, fmt.Errorf("failed to generate service: %w", err)
	}
	files[filepath.Join(sdkDir, "service.go")] = []byte(serviceCode)

	// Generate HTTP server if requested
	if opts.GenerateServer {
		serverGen := NewServerGenerator(ir)
		serverCode, err := serverGen.Generate(opts.Package)
		if err != nil {
			return nil, fmt.Errorf("failed to generate server: %w", err)
		}
		files[filepath.Join(opts.OutputDir, "cmd", "server", "main.go")] = []byte(serverCode)

		// Generate go.mod for server
		goModContent := generateGoMod(opts.Package)
		files[filepath.Join(opts.OutputDir, "go.mod")] = []byte(goModContent)
	}

	return files, nil
}

// generateGoMod создает go.mod файл для сервера
func generateGoMod(packageName string) string {
	// Get current working directory to find aiwf module
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	// Use unique module name to avoid conflicts
	// Package name is used for SDK files (e.g., "sdk")
	// Module name should be unique (e.g., "aiwf-server")
	moduleName := "aiwf-server"

	// Use replace directive to use local aiwf module
	return fmt.Sprintf(`module %s

go 1.24

require (
	github.com/andranikuz/aiwf v0.0.0
)

replace github.com/andranikuz/aiwf => %s
`, moduleName, cwd)
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
