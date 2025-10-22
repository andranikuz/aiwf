package openai

import (
	"testing"

	"github.com/andranikuz/aiwf/runtime/go/aiwf"
)

func TestImmediateRetryDecider(t *testing.T) {
	decider := NewImmediateRetryDecider(2, "Try again")

	// First attempt should request retry
	decision := decider.Decide(aiwf.DialogContext{
		Step:    "test",
		Output:  "output",
		Attempt: 1,
	})
	if decision.Action != aiwf.DialogActionRetry {
		t.Errorf("Expected Retry, got %v", decision.Action)
	}

	// Second attempt (at max) should still retry
	decision = decider.Decide(aiwf.DialogContext{
		Step:    "test",
		Output:  "output",
		Attempt: 2,
	})
	if decision.Action != aiwf.DialogActionRetry {
		t.Errorf("Expected Retry, got %v", decision.Action)
	}

	// Third attempt (exceeds max) should complete
	decision = decider.Decide(aiwf.DialogContext{
		Step:    "test",
		Output:  "output",
		Attempt: 3,
	})
	if decision.Action != aiwf.DialogActionComplete {
		t.Errorf("Expected Complete, got %v", decision.Action)
	}

	t.Logf("✓ ImmediateRetryDecider respects MaxAttempts")
}

func TestQualityCheckDecider(t *testing.T) {
	// Check function that validates minimum word count
	checkFn := func(output any) (bool, string) {
		str := output.(string)
		wordCount := len(str)
		if wordCount >= 10 {
			return true, ""
		}
		return false, "Need more detail"
	}

	decider := NewQualityCheckDecider(checkFn)

	// Good output should complete
	decision := decider.Decide(aiwf.DialogContext{
		Step:    "test",
		Output:  "This is a long enough response",
		Attempt: 1,
	})
	if decision.Action != aiwf.DialogActionComplete {
		t.Errorf("Expected Complete for good output, got %v", decision.Action)
	}

	// Bad output should retry
	decision = decider.Decide(aiwf.DialogContext{
		Step:    "test",
		Output:  "short",
		Attempt: 1,
	})
	if decision.Action != aiwf.DialogActionRetry {
		t.Errorf("Expected Retry for bad output, got %v", decision.Action)
	}
	if decision.Feedback != "Need more detail" {
		t.Errorf("Expected correct feedback, got %s", decision.Feedback)
	}

	t.Logf("✓ QualityCheckDecider validates output quality")
}

func TestLengthCheckDecider(t *testing.T) {
	decider := NewLengthCheckDecider(50)

	// Short output should retry
	decision := decider.Decide(aiwf.DialogContext{
		Step:    "test",
		Output:  "short",
		Attempt: 1,
	})
	if decision.Action != aiwf.DialogActionRetry {
		t.Errorf("Expected Retry for short output, got %v", decision.Action)
	}

	// Long output should complete
	longOutput := "This is a very long response that exceeds the minimum required length for validation"
	decision = decider.Decide(aiwf.DialogContext{
		Step:    "test",
		Output:  longOutput,
		Attempt: 1,
	})
	if decision.Action != aiwf.DialogActionComplete {
		t.Errorf("Expected Complete for long output, got %v", decision.Action)
	}

	// After max attempts, should complete even with short output
	decision = decider.Decide(aiwf.DialogContext{
		Step:    "test",
		Output:  "short",
		Attempt: 4,
	})
	if decision.Action != aiwf.DialogActionComplete {
		t.Errorf("Expected Complete after max attempts, got %v", decision.Action)
	}

	t.Logf("✓ LengthCheckDecider validates minimum length")
}

func TestJSONValidationDecider(t *testing.T) {
	schema := map[string]any{
		"message": nil,
		"status":  nil,
	}
	decider := NewJSONValidationDecider(schema)

	// Valid JSON with required fields should complete
	validJSON := `{"message":"hello","status":"ok"}`
	decision := decider.Decide(aiwf.DialogContext{
		Step:    "test",
		Output:  validJSON,
		Attempt: 1,
	})
	if decision.Action != aiwf.DialogActionComplete {
		t.Errorf("Expected Complete for valid JSON, got %v", decision.Action)
	}

	// Invalid JSON should retry
	invalidJSON := `{invalid json}`
	decision = decider.Decide(aiwf.DialogContext{
		Step:    "test",
		Output:  invalidJSON,
		Attempt: 1,
	})
	if decision.Action != aiwf.DialogActionRetry {
		t.Errorf("Expected Retry for invalid JSON, got %v", decision.Action)
	}

	// Valid JSON missing required field should retry
	incompletJSON := `{"message":"hello"}`
	decision = decider.Decide(aiwf.DialogContext{
		Step:    "test",
		Output:  incompletJSON,
		Attempt: 1,
	})
	if decision.Action != aiwf.DialogActionRetry {
		t.Errorf("Expected Retry for incomplete JSON, got %v", decision.Action)
	}

	t.Logf("✓ JSONValidationDecider validates JSON structure")
}

func TestChainedDecider(t *testing.T) {
	// Chain: length check + quality check
	lengthDecider := NewLengthCheckDecider(20)
	qualityDecider := NewQualityCheckDecider(func(output any) (bool, string) {
		str := output.(string)
		if len(str) > 100 {
			return true, ""
		}
		return false, "Too short for quality"
	})

	chained := NewChainedDecider(lengthDecider, qualityDecider)

	// Output too short for first check - should stop at length check
	decision := chained.Decide(aiwf.DialogContext{
		Step:    "test",
		Output:  "short",
		Attempt: 1,
	})
	if decision.Action != aiwf.DialogActionRetry {
		t.Errorf("Expected Retry from length check, got %v", decision.Action)
	}

	// Output passes length but fails quality - should get to quality check
	mediumOutput := "This is a medium length string but not quite long enough for quality"
	decision = chained.Decide(aiwf.DialogContext{
		Step:    "test",
		Output:  mediumOutput,
		Attempt: 1,
	})
	if decision.Action != aiwf.DialogActionRetry {
		t.Errorf("Expected Retry from quality check, got %v", decision.Action)
	}

	// Output passes all checks
	longOutput := "This is a very long response that passes all quality checks and should result in a complete decision being returned to the caller"
	decision = chained.Decide(aiwf.DialogContext{
		Step:    "test",
		Output:  longOutput,
		Attempt: 1,
	})
	if decision.Action != aiwf.DialogActionComplete {
		t.Errorf("Expected Complete from chained checks, got %v", decision.Action)
	}

	t.Logf("✓ ChainedDecider combines multiple checks")
}

func TestDefaultDialogDecider(t *testing.T) {
	decider := aiwf.DefaultDialogDecider{}

	decision := decider.Decide(aiwf.DialogContext{
		Step:    "test",
		Output:  "any output",
		Attempt: 1,
	})

	if decision.Action != aiwf.DialogActionComplete {
		t.Errorf("Expected Complete, got %v", decision.Action)
	}

	t.Logf("✓ DefaultDialogDecider always completes")
}
