# Workflow Engine Guide

## Overview

The AIWF Workflow Engine enables execution of complex multi-step workflows with:
- **DAG (Directed Acyclic Graph)** validation
- **Topological sorting** for dependency resolution
- **Parallel branch support** with join operations
- **Error handling** and **automatic retry logic**
- **Result tracking** for each step

## Architecture

### Core Concepts

```
WorkflowStep       - Single executable unit
  ├─ GetName()      - Step identifier
  ├─ Execute()      - Main logic
  └─ GetDependencies() - Required inputs

WorkflowDefinition - DAG structure
  ├─ Steps           - Map of WorkflowSteps
  └─ Dependencies    - Dependency graph

WorkflowExecutor   - Execution engine
  ├─ Validate()      - Check DAG validity
  ├─ Execute()       - Run workflow
  └─ GetResults()    - Collect outputs
```

### Execution Flow

```
1. Define workflow (add steps with dependencies)
   ↓
2. Validate DAG (detect cycles, check dependencies)
   ↓
3. Topological sort (determine execution order)
   ↓
4. Execute steps (with retry logic):
   - Check dependencies satisfied
   - Pass previous step output as input
   - Retry on failure (up to maxRetries)
   ↓
5. Collect results (success/failure/attempts)
```

## Usage

### 1. Creating Workflow Steps

#### Simple Step

```go
step := NewSimpleStep("download",
    func(ctx context.Context, input any) (any, error) {
        data := fetchData()
        return data, nil
    })
```

#### Step with Dependencies

```go
processStep := NewSimpleStep("process",
    func(ctx context.Context, input any) (any, error) {
        downloadedData := input.([]byte)
        return process(downloadedData), nil
    }).
    WithDependencies("download")
```

#### Custom Step Implementation

```go
type MyStep struct {
    name string
    deps []string
}

func (s *MyStep) GetName() string {
    return s.name
}

func (s *MyStep) Execute(ctx context.Context, input any) (any, error) {
    // Your custom logic
    return result, nil
}

func (s *MyStep) GetDependencies() []string {
    return s.deps
}
```

### 2. Building Workflows

#### Linear Workflow

```go
wf := NewWorkflowDefinition("linear")

step1 := NewSimpleStep("fetch", fetchFn)
step2 := NewSimpleStep("parse", parseFn).WithDependencies("fetch")
step3 := NewSimpleStep("store", storeFn).WithDependencies("parse")

wf.AddStep(step1)
wf.AddStep(step2)
wf.AddStep(step3)
```

#### Parallel Workflow with Join

```go
wf := NewWorkflowDefinition("parallel")

// Start
start := NewSimpleStep("start", startFn)

// Parallel branches
fetch1 := NewSimpleStep("fetch1", fetch1Fn).WithDependencies("start")
fetch2 := NewSimpleStep("fetch2", fetch2Fn).WithDependencies("start")

// Join step (receives map of previous results)
join := NewSimpleStep("merge",
    func(ctx context.Context, input any) (any, error) {
        results := input.(map[string]any)
        data1 := results["fetch1"]
        data2 := results["fetch2"]
        return merge(data1, data2), nil
    }).
    WithDependencies("fetch1", "fetch2")

wf.AddStep(start)
wf.AddStep(fetch1)
wf.AddStep(fetch2)
wf.AddStep(join)
```

### 3. Executing Workflows

```go
// Create executor
executor, err := NewWorkflowExecutor(wf)
if err != nil {
    log.Fatal(err)
}

// Configure retry behavior
executor.SetMaxRetries(3)

// Execute
ctx := context.Background()
results, err := executor.Execute(ctx, initialInput)
if err != nil {
    log.Printf("Workflow failed: %v", err)
}

// Collect results
for stepName, result := range results {
    if result.Error != nil {
        log.Printf("%s failed: %v", stepName, result.Error)
    } else {
        log.Printf("%s succeeded: %v", stepName, result.Output)
    }
}
```

## Workflow Patterns

### Pattern 1: Sequential Processing

```
Input → Step1 → Step2 → Step3 → Output

step1 result fed to step2
step2 result fed to step3
```

**Use case**: ETL pipelines, data transformation chains

```go
wf := NewWorkflowDefinition("etl")
wf.AddStep(NewSimpleStep("extract", extractFn))
wf.AddStep(NewSimpleStep("transform", transformFn).WithDependencies("extract"))
wf.AddStep(NewSimpleStep("load", loadFn).WithDependencies("transform"))
```

### Pattern 2: Parallel Processing with Join

```
        ┌─→ Step2 ─┐
Step1 ─→            → Step4
        └─→ Step3 ─┘

Step2 and Step3 run in parallel
Step4 receives outputs of both as map
```

**Use case**: Map-reduce, multi-source aggregation

```go
split := NewSimpleStep("split", splitFn)
proc1 := NewSimpleStep("proc1", proc1Fn).WithDependencies("split")
proc2 := NewSimpleStep("proc2", proc2Fn).WithDependencies("split")
join := NewSimpleStep("join", joinFn).WithDependencies("proc1", "proc2")
```

### Pattern 3: Conditional Branching (via dependencies)

```
Step1 → Step2A → Step4
     → Step2B → Step4

Step2A and Step2B both depend on Step1
Step4 depends on whichever Step2 completes
```

Note: For true conditional logic, use a single decision step.

### Pattern 4: Fan-out/Fan-in

```
        ├─→ Task1 ─┐
Input ─→├─→ Task2 ├─→ Aggregate
        └─→ Task3 ─┘

Many parallel tasks → combine results
```

**Use case**: Processing multiple items, batch operations

## Error Handling

### Step-level Errors

```go
step := NewSimpleStep("validate",
    func(ctx context.Context, input any) (any, error) {
        if !isValid(input) {
            return nil, fmt.Errorf("invalid input: %v", input)
        }
        return input, nil
    })
```

### Retry Logic

```go
executor.SetMaxRetries(3)
// Steps automatically retry up to maxRetries times

result := executor.Execute(ctx, input)
// For each failed step:
// - First attempt fails
// - Retried 2 more times
// - If still failing, stops workflow
```

### Dependency Error Propagation

```go
// If Step1 fails:
// - Step1.Error is recorded
// - Step2 (depends on Step1) is NOT executed
// - Workflow stops, returns error
```

## Advanced Features

### DAG Validation

```go
// Automatic cycle detection
wf := NewWorkflowDefinition("check")
wf.AddStep(step1)
wf.AddStep(step2.WithDependencies("step1"))
wf.AddStep(step1.WithDependencies("step2")) // Error! Cycle detected

executor, err := NewWorkflowExecutor(wf)
// err != nil: "workflow contains cycles"
```

### Topological Ordering

```go
order, err := wf.GetTopologicalOrder()
// Returns: [step1, step2, step3, ...]
// Guarantees all dependencies come before dependents
```

### Result Inspection

```go
result := results["step_name"]

// Available fields:
result.StepName     // "step_name"
result.Output       // any - the step output
result.Error        // error - if failed
result.Duration     // int64 - nanoseconds
result.Attempts     // int - how many retries

if result.Error != nil {
    log.Printf("Step failed after %d attempts", result.Attempts)
}
```

## Integration with Dialogs

Workflows can combine with DialogDeciders for approval flows:

```go
// Step that makes LLM decision
decisionStep := NewSimpleStep("approve",
    func(ctx context.Context, input any) (any, error) {
        // Use DialogDecider to validate
        decider := openai.NewQualityCheckDecider(checkFn)
        decision := decider.Decide(aiwf.DialogContext{
            Step:    "review",
            Output:  input,
            Attempt: 1,
        })

        if decision.Action == aiwf.DialogActionRetry {
            return nil, fmt.Errorf(decision.Feedback)
        }
        return input, nil
    })
```

## Performance Considerations

### Sequential vs Parallel

- **Sequential**: Each step waits for previous
  - Simple logic, easier to debug
  - Slower for independent steps

- **Parallel**: Independent branches run concurrently
  - Faster for multi-source workflows
  - Requires careful dependency management

### Memory Usage

- Results stored in-memory (map)
- For large workflows: implement streaming results
- Consider cleanup between iterations

### Timeout Management

```go
ctx, cancel := context.WithTimeout(
    context.Background(),
    30*time.Second,
)
defer cancel()

results, err := executor.Execute(ctx, input)
// Fails if any step exceeds total timeout
```

## Testing

### Unit Testing Steps

```go
func TestMyStep(t *testing.T) {
    step := NewSimpleStep("test", func(ctx context.Context, input any) (any, error) {
        return "result", nil
    })

    result, err := step.Execute(context.Background(), "input")
    assert.NoError(t, err)
    assert.Equal(t, "result", result)
}
```

### Integration Testing Workflows

```go
func TestWorkflow(t *testing.T) {
    wf := NewWorkflowDefinition("test")
    // Add steps...

    executor, _ := NewWorkflowExecutor(wf)
    results, err := executor.Execute(context.Background(), nil)

    assert.NoError(t, err)
    assert.Nil(t, results["step1"].Error)
    assert.NotNil(t, results["step1"].Output)
}
```

## Examples

See `runtime/go/aiwf/workflow_test.go` for:
- Basic sequential execution
- Parallel branching
- Error handling
- Retry logic
- Complex DAGs
- Performance benchmarks

## API Reference

### WorkflowStep

```go
type WorkflowStep interface {
    GetName() string
    Execute(context.Context, any) (any, error)
    GetDependencies() []string
}
```

### WorkflowDefinition

```go
func NewWorkflowDefinition(name string) *WorkflowDefinition
func (w *WorkflowDefinition) AddStep(step WorkflowStep) error
func (w *WorkflowDefinition) ValidateDAG() error
func (w *WorkflowDefinition) GetTopologicalOrder() ([]string, error)
```

### WorkflowExecutor

```go
func NewWorkflowExecutor(def *WorkflowDefinition) (*WorkflowExecutor, error)
func (e *WorkflowExecutor) SetMaxRetries(max int)
func (e *WorkflowExecutor) Execute(ctx context.Context, input any) (map[string]*WorkflowStepResult, error)
func (e *WorkflowExecutor) GetResult(name string) (*WorkflowStepResult, bool)
```

### SimpleStep

```go
func NewSimpleStep(name string, fn func(context.Context, any) (any, error)) *SimpleStep
func (s *SimpleStep) WithDependencies(deps ...string) *SimpleStep
```

## Troubleshooting

### "workflow contains cycles"

**Cause**: Circular dependencies in steps
**Solution**: Review dependencies, ensure DAG structure

```
✗ A depends on B, B depends on C, C depends on A
✓ A depends on B, B depends on C
```

### "dependency X not found"

**Cause**: Referenced dependency not added to workflow
**Solution**: Ensure all dependencies are added as steps first

```go
// Wrong:
wf.AddStep(step2.WithDependencies("step1")) // step1 not added yet!

// Right:
wf.AddStep(step1)
wf.AddStep(step2.WithDependencies("step1"))
```

### Step receives wrong input

**Cause**: Multiple dependencies - need to handle map
**Solution**: Check GetDependencies() returns all needed steps

```go
// For multiple deps, input is map[string]any
func(ctx context.Context, input any) (any, error) {
    results := input.(map[string]any)
    val1 := results["dep1"]
    val2 := results["dep2"]
    return combine(val1, val2), nil
}
```

## Roadmap

- [ ] Streaming results for memory efficiency
- [ ] Step-level timeouts
- [ ] Dynamic DAG modification
- [ ] Workflow visualization
- [ ] Metrics and observability
- [ ] Distributed execution
