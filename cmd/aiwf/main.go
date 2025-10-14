package main

import (
	"fmt"
	"os"

	"github.com/andranikuz/aiwf/cmd/aiwf/sdk"
	"github.com/andranikuz/aiwf/cmd/aiwf/validate"
	"github.com/spf13/cobra"
)

// globalFlags описывает глобальные параметры CLI.
type globalFlags struct {
	configPath  string
	environment string
	verbose     bool
}

func newRootCmd() *cobra.Command {
	flags := &globalFlags{}

	cmd := &cobra.Command{
		Use:   "aiwf",
		Short: "AI Workflow (AIWF) CLI",
		Long:  "Инструмент для валидации конфигураций, генерации SDK и разработки провайдеров AIWF.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Пока нет подпроцессов — выводим состояние глобальных флагов, чтобы подтвердить их обработку.
			message := "aiwf CLI пока находится в разработке\n"
			message += fmt.Sprintf("config: %s\n", flags.configPath)
			message += fmt.Sprintf("env: %s\n", flags.environment)
			message += fmt.Sprintf("verbose: %t\n", flags.verbose)
			cmd.Println(message)
			return nil
		},
	}

	cmd.PersistentFlags().StringVar(&flags.configPath, "config", "config.yaml", "Путь к YAML-конфигурации проекта")
	cmd.PersistentFlags().StringVar(&flags.environment, "env", "local", "Целевое окружение (local, staging, prod)")
	cmd.PersistentFlags().BoolVarP(&flags.verbose, "verbose", "v", false, "Расширенный лог вывода")

	cmd.AddCommand(validate.NewCommand())
	cmd.AddCommand(sdk.NewCommand())

	return cmd
}

func main() {
	rootCmd := newRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "aiwf: %v\n", err)
		os.Exit(1)
	}
}
