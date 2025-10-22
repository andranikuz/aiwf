package dialogs

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/andranikuz/aiwf/providers/openai"
	"github.com/andranikuz/aiwf/runtime/go/aiwf"
	customer_support "github.com/andranikuz/aiwf/test/integration/dialogs/generated/customer_support"
)

// TestDialogDecider_DefaultCompletes tests that DefaultDialogDecider completes after one pass
func TestDialogDecider_DefaultCompletes(t *testing.T) {
	skipIfNoAPIKey(t)

	decider := &aiwf.DefaultDialogDecider{}

	ctx := aiwf.DialogContext{
		Step:    "support_bot",
		Output:  "Test response",
		Attempt: 1,
	}

	decision := decider.Decide(ctx)
	if decision.Action != aiwf.DialogActionComplete {
		t.Fatalf("Expected Complete, got %v", decision.Action)
	}

	t.Logf("✓ DefaultDialogDecider completes immediately")
}

// TestDialogDecider_ImmediateRetry tests ImmediateRetryDecider behavior
func TestDialogDecider_ImmediateRetry(t *testing.T) {
	skipIfNoAPIKey(t)

	decider := openai.NewImmediateRetryDecider(2, "Please try again")

	// First attempt should retry
	decision := decider.Decide(aiwf.DialogContext{
		Step:    "support_bot",
		Output:  "response",
		Attempt: 1,
	})
	if decision.Action != aiwf.DialogActionRetry {
		t.Fatalf("Attempt 1: Expected Retry, got %v", decision.Action)
	}

	// Second attempt (at max) should still retry
	decision = decider.Decide(aiwf.DialogContext{
		Step:    "support_bot",
		Output:  "response",
		Attempt: 2,
	})
	if decision.Action != aiwf.DialogActionRetry {
		t.Fatalf("Attempt 2: Expected Retry, got %v", decision.Action)
	}

	// Third attempt (exceeds max) should complete
	decision = decider.Decide(aiwf.DialogContext{
		Step:    "support_bot",
		Output:  "response",
		Attempt: 3,
	})
	if decision.Action != aiwf.DialogActionComplete {
		t.Fatalf("Attempt 3: Expected Complete, got %v", decision.Action)
	}

	t.Logf("✓ ImmediateRetryDecider enforces max attempts correctly")
}

// TestDialogDecider_LengthCheck tests LengthCheckDecider with dialog context
func TestDialogDecider_LengthCheck(t *testing.T) {
	skipIfNoAPIKey(t)

	decider := openai.NewLengthCheckDecider(100)

	// Short response should trigger retry
	shortResp := &customer_support.SupportResponse{
		Message:            "Short",
		ResolutionStatus:   "pending",
		Escalate:           false,
		Actions:            []*customer_support.Action{},
	}

	decision := decider.Decide(aiwf.DialogContext{
		Step:    "support_bot",
		Output:  shortResp,
		Attempt: 1,
	})
	if decision.Action != aiwf.DialogActionRetry {
		t.Fatalf("Expected Retry for short response, got %v", decision.Action)
	}
	if !strings.Contains(decision.Feedback, "too short") {
		t.Fatalf("Expected feedback about length, got %q", decision.Feedback)
	}

	// Long response should complete
	longResp := &customer_support.SupportResponse{
		Message:            "This is a much longer response that should pass the length requirement by providing sufficient detail and context for the customer to understand the resolution",
		ResolutionStatus:   "resolved",
		Escalate:           false,
		Actions:            []*customer_support.Action{},
	}

	decision = decider.Decide(aiwf.DialogContext{
		Step:    "support_bot",
		Output:  longResp,
		Attempt: 1,
	})
	if decision.Action != aiwf.DialogActionComplete {
		t.Fatalf("Expected Complete for long response, got %v", decision.Action)
	}

	t.Logf("✓ LengthCheckDecider validates response length correctly")
}

// TestDialogDecider_QualityCheck tests QualityCheckDecider for response validation
func TestDialogDecider_QualityCheck(t *testing.T) {
	skipIfNoAPIKey(t)

	// Quality check: ensure message is not empty and contains helpful content
	decider := openai.NewQualityCheckDecider(func(output any) (bool, string) {
		resp, ok := output.(*customer_support.SupportResponse)
		if !ok {
			return false, "Invalid response type"
		}

		if resp.Message == "" {
			return false, "Message cannot be empty"
		}

		// Check if response addresses the issue or escalates
		if resp.ResolutionStatus == "pending" &&
			!resp.Escalate && len(resp.Actions) == 0 {
			return false, "Pending resolution should either suggest actions or escalate"
		}

		return true, ""
	})

	// Invalid: empty message
	badResp := &customer_support.SupportResponse{
		Message:            "",
		ResolutionStatus:   "pending",
		Escalate:           false,
		Actions:            []*customer_support.Action{},
	}

	decision := decider.Decide(aiwf.DialogContext{
		Step:    "support_bot",
		Output:  badResp,
		Attempt: 1,
	})
	if decision.Action != aiwf.DialogActionRetry {
		t.Fatalf("Expected Retry for empty message, got %v", decision.Action)
	}

	// Valid: resolved with message
	goodResp := &customer_support.SupportResponse{
		Message:            "Your issue has been resolved.",
		ResolutionStatus:   "resolved",
		Escalate:           false,
		Actions:            []*customer_support.Action{},
	}

	decision = decider.Decide(aiwf.DialogContext{
		Step:    "support_bot",
		Output:  goodResp,
		Attempt: 1,
	})
	if decision.Action != aiwf.DialogActionComplete {
		t.Fatalf("Expected Complete for valid response, got %v", decision.Action)
	}

	t.Logf("✓ QualityCheckDecider validates response quality")
}

// TestDialogDecider_Chained tests ChainedDecider combining multiple checks
func TestDialogDecider_Chained(t *testing.T) {
	skipIfNoAPIKey(t)

	// Chain: length check + quality check
	chained := openai.NewChainedDecider(
		openai.NewLengthCheckDecider(50),
		openai.NewQualityCheckDecider(func(output any) (bool, string) {
			resp, ok := output.(*customer_support.SupportResponse)
			if !ok {
				return false, "Invalid response type"
			}
			if resp.ResolutionStatus == "pending" && !resp.Escalate {
				return false, "Must escalate or resolve"
			}
			return true, ""
		}),
	)

	// Too short - fails first check
	tooShort := &customer_support.SupportResponse{
		Message:            "Hi",
		ResolutionStatus:   "pending",
		Escalate:           false,
		Actions:            []*customer_support.Action{},
	}

	decision := chained.Decide(aiwf.DialogContext{
		Step:    "support_bot",
		Output:  tooShort,
		Attempt: 1,
	})
	if decision.Action != aiwf.DialogActionRetry {
		t.Fatalf("Expected Retry from length check, got %v", decision.Action)
	}
	if !strings.Contains(decision.Feedback, "short") {
		t.Fatalf("Expected length feedback, got %q", decision.Feedback)
	}

	// Long but no escalation or resolution - fails second check
	noEscalation := &customer_support.SupportResponse{
		Message:            "This is a very long message that passes the length requirement but does not take action on the customer issue",
		ResolutionStatus:   "pending",
		Escalate:           false,
		Actions:            []*customer_support.Action{},
	}

	decision = chained.Decide(aiwf.DialogContext{
		Step:    "support_bot",
		Output:  noEscalation,
		Attempt: 1,
	})
	if decision.Action != aiwf.DialogActionRetry {
		t.Fatalf("Expected Retry from quality check, got %v", decision.Action)
	}
	if !strings.Contains(decision.Feedback, "escalate") {
		t.Fatalf("Expected quality feedback, got %q", decision.Feedback)
	}

	// Passes both checks
	good := &customer_support.SupportResponse{
		Message:            "Thank you for contacting us. I understand your issue and I am escalating this to our technical team for immediate attention.",
		ResolutionStatus:   "escalated",
		Escalate:           true,
		Actions:            []*customer_support.Action{},
	}

	decision = chained.Decide(aiwf.DialogContext{
		Step:    "support_bot",
		Output:  good,
		Attempt: 1,
	})
	if decision.Action != aiwf.DialogActionComplete {
		t.Fatalf("Expected Complete, got %v", decision.Action)
	}

	t.Logf("✓ ChainedDecider enforces all checks in sequence")
}

// TestDialogDecider_WithDialogFlow demonstrates DialogDecider in a realistic dialog scenario
func TestDialogDecider_WithDialogFlow(t *testing.T) {
	skipIfNoAPIKey(t)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Create service with ThreadManager
	threadManager := openai.NewInMemoryThreadManager()
	clientWithThreads := openaiClient.WithThreadManager(threadManager)
	service := customer_support.NewService(clientWithThreads)

	customer := &customer_support.Customer{
		Id:               "test-decider-customer",
		Name:             "Quality Test User",
		Email:            "quality@example.com",
		SubscriptionTier: "pro",
		AccountCreated:   mustParseTime(t, "2023-01-01T00:00:00Z"),
	}

	input := customer_support.SupportMessage{
		CustomerId: customer.Id,
		Message:    "I need help with my account setup. Can you guide me through the process?",
		Context: &customer_support.ConversationContext{
			SessionId:  "sess-decider-test",
			TicketId:   "ticket-decider-test",
			CustomerInfo: customer,
		},
		Attachments: []*customer_support.Attachment{},
	}

	// Use quality check decider
	qualityDecider := openai.NewQualityCheckDecider(func(output any) (bool, string) {
		resp, ok := output.(*customer_support.SupportResponse)
		if !ok {
			return false, "Invalid response type"
		}

		// Require substantive response
		if len(resp.Message) < 100 {
			return false, "Response must be detailed (min 100 chars)"
		}

		// Require at least some action or escalation
		if len(resp.Actions) == 0 && !resp.Escalate {
			return false, "Response must include actions or escalation"
		}

		return true, ""
	})

	t.Logf("┌─ Dialog with QualityCheckDecider")
	result, trace, err := service.Agents().SupportBot.Run(ctx, input)
	if err != nil {
		t.Fatalf("Agent failed: %v", err)
	}

	t.Logf("├─ Got response")
	t.Logf("│  Message length: %d chars", len(result.Message))
	t.Logf("│  Status: %s", result.ResolutionStatus)
	t.Logf("│  Escalate: %v", result.Escalate)
	t.Logf("│  Actions: %d", len(result.Actions))

	// Now validate with decider
	decision := qualityDecider.Decide(aiwf.DialogContext{
		Step:    "support_bot",
		Output:  result,
		Trace:   trace,
		Attempt: 1,
	})

	t.Logf("├─ QualityCheckDecider decision: %v", decision.Action)
	if decision.Action == aiwf.DialogActionRetry {
		t.Logf("│  Feedback: %s", decision.Feedback)
	}

	t.Logf("└─ Dialog complete")
	t.Logf("✓ DialogDecider integrated with dialog flow")
}
