package generate

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/andranikuz/aiwf/generator/core"
	"github.com/andranikuz/aiwf/internal/metasdk"
	"github.com/andranikuz/aiwf/providers/anthropic"
	"github.com/andranikuz/aiwf/providers/grok"
	"github.com/andranikuz/aiwf/providers/openai"
	"github.com/andranikuz/aiwf/runtime/go/aiwf"
	"github.com/spf13/cobra"
)

// GenerateOptions содержит параметры команды generate
type GenerateOptions struct {
	Task        string
	TaskFile    string
	Output      string
	Interactive bool
	Provider    string
	APIKey      string
}

// NewCommand создает команду generate
func NewCommand() *cobra.Command {
	opts := &GenerateOptions{}

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate AIWF config from task description using AI",
		Long: `Generate AIWF YAML configuration from natural language task description.

Uses meta-agents to analyze your task and generate production-ready configuration.

Examples:
  # Interactive mode (recommended)
  aiwf generate --interactive

  # Quick generation from command line
  aiwf generate -t "Create a spam filter for emails"

  # From task file
  aiwf generate --task-file task.txt -o spam-filter.yaml

  # Specify provider and API key
  aiwf generate -t "Translate text" --provider openai --api-key sk-...
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerate(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Task, "task", "t", "", "Task description (quoted string)")
	cmd.Flags().StringVar(&opts.TaskFile, "task-file", "", "File containing task description")
	cmd.Flags().StringVarP(&opts.Output, "output", "o", "generated-config.yaml", "Output YAML file")
	cmd.Flags().BoolVarP(&opts.Interactive, "interactive", "i", false, "Interactive mode with prompts and confirmations")
	cmd.Flags().StringVar(&opts.Provider, "provider", "openai", "LLM provider (openai, grok, anthropic)")
	cmd.Flags().StringVar(&opts.APIKey, "api-key", "", "API key (or use environment variable)")

	return cmd
}

func runGenerate(opts *GenerateOptions) error {
	// Create provider
	provider, err := createProvider(opts.Provider, opts.APIKey)
	if err != nil {
		return err
	}

	// Create meta-agent service
	service := metasdk.NewService(provider)

	// Get task description
	taskDesc, err := getTaskDescription(opts)
	if err != nil {
		return err
	}

	if opts.Interactive {
		return runInteractiveGeneration(service, taskDesc, opts.Output)
	}

	return runQuickGeneration(service, taskDesc, opts.Output)
}

// createProvider создает провайдера на основе конфигурации
func createProvider(providerName, apiKey string) (aiwf.ModelClient, error) {
	// Get API key from environment if not provided
	if apiKey == "" {
		switch providerName {
		case "openai":
			apiKey = os.Getenv("OPENAI_API_KEY")
		case "grok":
			apiKey = os.Getenv("GROK_API_KEY")
		case "anthropic":
			apiKey = os.Getenv("ANTHROPIC_API_KEY")
		}
	}

	if apiKey == "" {
		return nil, fmt.Errorf("API key required: set %s_API_KEY environment variable or use --api-key flag",
			strings.ToUpper(providerName))
	}

	// Create provider
	switch providerName {
	case "openai":
		return openai.NewClient(openai.ClientConfig{APIKey: apiKey})
	case "grok":
		return grok.NewClient(grok.ClientConfig{APIKey: apiKey})
	case "anthropic":
		return anthropic.NewClient(anthropic.ClientConfig{APIKey: apiKey})
	default:
		return nil, fmt.Errorf("unknown provider: %s (supported: openai, grok, anthropic)", providerName)
	}
}

// getTaskDescription получает описание задачи из различных источников
func getTaskDescription(opts *GenerateOptions) (string, error) {
	if opts.TaskFile != "" {
		data, err := os.ReadFile(opts.TaskFile)
		if err != nil {
			return "", fmt.Errorf("failed to read task file: %w", err)
		}
		return string(data), nil
	}

	if opts.Task != "" {
		return opts.Task, nil
	}

	if opts.Interactive {
		return "", nil // Will be prompted in interactive mode
	}

	return "", fmt.Errorf("task description required: use -t or --task-file or --interactive")
}

// runInteractiveGeneration запускает интерактивную генерацию с подтверждениями
func runInteractiveGeneration(service *metasdk.Service, initialTask, outputPath string) error {
	ctx := context.Background()
	reader := bufio.NewReader(os.Stdin)

	printHeader("🤖 AIWF Interactive Configuration Generator")

	// Step 1: Get task description
	var taskDesc string
	if initialTask != "" {
		taskDesc = initialTask
		fmt.Printf("\n📝 Task: %s\n", taskDesc)
	} else {
		fmt.Println("\n📝 Describe your task in natural language.")
		fmt.Println("   What do you want your AI agent(s) to do?")
		fmt.Print("\n> ")
		task, _ := reader.ReadString('\n')
		taskDesc = strings.TrimSpace(task)
	}

	if taskDesc == "" {
		return fmt.Errorf("task description cannot be empty")
	}

	// Step 2: Task Analysis
	printSeparator()
	printHeader("📊 Step 1: Task Analysis")
	fmt.Println("\nAnalyzing your task...")

	analysis, trace1, err := service.Agents().TaskAnalyzer.Run(ctx, taskDesc)
	if err != nil {
		return fmt.Errorf("task analysis failed: %w", err)
	}

	fmt.Printf("\n✓ Analysis complete (%.1fs)!\n\n", trace1.Duration.Seconds())
	printAnalysisResult(analysis)

	// Check if clarification needed
	userAnswers := make(map[string]string)
	if analysis.NeedsClarification {
		fmt.Println("\n❓ The agent has some questions to better understand your task:")
		for i, q := range analysis.Questions {
			fmt.Printf("\n%d. %s\n", i+1, q.Question)
			if q.Reason != "" {
				fmt.Printf("   Reason: %s\n", q.Reason)
			}
			if len(q.Suggestions) > 0 {
				fmt.Printf("   Suggestions: %s\n", strings.Join(q.Suggestions, ", "))
			}
			fmt.Print("   Your answer: ")
			answer, _ := reader.ReadString('\n')
			userAnswers[q.Question] = strings.TrimSpace(answer)
		}
	}

	// Step 3: Approval or refinement
	printSeparator()
	fmt.Println("\n✅ Review the analysis above.")
	fmt.Println("\nOptions:")
	fmt.Println("  [c]ontinue - Proceed to YAML generation")
	fmt.Println("  [r]efine   - Add additional instructions")
	fmt.Println("  [q]uit     - Cancel generation")
	fmt.Print("\nYour choice: ")

	choice, _ := reader.ReadString('\n')
	choice = strings.ToLower(strings.TrimSpace(choice))

	var refinements string
	switch choice {
	case "q", "quit":
		fmt.Println("\n❌ Generation cancelled.")
		return nil
	case "r", "refine":
		fmt.Println("\n📝 Enter your refinements (press Enter twice to finish):")
		fmt.Print("> ")
		var lines []string
		for {
			line, _ := reader.ReadString('\n')
			line = strings.TrimRight(line, "\n\r")
			if line == "" && len(lines) > 0 {
				break
			}
			if line != "" {
				lines = append(lines, line)
			}
		}
		refinements = strings.Join(lines, "\n")
		fmt.Printf("\n✓ Refinements recorded\n")
	case "c", "continue", "":
		// Continue
	default:
		fmt.Println("Invalid choice, continuing...")
	}

	// Step 4: YAML Generation
	printSeparator()
	printHeader("⚙️ Step 2: YAML Generation")
	fmt.Println("\nGenerating configuration...")

	// Prepare input
	analysisJSON, _ := json.Marshal(analysis)
	input := metasdk.GenerationInput{
		Analysis:               string(analysisJSON),
		RefinementInstructions: refinements,
		UserAnswers:            userAnswers,
	}

	config, trace2, err := service.Agents().YamlGenerator.Run(ctx, input)
	if err != nil {
		return fmt.Errorf("YAML generation failed: %w", err)
	}

	fmt.Printf("\n✓ Configuration generated (%.1fs)!\n", trace2.Duration.Seconds())

	// Step 5: Review generated YAML
	for {
		printSeparator()
		printHeader("📄 Generated Configuration")
		fmt.Println()
		fmt.Println(config.YamlContent)
		printSeparator()

		fmt.Println("\nOptions:")
		fmt.Println("  [s]ave     - Save to file and exit")
		fmt.Println("  [e]dit     - Request changes to the configuration")
		fmt.Println("  [q]uit     - Cancel without saving")
		fmt.Print("\nYour choice: ")

		choice, _ = reader.ReadString('\n')
		choice = strings.ToLower(strings.TrimSpace(choice))

		switch choice {
		case "q", "quit":
			fmt.Println("\n❌ Generation cancelled. Configuration not saved.")
			return nil

		case "e", "edit":
			fmt.Println("\n📝 What changes would you like? (press Enter twice to finish)")
			fmt.Print("> ")
			var editLines []string
			for {
				line, _ := reader.ReadString('\n')
				line = strings.TrimRight(line, "\n\r")
				if line == "" && len(editLines) > 0 {
					break
				}
				if line != "" {
					editLines = append(editLines, line)
				}
			}
			edits := strings.Join(editLines, "\n")

			fmt.Println("\n⚙️ Regenerating with your changes...")

			// Regenerate with edits
			input.RefinementInstructions = edits
			config, _, err = service.Agents().YamlGenerator.Run(ctx, input)
			if err != nil {
				return fmt.Errorf("regeneration failed: %w", err)
			}

			fmt.Println("\n✓ Updated configuration ready!")
			continue // Show config again

		case "s", "save", "":
			// Continue to save
			break

		default:
			fmt.Println("Invalid choice")
			continue
		}

		break // Exit loop
	}

	// Step 6: Save and validate
	if err := os.WriteFile(outputPath, []byte(config.YamlContent), 0644); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("\n💾 Configuration saved to: %s\n", outputPath)

	// Validate
	fmt.Println("\n✅ Validating configuration...")
	spec, err := core.LoadSpec(outputPath)
	if err != nil {
		fmt.Printf("\n⚠️  Validation warning: %v\n", err)
		fmt.Println("   The configuration was saved but may need manual fixes.")
		return nil
	}

	fmt.Printf("\n✓ Valid configuration!\n")
	fmt.Printf("   - %d types\n", len(spec.Types))
	fmt.Printf("   - %d assistants\n", len(spec.Assistants))

	// Show validation notes if any
	if len(config.ValidationNotes) > 0 {
		fmt.Println("\n📝 Notes:")
		for _, note := range config.ValidationNotes {
			fmt.Printf("   [%s] %s\n", note.Severity, note.Message)
		}
	}

	// Next steps
	printSeparator()
	printHeader("🚀 Next Steps")
	fmt.Println("\n1. Review the generated configuration:")
	fmt.Printf("   cat %s\n\n", outputPath)
	fmt.Println("2. Generate SDK:")
	fmt.Printf("   aiwf sdk -f %s -o ./generated\n\n", outputPath)
	fmt.Println("3. Or start HTTP server:")
	fmt.Printf("   aiwf serve -f %s\n", outputPath)

	return nil
}

// runQuickGeneration выполняет быструю генерацию без интерактивности
func runQuickGeneration(service *metasdk.Service, taskDesc, outputPath string) error {
	ctx := context.Background()

	fmt.Printf("🔧 Generating configuration for task: %s\n", taskDesc)
	fmt.Println("⚠️  Quick mode: no interactive confirmations")

	// Step 1: Analyze
	fmt.Println("\n📊 Analyzing task...")
	analysis, trace1, err := service.Agents().TaskAnalyzer.Run(ctx, taskDesc)
	if err != nil {
		return fmt.Errorf("task analysis failed: %w", err)
	}

	fmt.Printf("✓ Analysis complete (%.1fs)\n", trace1.Duration.Seconds())
	fmt.Printf("  Complexity: %s\n", analysis.Complexity)
	fmt.Printf("  Agents: %d\n", analysis.AgentCount)

	// Check for clarifications
	if analysis.NeedsClarification {
		fmt.Println("\n⚠️  Agent needs clarification:")
		for i, q := range analysis.Questions {
			fmt.Printf("  %d. %s\n", i+1, q.Question)
		}
		fmt.Println("\n💡 Use --interactive mode to answer questions")
		return fmt.Errorf("clarification needed, please use interactive mode")
	}

	// Step 2: Generate
	fmt.Println("\n⚙️ Generating YAML...")
	analysisJSON, _ := json.Marshal(analysis)
	input := metasdk.GenerationInput{
		Analysis:               string(analysisJSON),
		RefinementInstructions: "",
		UserAnswers:            make(map[string]string),
	}

	config, trace2, err := service.Agents().YamlGenerator.Run(ctx, input)
	if err != nil {
		return fmt.Errorf("YAML generation failed: %w", err)
	}

	fmt.Printf("✓ Generation complete (%.1fs)\n", trace2.Duration.Seconds())

	// Step 3: Save
	if err := os.WriteFile(outputPath, []byte(config.YamlContent), 0644); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("\n💾 Saved to: %s\n", outputPath)

	// Step 4: Validate
	fmt.Println("\n✅ Validating...")
	spec, err := core.LoadSpec(outputPath)
	if err != nil {
		fmt.Printf("⚠️  Validation warning: %v\n", err)
		return nil
	}

	fmt.Printf("✓ Valid! %d types, %d assistants\n", len(spec.Types), len(spec.Assistants))

	return nil
}

// UI Helper functions
func printHeader(text string) {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println(text)
	fmt.Println(strings.Repeat("=", 70))
}

func printSeparator() {
	fmt.Println(strings.Repeat("-", 70))
}

func printAnalysisResult(analysis *metasdk.TaskAnalysis) {
	fmt.Printf("📊 Complexity: %s (%.1f/10)\n", strings.ToUpper(analysis.Complexity), analysis.ComplexityScore)
	fmt.Printf("🤖 Suggested agents: %d\n\n", analysis.AgentCount)

	for i, agent := range analysis.Agents {
		fmt.Printf("%d. %s (%s/%s)\n", i+1, agent.Name, agent.Provider, agent.Model)
		fmt.Printf("   Role: %s\n", agent.Role)
		if agent.Reasoning != "" {
			fmt.Printf("   Why: %s\n", agent.Reasoning)
		}
		fmt.Println()
	}

	if analysis.RequiresThread {
		fmt.Println("🔗 Thread context: Required")
	}
	if analysis.RequiresDialog {
		fmt.Println("💬 Dialog mode: Required")
	}

	if analysis.ArchitectureNotes != "" {
		fmt.Printf("\n💡 Architecture: %s\n", analysis.ArchitectureNotes)
	}

	if len(analysis.ImplementationHints) > 0 {
		fmt.Println("\n💡 Implementation hints:")
		for _, hint := range analysis.ImplementationHints {
			fmt.Printf("   • %s\n", hint)
		}
	}
}
