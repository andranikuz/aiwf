package main

import (
	"context"
	"fmt"
	"log"
	"os"

	sdk "github.com/andranikuz/aiwf/examples/generated"
	"github.com/andranikuz/aiwf/providers/openai"
)

func main() {
	// Check for OpenAI API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Initialize OpenAI client
	config := openai.ClientConfig{
		APIKey:  apiKey,
		BaseURL: "https://api.openai.com/v1",
	}
	provider, _ := openai.NewClient(config)

	// Create service with all agents
	service := sdk.NewService(provider)

	// Optional: Setup thread manager for conversational agent
	// In real usage, use: threadManager := aiwf.NewInMemoryThreadManager()
	// service.WithThreadManager(threadManager)

	ctx := context.Background()

	fmt.Println("=== AIWF Example: Demonstrating Three Agent Types ===\n")

	// Example 1: Complex Structured Output - Data Analyst
	fmt.Println("1. DATA ANALYST (Complex Structured Output)")
	fmt.Println("-------------------------------------------")

	dataRequest := sdk.DataAnalysisRequest{
		Dataset: `Sales data for Q4 2023:
			- October: $1.2M (500 transactions)
			- November: $1.8M (750 transactions)
			- December: $2.1M (920 transactions)
			- Returns: 3.2% average across quarter
			- Top category: Electronics (45% of sales)`,
		Query:                "Analyze the sales trend and provide insights",
		AnalysisType:         "statistical",
		IncludeVisualization: true,
		ConfidenceThreshold:  0.75,
	}

	dataResult, trace, err := service.Agents().DataAnalyst.Run(ctx, dataRequest)
	if err != nil {
		log.Printf("Data analysis error: %v\n", err)
	} else {
		fmt.Printf("Analysis Summary: %s\n", dataResult.Summary)
		fmt.Printf("Confidence Score: %.2f\n", dataResult.ConfidenceScore)
		fmt.Printf("Key Findings: %d found\n", len(dataResult.Findings))
		for i, finding := range dataResult.Findings {
			fmt.Printf("  %d. [%s] %s\n", i+1, finding.Importance, finding.Title)
		}
		fmt.Printf("Processing Time: %dms\n", dataResult.Metrics.ProcessingTimeMs)
		if trace != nil {
			fmt.Printf("Tokens Used: Prompt=%d, Total=%d\n", trace.Usage.Prompt, trace.Usage.Total)
		}
	}

	fmt.Println()

	// Example 2: Simple Text Output - Creative Writer
	fmt.Println("2. CREATIVE WRITER (Simple Text Output)")
	fmt.Println("----------------------------------------")

	writingRequest := sdk.CreativeWritingRequest{
		Prompt:    "Write about a mysterious lighthouse keeper who discovers an ancient map",
		Style:     "narrative",
		WordCount: 150,
		Tone:      "dramatic",
	}

	// Creative writer returns plain string (no structured output)
	creativeText, trace, err := service.Agents().CreativeWriter.Run(ctx, writingRequest)
	if err != nil {
		log.Printf("Creative writing error: %v\n", err)
	} else {
		fmt.Println("Generated Story:")
		fmt.Println(*creativeText)
		if trace != nil {
			fmt.Printf("\n[Generated in %s using %d tokens]\n", trace.Duration, trace.Usage.Total)
		}
	}

	fmt.Println()

	// Example 3: Thread-Aware Conversational Agent - Customer Support
	fmt.Println("3. CUSTOMER SUPPORT (Thread-Aware Conversation)")
	fmt.Println("------------------------------------------------")

	// First customer message
	query1 := sdk.CustomerQuery{
		Message:    "Hi, I'm having trouble logging into my account. It says my password is incorrect but I'm sure it's right.",
		Category:   "technical",
		Urgency:    "high",
		CustomerId: "CUST-12345",
	}

	// Create or get thread state for this customer
	// threadID := "support-session-" + query1.CustomerId
	// In real usage: thread := threadManager.GetOrCreateThread(ctx, threadID)

	// First interaction (simplified without thread manager)
	response1, trace, err := service.Agents().CustomerSupport.Run(ctx, query1)
	if err != nil {
		log.Printf("Support error: %v\n", err)
	} else {
		fmt.Printf("Support: %s\n", response1.Reply)
		fmt.Printf("Ticket ID: %s\n", response1.TicketId)
		fmt.Printf("Issue Resolved: %v\n", response1.Resolved)
		if response1.FollowupNeeded {
			fmt.Println("Follow-up needed: Yes")
		}
	}

	// Follow-up message in the same thread
	query2 := sdk.CustomerQuery{
		Message:    "I tried resetting my password but didn't receive the email.",
		Category:   "technical",
		Urgency:    "high",
		CustomerId: "CUST-12345",
	}

	// Continue conversation (simplified without thread context)
	response2, _, err := service.Agents().CustomerSupport.Run(ctx, query2)
	if err != nil {
		log.Printf("Support follow-up error: %v\n", err)
	} else {
		fmt.Printf("\nFollow-up Support: %s\n", response2.Reply)
		fmt.Printf("Next Steps:\n")
		for i, step := range response2.NextSteps {
			fmt.Printf("  %d. %s\n", i+1, step)
		}
		fmt.Printf("Predicted Satisfaction: %.1f/5.0\n", response2.SatisfactionPredicted)
	}

	// Example of using dialog mode (if max_rounds > 1)
	if service.Agents().CustomerSupport.ThreadBinding() != nil {
		fmt.Println("\n[Thread-aware agent maintains conversation context across interactions]")
	}

	fmt.Println("\n=== Example Complete ===")
	fmt.Println("\nThis example demonstrated:")
	fmt.Println("1. Complex structured output with nested types (DataAnalyst)")
	fmt.Println("2. Simple text generation without JSON structure (CreativeWriter)")
	fmt.Println("3. Conversational agent with thread/context awareness (CustomerSupport)")
}

// Helper function to demonstrate type validation
func validateInputs(service *sdk.Service) {
	fmt.Println("\n=== Type Validation Examples ===")

	// Invalid data request (confidence threshold out of range)
	invalidRequest := sdk.DataAnalysisRequest{
		Dataset:             "test",
		Query:               "analyze",
		AnalysisType:        "statistical",
		ConfidenceThreshold: 1.5, // Invalid: exceeds max of 1.0
	}

	if err := sdk.ValidateDataAnalysisRequest(&invalidRequest); err != nil {
		fmt.Printf("Validation correctly caught error: %v\n", err)
	}

	// Valid creative request
	validCreative := sdk.CreativeWritingRequest{
		Prompt:    "Write something",
		Style:     "poetic",
		WordCount: 100,
		Tone:      "formal",
	}

	if err := sdk.ValidateCreativeWritingRequest(&validCreative); err == nil {
		fmt.Println("Creative writing request validation passed")
	}
}
