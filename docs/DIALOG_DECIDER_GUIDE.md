# DialogDecider Implementation Guide

## Overview

`DialogDecider` controls multi-turn dialog flow by deciding what happens after each LLM call in a dialog loop. It implements the strategy pattern for flexible dialog control.

## Interface

```go
type DialogDecider interface {
    Decide(DialogContext) DialogDecision
}

type DialogContext struct {
    Step    string      // Current step/agent name
    Output  any         // LLM output to evaluate
    Trace   *Trace      // Token usage and metadata
    Attempt int         // Current attempt number (1-based)
}

type DialogDecision struct {
    Action   DialogAction  // Continue, Retry, Goto, Stop, Complete
    Feedback string        // For Retry: feedback to include in next prompt
    Target   string        // For Goto: target step name
}

type DialogAction int

const (
    DialogActionContinue = iota
    DialogActionRetry         // Retry current step with feedback
    DialogActionGoto          // Jump to different step
    DialogActionStop          // Stop dialog immediately
    DialogActionComplete      // Mark dialog as complete
)
```

## Built-in Implementations

### 1. DefaultDialogDecider
Always returns `DialogActionComplete` after first pass.
```go
decider := &aiwf.DefaultDialogDecider{}
// Immediately completes dialog without review
```

### 2. ImmediateRetryDecider
Requests retry up to a maximum number of attempts.
```go
decider := openai.NewImmediateRetryDecider(maxAttempts, message)
// Always retries until maxAttempts exceeded

// Example:
decider := openai.NewImmediateRetryDecider(3, "Please refine your response")
decision := decider.Decide(ctx)
// Attempt 1-3: returns Retry
// Attempt 4+: returns Complete
```

### 3. QualityCheckDecider
Validates output using custom check function.
```go
decider := openai.NewQualityCheckDecider(
    func(output any) (bool, string) {
        // Return (passed, feedback)
        // If passed=true, output is acceptable
        // If passed=false, will retry with feedback
        return len(output.(string)) > 50, "Please write more detail"
    },
)

// Example: Ensure response mentions specific topics
decider := openai.NewQualityCheckDecider(
    func(output any) (bool, string) {
        str := output.(string)
        if strings.Contains(strings.ToLower(str), "solution") &&
           strings.Contains(strings.ToLower(str), "recommendation") {
            return true, ""
        }
        return false, "Response must include 'solution' and 'recommendation'"
    },
)
```

### 4. LengthCheckDecider
Ensures output meets minimum length requirement.
```go
decider := openai.NewLengthCheckDecider(minLength)
// Retries if output.String length < minLength

// Example: Require at least 100 characters
decider := openai.NewLengthCheckDecider(100)
decision := decider.Decide(ctx)
```

### 5. JSONValidationDecider
Validates that output is valid JSON matching schema.
```go
schema := map[string]any{
    "message": nil,
    "status": nil,
}
decider := openai.NewJSONValidationDecider(schema)

// Retries if:
// - Output is not valid JSON
// - Required fields are missing
```

### 6. ChainedDecider
Combines multiple deciders in sequence.
```go
decider := openai.NewChainedDecider(
    openai.NewLengthCheckDecider(50),
    openai.NewQualityCheckDecider(checkFn),
    openai.NewJSONValidationDecider(schema),
)

// Checks in order:
// 1. Length check - if fails, returns Retry immediately
// 2. Quality check - if passes length, then checks quality
// 3. JSON validation - if passes quality, validates JSON
// Returns Complete only if all pass
```

## Usage Patterns

### Pattern 1: Simple Single-Turn (Default)
```go
client := openai.NewClient(config)
service := myservice.NewService(client)

result, trace, err := service.Agents().MyAgent.Run(ctx, input)
// Uses DefaultDialogDecider - completes after first LLM call
```

### Pattern 2: Retry Until Quality
```go
// Create decider that validates response quality
decider := openai.NewQualityCheckDecider(func(output any) (bool, string) {
    response := output.(*SupportResponse)
    if response.Escalate && len(response.Message) < 100 {
        return false, "Escalation reasons must be detailed (min 100 chars)"
    }
    return true, ""
})

// Inject into agent
client := openai.NewClient(config)
// TODO: Add SetDialogDecider method to service
service := myservice.NewService(client)
// service.SetDialogDecider(decider)

result, trace, err := service.Agents().SupportBot.Run(ctx, input)
```

### Pattern 3: Multi-Check Pipeline
```go
deciders := openai.NewChainedDecider(
    // First check: output length
    openai.NewLengthCheckDecider(50),
    // Then check: custom quality
    openai.NewQualityCheckDecider(func(output any) (bool, string) {
        response := output.(*SupportResponse)
        // Ensure at least one action suggested
        if len(response.Actions) == 0 && !response.Escalate {
            return false, "Must either suggest actions or escalate"
        }
        return true, ""
    }),
    // Finally check: JSON structure
    openai.NewJSONValidationDecider(map[string]any{
        "message": nil,
        "status": nil,
    }),
)

// service.SetDialogDecider(deciders)
```

## Custom Implementation

Create your own DialogDecider for domain-specific logic:

```go
type DomainSpecificDecider struct {
    Rules map[string]string
}

func (d *DomainSpecificDecider) Decide(ctx aiwf.DialogContext) aiwf.DialogDecision {
    if ctx.Attempt > 5 {
        return aiwf.DialogDecision{Action: aiwf.DialogActionComplete}
    }

    // Your domain logic here
    switch ctx.Step {
    case "validate_email":
        // Special handling for email validation
        if isValidEmail(ctx.Output) {
            return aiwf.DialogDecision{Action: aiwf.DialogActionComplete}
        }
        return aiwf.DialogDecision{
            Action: aiwf.DialogActionRetry,
            Feedback: "Email format invalid. Please provide valid email.",
        }
    }

    return aiwf.DialogDecision{Action: aiwf.DialogActionComplete}
}
```

## Integration with Dialog Flow

The DialogDecider is called after each LLM invocation:

```
Agent.RunDialog()
    ↓
[Loop]
    ├─ ThreadManager.Start/Continue
    ├─ Agent.CallModel(input)
    ├─ Get output and trace
    ├─ Create DialogContext
    ├─ Call DialogDecider.Decide()
    └─ Based on Decision:
        ├─ Continue → next iteration (no action)
        ├─ Retry → add feedback to messages, loop
        ├─ Goto → switch to target step
        ├─ Stop → break loop
        └─ Complete → exit loop with result
    ↓
ThreadManager.Close()
    ↓
Return result
```

## Testing

All deciders are unit tested in `dialog_decider_test.go`:

```bash
go test ./providers/openai -run "DialogDecider" -v
```

## Best Practices

1. **Keep checks simple**: Each decider should have one clear purpose
2. **Use ChainedDecider for complex logic**: Compose simple deciders
3. **Provide helpful feedback**: Include specific guidance for retry
4. **Set reasonable MaxAttempts**: Prevent infinite loops (typically 3-5)
5. **Test with real data**: Validate checks with actual LLM outputs

## Future Enhancements

- [ ] LLMDialogDecider: Use another LLM to evaluate responses
- [ ] RuleBasedDialogDecider: YAML-configurable rule engine
- [ ] MetricsDecider: Track and adjust based on success metrics
- [ ] AdaptiveDecider: Learn from feedback over time
