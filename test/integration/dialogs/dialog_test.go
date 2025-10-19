package dialogs

import (
	"context"
	"testing"
	"time"

	"github.com/andranikuz/aiwf/providers/openai"
	"github.com/andranikuz/aiwf/runtime/go/aiwf"
	customer_support "github.com/andranikuz/aiwf/test/integration/dialogs/generated/customer_support"
)

// TestDialog_ThreadManagerIntegration tests ThreadManager with dialog service
func TestDialog_ThreadManagerIntegration(t *testing.T) {
	skipIfNoAPIKey(t)

	// Create service with ThreadManager
	threadManager := openai.NewInMemoryThreadManager()
	clientWithThreads := openaiClient.WithThreadManager(threadManager)

	service := customer_support.NewService(clientWithThreads)
	if service == nil {
		t.Fatal("Expected service, got nil")
	}

	t.Logf("✓ Service created with ThreadManager")

	// Verify ThreadManager is accessible
	if threadManager == nil {
		t.Fatal("Expected ThreadManager, got nil")
	}

	t.Logf("✓ ThreadManager initialized and configured")
}

// TestDialog_ThreadLifecycle tests complete thread lifecycle
func TestDialog_ThreadLifecycle(t *testing.T) {
	skipIfNoAPIKey(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	threadManager := openai.NewInMemoryThreadManager()

	// Test Start - create new thread
	binding := aiwf.ThreadBinding{
		Name:     "default",
		Provider: "openai",
		Strategy: "append",
	}

	threadState, err := threadManager.Start(ctx, "support_bot", binding)
	if err != nil {
		t.Fatalf("Failed to start thread: %v", err)
	}

	if threadState == nil {
		t.Fatal("Expected ThreadState, got nil")
	}

	if threadState.ID == "" {
		t.Fatal("Expected thread ID, got empty string")
	}

	t.Logf("✓ Thread created: %s", threadState.ID)

	// Test Continue - append message to thread
	feedback := "Please provide more details about the issue"
	err = threadManager.Continue(ctx, threadState, feedback)
	if err != nil {
		t.Fatalf("Failed to continue thread: %v", err)
	}

	t.Logf("✓ Feedback appended to thread")

	// Test Close - close the thread
	err = threadManager.Close(ctx, threadState)
	if err != nil {
		t.Fatalf("Failed to close thread: %v", err)
	}

	t.Logf("✓ Thread closed successfully")
}

// TestDialog_MultipleThreads tests managing multiple threads concurrently
func TestDialog_MultipleThreads(t *testing.T) {
	skipIfNoAPIKey(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	threadManager := openai.NewInMemoryThreadManager()
	binding := aiwf.ThreadBinding{
		Name:     "default",
		Provider: "openai",
		Strategy: "append",
	}

	threads := make([]*aiwf.ThreadState, 3)

	// Create multiple threads
	for i := 0; i < 3; i++ {
		state, err := threadManager.Start(ctx, "support_bot", binding)
		if err != nil {
			t.Fatalf("Failed to create thread %d: %v", i, err)
		}
		threads[i] = state
		t.Logf("✓ Thread %d created: %s", i, state.ID)
	}

	// Add messages to each thread
	for i, state := range threads {
		feedback := "Feedback for thread " + state.ID
		err := threadManager.Continue(ctx, state, feedback)
		if err != nil {
			t.Fatalf("Failed to continue thread %d: %v", i, err)
		}
	}

	t.Logf("✓ Messages appended to all %d threads", len(threads))

	// Close all threads
	for i, state := range threads {
		err := threadManager.Close(ctx, state)
		if err != nil {
			t.Fatalf("Failed to close thread %d: %v", i, err)
		}
	}

	t.Logf("✓ All %d threads closed successfully", len(threads))
}

// TestDialog_SingleTurnWithThreads tests single-turn dialog with thread support
func TestDialog_SingleTurnWithThreads(t *testing.T) {
	skipIfNoAPIKey(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Set up service with ThreadManager
	threadManager := openai.NewInMemoryThreadManager()
	clientWithThreads := openaiClient.WithThreadManager(threadManager)
	service := customer_support.NewService(clientWithThreads)

	customer := &customer_support.Customer{
		Id:                 "550e8400-e29b-41d4-a716-446655440000",
		Name:               "John Doe",
		Email:              "john@example.com",
		SubscriptionTier:   "pro",
		AccountCreated:     mustParseTime(t, "2023-01-15T10:30:00Z"),
	}

	input := customer_support.SupportMessage{
		CustomerId: customer.Id,
		Message:    "I'm unable to access my dashboard. I keep getting a 403 error when I try to log in.",
		Context: &customer_support.ConversationContext{
			SessionId:  "sess-thread-test",
			TicketId:   "ticket-thread-test",
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

	if trace == nil {
		t.Fatal("Expected trace, got nil")
	}

	t.Logf("✓ Single-turn dialog completed")
	t.Logf("  Customer: %s", customer.Name)
	t.Logf("  Status: %s", result.ResolutionStatus)
	t.Logf("  Escalate: %v", result.Escalate)
	t.Logf("  Tokens: %d (prompt) / %d (completion)", trace.Usage.Prompt, trace.Usage.Completion)
}

// TestDialog_ThreadStateMetadata tests thread metadata preservation
func TestDialog_ThreadStateMetadata(t *testing.T) {
	skipIfNoAPIKey(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	threadManager := openai.NewInMemoryThreadManager()
	binding := aiwf.ThreadBinding{
		Name:     "default",
		Provider: "openai",
		Strategy: "append",
	}

	state, err := threadManager.Start(ctx, "support_bot", binding)
	if err != nil {
		t.Fatalf("Failed to start thread: %v", err)
	}

	// Verify metadata is present
	if state.Metadata == nil {
		t.Fatal("Expected metadata map, got nil")
	}

	t.Logf("✓ Thread metadata initialized")

	// Add metadata
	state.Metadata["customer_id"] = "cust-123"
	state.Metadata["priority"] = "high"

	// Verify metadata persists
	if val, ok := state.Metadata["customer_id"]; !ok || val != "cust-123" {
		t.Fatal("Metadata not preserved correctly")
	}

	t.Logf("✓ Thread metadata stored and retrieved correctly")

	// Clean up
	_ = threadManager.Close(ctx, state)
}

// TestDialog_ErrorHandling tests error handling in ThreadManager
func TestDialog_ErrorHandling(t *testing.T) {
	skipIfNoAPIKey(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	threadManager := openai.NewInMemoryThreadManager()

	// Test: Continue on non-existent thread
	nonExistentState := &aiwf.ThreadState{
		ID:       "non-existent-thread",
		Metadata: map[string]any{},
	}

	err := threadManager.Continue(ctx, nonExistentState, "feedback")
	if err == nil {
		t.Fatal("Expected error for non-existent thread, got nil")
	}

	t.Logf("✓ Error handling works: %v", err)

	// Test: Close on non-existent thread
	err = threadManager.Close(ctx, nonExistentState)
	if err == nil {
		t.Fatal("Expected error for closing non-existent thread, got nil")
	}

	t.Logf("✓ Close error handling works: %v", err)

	// Test: Start with nil state should fail gracefully
	err = threadManager.Continue(ctx, nil, "feedback")
	if err == nil {
		t.Fatal("Expected error for nil thread state, got nil")
	}

	t.Logf("✓ Nil state error handling works: %v", err)
}

// TestDialog_VeryBasicWorkflow tests a very basic dialog workflow
func TestDialog_BasicWorkflow(t *testing.T) {
	skipIfNoAPIKey(t)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	threadManager := openai.NewInMemoryThreadManager()
	clientWithThreads := openaiClient.WithThreadManager(threadManager)
	service := customer_support.NewService(clientWithThreads)

	customer := &customer_support.Customer{
		Id:               "550e8400-e29b-41d4-a716-446655440001",
		Name:             "Jane Smith",
		Email:            "jane@example.com",
		SubscriptionTier: "enterprise",
		AccountCreated:   mustParseTime(t, "2022-06-01T14:00:00Z"),
	}

	// Simulate first turn of conversation
	firstMessage := customer_support.SupportMessage{
		CustomerId: customer.Id,
		Message:    "Our API integration stopped working after your recent update. The v2 endpoints are returning 500 errors.",
		Context: &customer_support.ConversationContext{
			SessionId:        "sess-workflow-001",
			TicketId:         "ticket-workflow-001",
			PreviousMessages: []*customer_support.Message{},
			CustomerInfo:     customer,
		},
		Attachments: []*customer_support.Attachment{},
	}

	t.Logf("┌─ Dialog Workflow Test")
	t.Logf("├─ Turn 1: Initial Support Request")
	result1, trace1, err := service.Agents().SupportBot.Run(ctx, firstMessage)
	if err != nil {
		t.Fatalf("Turn 1 failed: %v", err)
	}

	if result1 == nil {
		t.Fatal("Turn 1: Expected result, got nil")
	}

	t.Logf("│  Response: %s", result1.Message[:60]+"...")
	t.Logf("│  Status: %s | Escalate: %v", result1.ResolutionStatus, result1.Escalate)
	t.Logf("│  Tokens: %d+%d", trace1.Usage.Prompt, trace1.Usage.Completion)

	// Simulate second turn
	secondMessage := customer_support.SupportMessage{
		CustomerId: customer.Id,
		Message:    "We've implemented the fix on our end but need to verify with your support team that the issue is resolved.",
		Context: &customer_support.ConversationContext{
			SessionId:    "sess-workflow-001",
			TicketId:     "ticket-workflow-001",
			CustomerInfo: customer,
			PreviousMessages: []*customer_support.Message{
				{
					Id:        "msg-1",
					SenderType: "customer",
					Content:   firstMessage.Message,
					Timestamp: mustParseTime(t, "2024-10-19T12:00:00Z"),
				},
			},
		},
	}

	t.Logf("├─ Turn 2: Follow-up Message")
	result2, trace2, err := service.Agents().SupportBot.Run(ctx, secondMessage)
	if err != nil {
		t.Fatalf("Turn 2 failed: %v", err)
	}

	if result2 == nil {
		t.Fatal("Turn 2: Expected result, got nil")
	}

	t.Logf("│  Response: %s", result2.Message[:60]+"...")
	t.Logf("│  Status: %s | Escalate: %v", result2.ResolutionStatus, result2.Escalate)
	t.Logf("│  Tokens: %d+%d", trace2.Usage.Prompt, trace2.Usage.Completion)

	totalTokens := trace1.Usage.Total + trace2.Usage.Total
	t.Logf("└─ Total tokens used: %d", totalTokens)
	t.Logf("✓ Dialog workflow completed successfully")
}
