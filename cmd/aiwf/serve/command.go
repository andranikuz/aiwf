package serve

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/andranikuz/aiwf/cmd/aiwf/sdk"
	"github.com/spf13/cobra"
)

// ServeOptions содержит параметры команды serve
type ServeOptions struct {
	ConfigPath string
	Output     string // Если указан - файлы сохраняются для дебага
	Port       int
	Host       string
}

// NewCommand создает команду serve
func NewCommand() *cobra.Command {
	opts := &ServeOptions{}

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start HTTP server for agents (ephemeral mode)",
		Long: `Start HTTP server that exposes agents as REST API endpoints.

By default, SDK is generated in a temporary directory and cleaned up on exit.
Use --output to persist generated files for debugging.

Examples:
  # Quick start (ephemeral mode)
  aiwf serve -f config.yaml

  # With custom port
  aiwf serve -f config.yaml --port 3000

  # Persist generated SDK for debugging
  aiwf serve -f config.yaml --output ./generated

  # Custom host binding
  aiwf serve -f config.yaml --host 0.0.0.0 --port 8080
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServe(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.ConfigPath, "file", "f", "config.yaml", "Path to YAML config file")
	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output directory (persistent mode, keeps files after exit)")
	cmd.Flags().IntVarP(&opts.Port, "port", "p", 8080, "Server port")
	cmd.Flags().StringVar(&opts.Host, "host", "127.0.0.1", "Server host")

	return cmd
}

func runServe(opts *ServeOptions) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Определяем output директорию
	outputDir, shouldCleanup, err := determineOutputDir(opts)
	if err != nil {
		return fmt.Errorf("failed to determine output directory: %w", err)
	}

	// Cleanup при выходе (если ephemeral mode)
	if shouldCleanup {
		defer func() {
			fmt.Printf("🧹 Cleaning up %s...\n", outputDir)
			if err := os.RemoveAll(outputDir); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to cleanup: %v\n", err)
			}
		}()
	} else {
		fmt.Printf("📁 Generated files will be kept in: %s\n", outputDir)
	}

	// Генерируем SDK
	fmt.Printf("🔨 Generating SDK in %s...\n", outputDir)
	if err := generateSDK(opts.ConfigPath, outputDir); err != nil {
		return fmt.Errorf("SDK generation failed: %w", err)
	}

	// Компилируем сервер
	fmt.Println("🔧 Building server...")
	binaryPath := filepath.Join(outputDir, "aiwf-server")
	if err := buildServer(outputDir, binaryPath); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	// Запускаем сервер
	fmt.Printf("🚀 Starting server on %s:%d...\n", opts.Host, opts.Port)
	fmt.Println("📡 Press Ctrl+C to stop")
	return runServer(ctx, binaryPath, opts)
}

// determineOutputDir определяет директорию для генерации и нужна ли очистка
func determineOutputDir(opts *ServeOptions) (outputDir string, shouldCleanup bool, err error) {
	if opts.Output != "" {
		// Persistent mode - используем указанную директорию
		absPath, err := filepath.Abs(opts.Output)
		if err != nil {
			return "", false, err
		}
		return absPath, false, nil
	}

	// Ephemeral mode - создаем временную директорию
	hash := computeConfigHash(opts.ConfigPath)
	tempDir, err := os.MkdirTemp("", fmt.Sprintf("aiwf-serve-%s-*", hash[:8]))
	if err != nil {
		return "", false, err
	}

	return tempDir, true, nil
}

// generateSDK генерирует SDK используя существующую логику
func generateSDK(configPath, outputDir string) error {
	// Используем существующую команду sdk с генерацией сервера
	// Package name = "sdk" чтобы можно было импортировать из cmd/server
	return sdk.GenerateSDKWithOptions(configPath, outputDir, "sdk", true)
}

// buildServer компилирует server binary
func buildServer(sourceDir, outputBinary string) error {
	// Проверяем наличие cmd/server/main.go
	serverMainPath := filepath.Join(sourceDir, "cmd", "server", "main.go")
	if _, err := os.Stat(serverMainPath); os.IsNotExist(err) {
		return fmt.Errorf("server main.go not found at %s (ensure server generation is enabled)", serverMainPath)
	}

	// Сначала запускаем go mod tidy
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = sourceDir
	tidyCmd.Stdout = os.Stdout
	tidyCmd.Stderr = os.Stderr
	if err := tidyCmd.Run(); err != nil {
		return fmt.Errorf("go mod tidy failed: %w", err)
	}

	// Теперь компилируем
	cmd := exec.Command("go", "build",
		"-o", outputBinary,
		"./cmd/server")

	cmd.Dir = sourceDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0") // Static binary

	return cmd.Run()
}

// runServer запускает скомпилированный сервер
func runServer(ctx context.Context, binaryPath string, opts *ServeOptions) error {
	cmd := exec.CommandContext(ctx, binaryPath)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PORT=%d", opts.Port),
		fmt.Sprintf("HOST=%s", opts.Host),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	// Ждем сигнала
	go func() {
		<-sigChan
		fmt.Println("\n🛑 Shutting down server...")
		// Даем серверу время на graceful shutdown
		time.Sleep(100 * time.Millisecond)
		if cmd.Process != nil {
			cmd.Process.Signal(syscall.SIGTERM)
		}
	}()

	err := cmd.Wait()
	if err != nil {
		// Игнорируем ошибку от SIGTERM/SIGINT
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 130 || exitErr.ExitCode() == 143 {
				return nil
			}
		}
		return fmt.Errorf("server exited with error: %w", err)
	}

	return nil
}

// computeConfigHash вычисляет хеш конфига для уникальной директории
func computeConfigHash(configPath string) string {
	data, err := os.ReadFile(configPath)
	if err != nil {
		// Fallback к имени файла
		return filepath.Base(configPath)
	}
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)[:12]
}
