package sdk

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/andranikuz/aiwf/cmd/aiwf/validate"
	backendgo "github.com/andranikuz/aiwf/generator/backend-go"
	clientphp "github.com/andranikuz/aiwf/generator/client-php"
	"github.com/andranikuz/aiwf/generator/core"
	"github.com/spf13/cobra"
)

// SDKOptions содержит параметры команды sdk
type SDKOptions struct {
	File    string
	OutDir  string
	Package string
	Lang    string
	Type    string // "full" или "client"
	BaseURL string // для HTTP клиентов
}

// NewCommand возвращает улучшенную команду `aiwf sdk`
func NewCommand() *cobra.Command {
	opts := &SDKOptions{}

	cmd := &cobra.Command{
		Use:   "sdk",
		Short: "Generate SDK from YAML config",
		Long: `Generate SDK code from YAML specification.

Two types of SDK generation:

1. Full SDK (--type full, default for Go)
   - Embeds runtime and provider logic
   - Works standalone without HTTP server
   - Best for: edge computing, CLI tools, embedded systems
   - Languages: go

2. HTTP Client (--type client, default for other languages)
   - Lightweight HTTP wrapper
   - Connects to 'aiwf serve' server
   - Best for: web apps, microservices, most use cases
   - Languages: python, typescript, go, ruby, php, java, csharp

The type is auto-detected based on language but can be overridden.

Examples:
  # Full Go SDK (auto-detected, current behavior)
  aiwf sdk -f config.yaml -o ./sdk

  # Python HTTP client (auto-detected)
  aiwf sdk -f config.yaml -l python -o ./client.py

  # TypeScript HTTP client
  aiwf sdk -f config.yaml -l typescript -o ./client.ts

  # Force HTTP client for Go
  aiwf sdk -f config.yaml -l go --type client -o ./client.go

  # Specify base URL for client
  aiwf sdk -f config.yaml -l typescript --base-url https://api.example.com
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSDK(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.File, "file", "f", "", "Path to YAML config file")
	cmd.Flags().StringVarP(&opts.OutDir, "out", "o", "", "Output directory")
	cmd.Flags().StringVarP(&opts.Lang, "lang", "l", "go", "Target language (go, php)")
	cmd.Flags().StringVar(&opts.Package, "package", "aiwfgen", "Package/module name")
	cmd.Flags().StringVar(&opts.Type, "type", "", "SDK type: 'full' or 'client' (auto-detected if not specified)")
	cmd.Flags().StringVar(&opts.BaseURL, "base-url", "http://127.0.0.1:8080", "Base URL for HTTP client")

	return cmd
}

func runSDK(opts *SDKOptions) error {
	// Validate inputs
	if opts.File == "" {
		return errors.New("--file is required")
	}
	if opts.OutDir == "" {
		return errors.New("--out is required")
	}

	// Auto-detect SDK type if not specified
	if opts.Type == "" {
		opts.Type = detectSDKType(opts.Lang)
	}

	// Validate SDK type
	if opts.Type != "full" && opts.Type != "client" {
		return fmt.Errorf("invalid --type: %s (must be 'full' or 'client')", opts.Type)
	}

	// Load config
	spec, err := core.LoadSpec(opts.File)
	if err != nil {
		fmt.Fprintln(os.Stderr, validate.FormatError(err))
		return err
	}

	ir, err := core.BuildIR(spec)
	if err != nil {
		fmt.Fprintln(os.Stderr, validate.FormatError(err))
		if me, ok := err.(*core.MultiError); ok && !me.HasErrors() {
			// только предупреждения — продолжаем.
		} else {
			return err
		}
	}

	// Generate based on type and language
	if opts.Type == "full" {
		return generateFullSDK(ir, opts)
	} else {
		return generateHTTPClient(ir, opts)
	}
}

// detectSDKType определяет тип SDK на основе языка
func detectSDKType(lang string) string {
	switch lang {
	case "go":
		return "full" // Go имеет runtime - можем делать full SDK
	default:
		return "client" // Остальные языки - только HTTP client
	}
}

// generateFullSDK генерирует полноценный SDK (только для Go сейчас)
func generateFullSDK(ir *core.IR, opts *SDKOptions) error {
	if opts.Lang != "go" {
		return fmt.Errorf("full SDK is only supported for Go (got: %s)", opts.Lang)
	}

	pkg := opts.Package
	if pkg == "" {
		pkg = "aiwfgen"
	}

	files, err := backendgo.Generate(ir, backendgo.Options{
		Package:        pkg,
		OutputDir:      opts.OutDir,
		GenerateServer: false,
	})
	if err != nil {
		return err
	}

	for path, data := range files {
		if err := writeFile(path, data); err != nil {
			return err
		}
	}

	fmt.Printf("✓ Full Go SDK generated in %s\n", opts.OutDir)
	return nil
}

// generateHTTPClient генерирует легковесный HTTP клиент
func generateHTTPClient(ir *core.IR, opts *SDKOptions) error {
	var content string
	var err error
	var outputFile string

	switch opts.Lang {
	case "python":
		content, err = generatePythonClient(ir, opts)
		outputFile = filepath.Join(opts.OutDir, "client.py")
	case "typescript":
		content, err = generateTypeScriptClient(ir, opts)
		outputFile = filepath.Join(opts.OutDir, "client.ts")
	case "go":
		content, err = generateGoHTTPClient(ir, opts)
		outputFile = filepath.Join(opts.OutDir, "client.go")
	case "ruby":
		content, err = generateRubyClient(ir, opts)
		outputFile = filepath.Join(opts.OutDir, "client.rb")
	case "php":
		content, err = generatePHPClient(ir, opts)
		outputFile = filepath.Join(opts.OutDir, "client.php")
	case "java":
		content, err = generateJavaClient(ir, opts)
		outputFile = filepath.Join(opts.OutDir, "AIWFClient.java")
	case "csharp":
		content, err = generateCSharpClient(ir, opts)
		outputFile = filepath.Join(opts.OutDir, "AIWFClient.cs")
	default:
		return fmt.Errorf("unsupported language: %s", opts.Lang)
	}

	if err != nil {
		return fmt.Errorf("client generation failed: %w", err)
	}

	// Write file
	if err := os.MkdirAll(filepath.Dir(outputFile), 0o755); err != nil {
		return fmt.Errorf("create dirs: %w", err)
	}

	if err := os.WriteFile(outputFile, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	fmt.Printf("✓ HTTP client generated: %s\n", outputFile)
	fmt.Printf("  Language: %s\n", opts.Lang)
	fmt.Printf("  Base URL: %s\n", opts.BaseURL)
	fmt.Printf("\nUsage example:\n")
	printUsageExample(opts.Lang, outputFile)

	return nil
}

func printUsageExample(lang, path string) {
	switch lang {
	case "python":
		fmt.Printf("  python %s\n", path)
	case "typescript":
		fmt.Printf("  ts-node %s\n", path)
	case "go":
		fmt.Printf("  go run %s\n", path)
	case "ruby":
		fmt.Printf("  ruby %s\n", path)
	case "php":
		fmt.Printf("  php %s\n", path)
	}
}

// Placeholder generators - to be implemented
func generatePythonClient(ir *core.IR, opts *SDKOptions) (string, error) {
	// TODO: implement
	return "", fmt.Errorf("Python client generator not implemented yet")
}

func generateTypeScriptClient(ir *core.IR, opts *SDKOptions) (string, error) {
	// TODO: implement
	return "", fmt.Errorf("TypeScript client generator not implemented yet")
}

func generateGoHTTPClient(ir *core.IR, opts *SDKOptions) (string, error) {
	// TODO: implement
	return "", fmt.Errorf("Go HTTP client generator not implemented yet")
}

func generateRubyClient(ir *core.IR, opts *SDKOptions) (string, error) {
	return "", fmt.Errorf("Ruby client generator not implemented yet")
}

func generatePHPClient(ir *core.IR, opts *SDKOptions) (string, error) {
	gen := clientphp.New(ir, opts.BaseURL)
	return gen.Generate()
}

func generateJavaClient(ir *core.IR, opts *SDKOptions) (string, error) {
	return "", fmt.Errorf("Java client generator not implemented yet")
}

func generateCSharpClient(ir *core.IR, opts *SDKOptions) (string, error) {
	return "", fmt.Errorf("C# client generator not implemented yet")
}

func GenerateSDKWithOptions(file, outDir, pkg string, generateServer bool) error {
	if pkg == "" {
		pkg = "aiwfgen"
	}

	spec, err := core.LoadSpec(file)
	if err != nil {
		return err
	}

	ir, err := core.BuildIR(spec)
	if err != nil {
		if me, ok := err.(*core.MultiError); ok && !me.HasErrors() {
			// только предупреждения — продолжаем.
		} else {
			return err
		}
	}

	files, err := backendgo.Generate(ir, backendgo.Options{
		Package:        pkg,
		OutputDir:      outDir,
		GenerateServer: generateServer,
	})
	if err != nil {
		return err
	}

	for path, data := range files {
		if err := writeFile(path, data); err != nil {
			return err
		}
	}

	return nil
}

func writeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create dirs: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}
