package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	campaign "github.com/andranikuz/aiwf/examples/campaign/sdk"
	"github.com/andranikuz/aiwf/providers/openai"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY env var is required")
	}

	client, err := openai.NewClient(openai.ClientConfig{APIKey: apiKey})
	if err != nil {
		log.Fatalf("create openai client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	svc := campaign.NewService(client)

	researchInput := campaign.MarketResearchInput{
		ProductName:  "Nimbus",
		Description:  "AI copilot that orchestrates real-time approvals across SaaS tools",
		TargetMarket: "Mid-market operations teams",
		Goals:        []string{"Accelerate onboarding", "Reduce manual follow-ups"},
	}

	research, trResearch, err := svc.Agents().MarketResearch().Run(ctx, researchInput)
	if err != nil {
		log.Fatalf("market research failed: %v", err)
	}

	fmt.Println("\n=== Market Research Summary ===")
	fmt.Println(research.Summary)
	for _, segment := range research.Segments {
		name := stringField(segment, "name")
		persona := stringField(segment, "persona")
		pains := strings.Join(stringSlice(segment, "pain_points"), ", ")
		fmt.Printf("- %s: %s | pains: %s\n", name, persona, pains)
	}

	messagingInput := campaign.ValuePropositionInput{
		ProductName: researchInput.ProductName,
		Segments:    research.Segments,
	}
	messaging, _, err := svc.Agents().ValueProposition().Run(ctx, messagingInput)
	if err != nil {
		log.Fatalf("value proposition failed: %v", err)
	}

	fmt.Println("\n=== Value Map ===")
	for _, item := range messaging.ValueMap {
		fmt.Printf("%s → %s\n", stringField(item, "segment"), stringField(item, "core_message"))
	}

	strategyInput := campaign.ContentStrategyInput{
		ValueMap:    messaging.ValueMap,
		PrimaryGoal: "Generate 500 qualified leads",
		BudgetLevel: "medium",
	}
	strategy, _, err := svc.Agents().ContentStrategy().Run(ctx, strategyInput)
	if err != nil {
		log.Fatalf("content strategy failed: %v", err)
	}

	fmt.Println("\n=== Channel Plan ===")
	for _, ch := range strategy.Channels {
		fmt.Printf("%s (%s) – cadence: %s\n", stringField(ch, "name"), stringField(ch, "format"), stringField(ch, "cadence"))
	}

	timeline, _, err := svc.Agents().LaunchTimeline().Run(ctx, campaign.LaunchTimelineInput(strategy))
	if err != nil {
		log.Fatalf("timeline failed: %v", err)
	}

	fmt.Println("\n=== Launch Timeline ===")
	for _, m := range timeline.Milestones {
		fmt.Printf("Week %s: %s (owner: %s)\n", stringField(m, "week"), stringField(m, "objective"), stringField(m, "owner"))
	}

	risks, trRisks, err := svc.Agents().RiskAssessment().Run(ctx, campaign.RiskAssessmentInput(timeline))
	if err != nil {
		log.Fatalf("risk assessment failed: %v", err)
	}

	fmt.Println("\n=== Risk Matrix ===")
	for _, r := range risks.Risks {
		fmt.Printf("%s → impact: %s, mitigation: %s\n", stringField(r, "label"), stringField(r, "impact"), stringField(r, "mitigation"))
	}

	fmt.Printf("\nTokens used: research=%d, risks=%d\n", trResearch.Usage.Total, trRisks.Usage.Total)
}

func stringField(m map[string]any, key string) string {
	if val, ok := m[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
		if f, ok := val.(float64); ok {
			return fmt.Sprintf("%g", f)
		}
	}
	return ""
}

func stringSlice(m map[string]any, key string) []string {
	raw, ok := m[key]
	if !ok {
		return nil
	}
	arr, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}
