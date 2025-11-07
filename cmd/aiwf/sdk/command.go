package sdk

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/andranikuz/aiwf/cmd/aiwf/validate"
	backendgo "github.com/andranikuz/aiwf/generator/backend-go"
	"github.com/andranikuz/aiwf/generator/core"
	"github.com/spf13/cobra"
)

// NewCommand возвращает `aiwf sdk` для генерации SDK.
func NewCommand() *cobra.Command {
	var (
		file   string
		outDir string
		pkg    string
	)

	cmd := &cobra.Command{
		Use:   "sdk",
		Short: "Генерирует Go SDK на основе YAML-описания",
		RunE: func(cmd *cobra.Command, args []string) error {
			if file == "" {
				err := errors.New("нужно указать --file")
				fmt.Fprintln(cmd.ErrOrStderr(), err)
				return err
			}
			if outDir == "" {
				err := errors.New("нужно указать --out")
				fmt.Fprintln(cmd.ErrOrStderr(), err)
				return err
			}

			if err := GenerateSDK(file, outDir, pkg); err != nil {
				fmt.Fprintln(cmd.ErrOrStderr(), validate.FormatError(err))
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "✓ SDK сгенерирован в %s\n", outDir)
			return nil
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Путь к YAML-конфигурации")
	cmd.Flags().StringVarP(&outDir, "out", "o", "", "Каталог для вывода SDK")
	cmd.Flags().StringVar(&pkg, "package", "aiwfgen", "Имя Go-пакета")

	return cmd
}

// GenerateSDK генерирует SDK из конфигурации (публичная функция для serve команды)
func GenerateSDK(file, outDir, pkg string) error {
	return GenerateSDKWithOptions(file, outDir, pkg, false)
}

// GenerateSDKWithOptions генерирует SDK с дополнительными опциями
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
