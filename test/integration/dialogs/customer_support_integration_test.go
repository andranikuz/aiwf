package dialogs

import (
	"context"
	"testing"
	"time"

	customer_support "github.com/andranikuz/aiwf/test/integration/dialogs/generated/customer_support"
)

func TestCustomerSupport_Integration(t *testing.T) {
	skipIfNoAPIKey(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	service := customer_support.NewService(openaiClient)

	// Create sample customer
	customer := &customer_support.Customer{
		ID:                "550e8400-e29b-41d4-a716-446655440000",
		Name:              "John Doe",
		Email:             "john@example.com",
		SubscriptionTier:  "pro",
		AccountCreated:    mustParseTime(t, "2023-01-15T10:30:00Z"),
	}

	// Create support message
	input := customer_support.SupportMessage{
		CustomerID: "550e8400-e29b-41d4-a716-446655440000",
		Message:    "I'm unable to access my dashboard. I keep getting a 403 error when I try to log in.",
		Context: &customer_support.ConversationContext{
			SessionID:      "sess-123456789",
			TicketID:       "ticket-987654321",
			PreviousMessages: []*customer_support.Message{},
			CustomerInfo:   customer,
		},
		Attachments: []*customer_support.Attachment{},
	}

	result, trace, err := service.Agents().SupportBot.Run(ctx, input)
	if err != nil {
		t.Fatalf("SupportBot agent failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if trace == nil {
		t.Fatal("Expected trace, got nil")
	}

	t.Logf("✓ Customer support response generated")
	t.Logf("  Message: %s", result.Message)
	t.Logf("  Status: %s", result.ResolutionStatus)
	t.Logf("  Escalate: %v", result.Escalate)

	if len(result.Actions) > 0 {
		t.Logf("  Suggested actions (%d):", len(result.Actions))
		for i, action := range result.Actions {
			t.Logf("    [%d] %s: %s", i, action.Type, action.Label)
		}
	}

	t.Logf("  Tokens (in/out): %d/%d", trace.Usage.Prompt, trace.Usage.Completion)
}

func TestCustomerSupport_ComplexDialog(t *testing.T) {
	skipIfNoAPIKey(t)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	service := customer_support.NewService(openaiClient)

	customer := &customer_support.Customer{
		ID:               "550e8400-e29b-41d4-a716-446655440000",
		Name:             "Jane Smith",
		Email:            "jane@example.com",
		SubscriptionTier: "enterprise",
		AccountCreated:   mustParseTime(t, "2022-06-01T14:00:00Z"),
	}

	// Simulate a multi-turn conversation
	conversations := []struct {
		message string
		context string
	}{
		{
			message: "Our API integration stopped working after your recent update. The v2 endpoints are returning 500 errors.",
			context: "Integration issue",
		},
		{
			message: "We've implemented the fix on our end but need to verify with your support team that the issue is resolved.",
			context: "Follow-up on fix verification",
		},
	}

	for _, conv := range conversations {
		t.Run(conv.context, func(t *testing.T) {
			input := customer_support.SupportMessage{
				CustomerID: customer.ID,
				Message:    conv.message,
				Context: &customer_support.ConversationContext{
					SessionID:     "sess-enterprise-001",
					TicketID:      "ticket-critical-001",
					CustomerInfo: customer,
				},
				Attachments: []*customer_support.Attachment{},
			}

			result, trace, err := service.Agents().SupportBot.Run(ctx, input)
			if err != nil {
				t.Fatalf("SupportBot failed: %v", err)
			}

			if result == nil {
				t.Fatal("Expected result, got nil")
			}

			t.Logf("Context: %s", conv.context)
			t.Logf("  Customer: %s (%s)", customer.Name, customer.SubscriptionTier)
			t.Logf("  Response: %s", result.Message[:50]+"...")
			t.Logf("  Status: %s | Escalate: %v", result.ResolutionStatus, result.Escalate)
			t.Logf("  Tokens: %d", trace.Usage.Prompt+trace.Usage.Completion)
		})
	}
}

func TestCustomerSupport_DifferentSubscriptions(t *testing.T) {
	skipIfNoAPIKey(t)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	service := customer_support.NewService(openaiClient)

	subscriptions := []struct {
		tier  string
		issue string
	}{
		{
			tier:  "free",
			issue: "I'm on the free plan and want to know about upgrading to paid features",
		},
		{
			tier:  "basic",
			issue: "I'm experiencing slow response times on the API. Can this be optimized?",
		},
		{
			tier:  "pro",
			issue: "I need to integrate your API with my enterprise system. What's the best approach?",
		},
		{
			tier:  "enterprise",
			issue: "We need a dedicated account manager and custom SLA. What are the next steps?",
		},
	}

	for _, sub := range subscriptions {
		t.Run(sub.tier, func(t *testing.T) {
			customer := &customer_support.Customer{
				ID:               "550e8400-e29b-41d4-a716-446655440000",
				Name:             "Support User",
				Email:            "user@example.com",
				SubscriptionTier: sub.tier,
				AccountCreated:   mustParseTime(t, "2023-01-01T00:00:00Z"),
			}

			input := customer_support.SupportMessage{
				CustomerID: customer.ID,
				Message:    sub.issue,
				Context: &customer_support.ConversationContext{
					SessionID:    "sess-sub-test",
					TicketID:     "ticket-sub-" + sub.tier,
					CustomerInfo: customer,
				},
				Attachments: []*customer_support.Attachment{},
			}

			result, _, err := service.Agents().SupportBot.Run(ctx, input)
			if err != nil {
				t.Fatalf("SupportBot failed: %v", err)
			}

			if result == nil {
				t.Fatal("Expected result, got nil")
			}

			t.Logf("Subscription: %s", sub.tier)
			t.Logf("  Response status: %s", result.ResolutionStatus)
			t.Logf("  Escalate: %v", result.Escalate)
			if len(result.Actions) > 0 {
				t.Logf("  Actions: %d", len(result.Actions))
			}
		})
	}
}

func TestCustomerSupport_WithAttachments(t *testing.T) {
	skipIfNoAPIKey(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	service := customer_support.NewService(openaiClient)

	customer := &customer_support.Customer{
		ID:               "550e8400-e29b-41d4-a716-446655440000",
		Name:             "Bob Wilson",
		Email:            "bob@example.com",
		SubscriptionTier: "pro",
		AccountCreated:   mustParseTime(t, "2023-03-15T08:30:00Z"),
	}

	attachments := []*customer_support.Attachment{
		{
			Filename: "error_screenshot.png",
			MimeType: "image/png",
			SizeBytes: 256000,
			URL:      "https://storage.example.com/attachments/error_screenshot.png",
		},
		{
			Filename: "logs.txt",
			MimeType: "text/plain",
			SizeBytes: 4096,
			URL:      "https://storage.example.com/attachments/logs.txt",
		},
	}

	input := customer_support.SupportMessage{
		CustomerID:  customer.ID,
		Message:     "I'm getting an error when trying to use the bulk import feature. I've attached a screenshot and the error logs.",
		Attachments: attachments,
		Context: &customer_support.ConversationContext{
			SessionID:    "sess-with-attachments",
			TicketID:     "ticket-with-attachments",
			CustomerInfo: customer,
		},
	}

	result, trace, err := service.Agents().SupportBot.Run(ctx, input)
	if err != nil {
		t.Fatalf("SupportBot failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	t.Logf("✓ Support request with attachments processed")
	t.Logf("  Attachments: %d", len(input.Attachments))
	for _, att := range input.Attachments {
		t.Logf("    - %s (%s, %d bytes)", att.Filename, att.MimeType, att.SizeBytes)
	}
	t.Logf("  Response status: %s", result.ResolutionStatus)
	t.Logf("  Tokens (in/out): %d/%d", trace.Usage.Prompt, trace.Usage.Completion)
}

// Helper function to parse time
func mustParseTime(t *testing.T, timeStr string) time.Time {
	parsed, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		t.Fatalf("Failed to parse time: %v", err)
	}
	return parsed
}
