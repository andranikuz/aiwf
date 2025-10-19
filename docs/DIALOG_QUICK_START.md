# Dialog Quick Start Guide

## Overview

The AIWF Dialog system enables multi-turn conversations with LLMs using:
- **ThreadManager**: Manages conversation thread lifecycle
- **DialogDecider**: Controls multi-turn flow with validation strategies

## 5-Minute Setup

### 1. Define Your Dialog Type

```yaml
# templates/my_dialog.yaml
version: 0.3-proposal

assistants:
  my_agent:
    use: openai
    model: gpt-4o
    thread:
      use: default
      strategy: append
    dialog:
      max_rounds: 10
    system_prompt: "You are a helpful assistant."
    input_type: MyInput
    output_type: MyOutput

types:
  MyInput:
    message: string
    context: string

  MyOutput:
    response: string
    confidence: int
```

### 2. Generate SDK

```bash
go run ./cmd/aiwf sdk -f templates/my_dialog.yaml \
  -o ./generated/my_dialog \
  --package my_dialog_sdk
```

### 3. Use in Code

```go
package main

import (
    "context"
    openai "github.com/andranikuz/aiwf/providers/openai"
    "github.com/andranikuz/aiwf/runtime/go/aiwf"
    myservice "github.com/andranikuz/aiwf/generated/my_dialog"
)

func main() {
    // Setup
    apiKey := os.Getenv("OPENAI_API_KEY")
    client, _ := openai.NewClient(openai.ClientConfig{APIKey: apiKey})

    // Add ThreadManager for multi-turn
    threadMgr := openai.NewInMemoryThreadManager()
    client = client.WithThreadManager(threadMgr)

    // Create service
    service := myservice.NewService(client)

    // Simple dialog
    input := myservice.MyInput{
        Message: "Hello, what's your name?",
        Context: "greeting",
    }

    ctx := context.Background()
    result, trace, err := service.Agents().MyAgent.Run(ctx, input)

    // Result contains response with traced tokens
    println(result.Response)
}
```

## Using DialogDecider

DialogDecider validates responses and decides whether to retry or complete:

### Example 1: Length Validation

```go
// Require at least 100 characters
decider := openai.NewLengthCheckDecider(100)

decision := decider.Decide(aiwf.DialogContext{
    Step:    "my_agent",
    Output:  result,
    Attempt: 1,
})

if decision.Action == aiwf.DialogActionRetry {
    // Retry with decision.Feedback
    println("Too short, retrying:", decision.Feedback)
}
```

### Example 2: Custom Quality Check

```go
decider := openai.NewQualityCheckDecider(
    func(output any) (bool, string) {
        result := output.(*myservice.MyOutput)

        // Check confidence threshold
        if result.Confidence < 70 {
            return false, "Confidence too low, please try again with more certainty"
        }

        return true, ""
    },
)
```

### Example 3: Multi-Check Chain

```go
// Length → Quality → JSON validation
decider := openai.NewChainedDecider(
    openai.NewLengthCheckDecider(50),
    openai.NewQualityCheckDecider(customCheck),
    openai.NewJSONValidationDecider(schema),
)

decision := decider.Decide(ctx)
// Stops at first failure, otherwise returns Complete
```

## Dialog Flow

```
Input
  ↓
Start Thread (ThreadManager.Start)
  ↓
Loop (max_rounds times):
  ├─ Call LLM (Agent.CallModel)
  ├─ Get response + token usage
  ├─ Create DialogContext (step, output, attempt)
  ├─ Validate with DialogDecider (CustomCheck.Decide)
  └─ Based on decision:
      ├─ Retry → append feedback to messages, loop
      ├─ Continue → move to next step
      ├─ Complete → exit loop
      └─ Goto/Stop → not yet implemented
  ↓
Close Thread (ThreadManager.Close)
  ↓
Return result + Trace
```

## Available DialogDeciders

| Decider | Purpose | Use Case |
|---------|---------|----------|
| DefaultDialogDecider | Always complete | Single-turn dialogs |
| ImmediateRetryDecider | Retry N times | Refinement loops |
| LengthCheckDecider | Min length | Ensure detail |
| QualityCheckDecider | Custom validation | Domain logic |
| JSONValidationDecider | Schema check | Structured data |
| ChainedDecider | Combine multiple | Complex validation |

## Key Concepts

### ThreadManager
Manages conversation thread lifecycle:
- **Start**: Create new thread for conversation
- **Continue**: Add feedback message to thread
- **Close**: Clean up thread resources

```go
tm := openai.NewInMemoryThreadManager()

// Create new thread
state, _ := tm.Start(ctx, "my_agent", binding)

// Send feedback
_ = tm.Continue(ctx, state, "Please clarify...")

// Clean up
_ = tm.Close(ctx, state)
```

### DialogContext
Information available to DialogDecider:
```go
type DialogContext struct {
    Step    string   // Agent/step name
    Output  any      // LLM output to evaluate
    Trace   *Trace   // Token usage + metadata
    Attempt int      // Current attempt (1-based)
}
```

### DialogDecision
What happens next:
```go
type DialogDecision struct {
    Action   DialogAction // Retry/Continue/Complete/Goto/Stop
    Feedback string       // For Retry: shown to LLM
    Target   string       // For Goto: target step
}
```

## Testing Your Dialog

```go
func TestMyDialog(t *testing.T) {
    client, _ := openai.NewClient(config)
    service := myservice.NewService(client)

    input := myservice.MyInput{Message: "test"}
    result, trace, err := service.Agents().MyAgent.Run(ctx, input)

    if err != nil {
        t.Fatal(err)
    }

    // Verify result
    assert.NotEmpty(t, result.Response)
    assert.Greater(t, result.Confidence, 0)

    // Check token usage
    println("Tokens:", trace.Usage.Total)
}
```

## Common Patterns

### Pattern 1: Simple Single-Turn
```go
result, trace, _ := service.Agents().MyAgent.Run(ctx, input)
// Uses default DialogDecider, completes after one pass
```

### Pattern 2: Retry Until Quality
```go
decider := openai.NewQualityCheckDecider(customCheck)
// Integrate into service (pending API)
// service.Agents().MyAgent.SetDialogDecider(decider)
result, trace, _ := service.Agents().MyAgent.Run(ctx, input)
```

### Pattern 3: Multi-Check Validation
```go
decider := openai.NewChainedDecider(
    openai.NewLengthCheckDecider(100),
    openai.NewQualityCheckDecider(check1),
    openai.NewQualityCheckDecider(check2),
)
```

## Troubleshooting

### Thread Tests Fail
Ensure `OPENAI_API_KEY` is set:
```bash
export OPENAI_API_KEY=sk-...
go test ./test/integration/dialogs -v
```

### Dialog Response Too Short
Use LengthCheckDecider:
```go
decider := openai.NewLengthCheckDecider(200)  // Min 200 chars
```

### Invalid JSON Response
Use JSONValidationDecider:
```go
decider := openai.NewJSONValidationDecider(map[string]any{
    "response": nil,
    "status": nil,
})
```

## Next Steps

- [ ] Run dialog tests: `go test ./test/integration/dialogs/... -v`
- [ ] Read DIALOG_DECIDER_GUIDE.md for advanced usage
- [ ] Integrate DialogDecider into Service
- [ ] Implement custom DialogDeciders for your domain
- [ ] Deploy ThreadManager to production (OpenAI Threads API)

## Resources

- **DIALOG_IMPLEMENTATION_PROGRESS.md** - Architecture and status
- **DIALOG_DECIDER_GUIDE.md** - Detailed API documentation
- **test/integration/dialogs/** - Working examples
- **providers/openai/dialog_decider.go** - Source code
