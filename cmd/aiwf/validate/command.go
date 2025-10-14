package validate

import (
	"errors"
	"fmt"

	"github.com/andranikuz/aiwf/generator/core"
	"github.com/spf13/cobra"
)

// NewCommand возвращает подкоманду `aiwf validate`.
func NewCommand() *cobra.Command {
	var inputPath string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Проверяет YAML-конфигурацию AIWF",
		RunE: func(cmd *cobra.Command, args []string) error {
            if inputPath == "" {
                err := errors.New("нужно указать путь к YAML через --file")
                fmt.Fprintln(cmd.ErrOrStderr(), FormatError(err))
                return err
            }

            spec, err := core.LoadSpec(inputPath)
            if err != nil {
                fmt.Fprintln(cmd.ErrOrStderr(), FormatError(err))
                return err
            }

            ir, err := core.BuildIR(spec)
            if err != nil {
                fmt.Fprintln(cmd.ErrOrStderr(), FormatError(err))
                if me, ok := err.(*core.MultiError); ok && !me.HasErrors() {
                    // Только предупреждения: продолжаем и считаем валидацию успешной.
                } else {
					return err
				}
			}

			fmt.Fprintf(cmd.OutOrStdout(), "✓ YAML валиден. Assistants: %d, Workflows: %d\n", len(ir.Assistants), len(ir.Workflows))
			return nil
		},
	}

	cmd.Flags().StringVarP(&inputPath, "file", "f", "", "Путь к YAML-конфигурации (обязательно)")

	return cmd
}
