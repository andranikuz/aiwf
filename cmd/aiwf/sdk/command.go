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
			if pkg == "" {
				pkg = "aiwfgen"
			}

			spec, err := core.LoadSpec(file)
			if err != nil {
				fmt.Fprintln(cmd.ErrOrStderr(), validate.FormatError(err))
				if _, ok := err.(*core.MultiError); ok {
					return err
				}
				return err
			}

			ir, err := core.BuildIR(spec)
			if err != nil {
				fmt.Fprintln(cmd.ErrOrStderr(), validate.FormatError(err))
				if me, ok := err.(*core.MultiError); ok && !me.HasErrors() {
					// только предупреждения — продолжаем.
				} else {
					return err
				}
			}

			files, err := backendgo.Generate(ir, backendgo.Options{Package: pkg, OutputDir: outDir})
			if err != nil {
				return err
			}

			for path, data := range files {
				if err := writeFile(path, data); err != nil {
					return err
				}
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

func writeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create dirs: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}
