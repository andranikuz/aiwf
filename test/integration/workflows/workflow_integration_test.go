package workflows

import (
	"context"
	"fmt"
	"testing"

	"github.com/andranikuz/aiwf/providers/openai"
	"github.com/andranikuz/aiwf/runtime/go/aiwf"
)

func TestWorkflowEngine_BasicSequential(t *testing.T) {
	wf := aiwf.NewWorkflowDefinition("sequential")

	// Step 1: Generate initial content
	step1 := aiwf.NewSimpleStep("generate",
		func(ctx context.Context, input any) (any, error) {
			return "generated content", nil
		})

	// Step 2: Process content
	step2 := aiwf.NewSimpleStep("process",
		func(ctx context.Context, input any) (any, error) {
			content := input.(string)
			return fmt.Sprintf("processed: %s", content), nil
		}).WithDependencies("generate")

	// Step 3: Validate content
	step3 := aiwf.NewSimpleStep("validate",
		func(ctx context.Context, input any) (any, error) {
			content := input.(string)
			if len(content) < 5 {
				return nil, fmt.Errorf("content too short")
			}
			return content, nil
		}).WithDependencies("process")

	wf.AddStep(step1)
	wf.AddStep(step2)
	wf.AddStep(step3)

	executor, err := aiwf.NewWorkflowExecutor(wf)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	results, err := executor.Execute(context.Background(), nil)
	if err != nil {
		t.Fatalf("Workflow execution failed: %v", err)
	}

	// Verify results
	if results["generate"].Output != "generated content" {
		t.Errorf("Step 1 output wrong")
	}

	if !contains(results["process"].Output.(string), "processed") {
		t.Errorf("Step 2 output wrong")
	}

	t.Logf("✓ Sequential workflow completed successfully")
}

func TestWorkflowEngine_ParallelWithDialogDecider(t *testing.T) {
	wf := aiwf.NewWorkflowDefinition("parallel_dialog")

	// Fetch multiple data sources in parallel
	fetch1 := aiwf.NewSimpleStep("fetch_data1",
		func(ctx context.Context, input any) (any, error) {
			return map[string]any{
				"source": "source1",
				"data":   []int{1, 2, 3},
			}, nil
		})

	fetch2 := aiwf.NewSimpleStep("fetch_data2",
		func(ctx context.Context, input any) (any, error) {
			return map[string]any{
				"source": "source2",
				"data":   []int{4, 5, 6},
			}, nil
		})

	// Aggregate results
	aggregate := aiwf.NewSimpleStep("aggregate",
		func(ctx context.Context, input any) (any, error) {
			results := input.(map[string]any)
			data1 := results["fetch_data1"].(map[string]any)
			data2 := results["fetch_data2"].(map[string]any)

			return map[string]any{
				"all_data": []any{data1, data2},
				"status":   "aggregated",
			}, nil
		}).WithDependencies("fetch_data1", "fetch_data2")

	// Validate with DialogDecider
	validate := aiwf.NewSimpleStep("validate",
		func(ctx context.Context, input any) (any, error) {
			aggregated := input.(map[string]any)

			// Use DialogDecider to validate
			decider := openai.NewQualityCheckDecider(
				func(output any) (bool, string) {
					agg := output.(map[string]any)
					if agg["status"] != "aggregated" {
						return false, "Status must be 'aggregated'"
					}
					if len(agg["all_data"].([]any)) != 2 {
						return false, "Must have 2 data sources"
					}
					return true, ""
				})

			decision := decider.Decide(aiwf.DialogContext{
				Step:    "validate",
				Output:  aggregated,
				Attempt: 1,
			})

			if decision.Action == aiwf.DialogActionRetry {
				feedback := decision.Feedback
				return nil, fmt.Errorf("validation failed: %s", feedback)
			}

			return aggregated, nil
		}).WithDependencies("aggregate")

	wf.AddStep(fetch1)
	wf.AddStep(fetch2)
	wf.AddStep(aggregate)
	wf.AddStep(validate)

	executor, _ := aiwf.NewWorkflowExecutor(wf)
	results, err := executor.Execute(context.Background(), nil)

	if err != nil {
		t.Fatalf("Workflow failed: %v", err)
	}

	if results["validate"].Error != nil {
		t.Fatalf("Validation failed: %v", results["validate"].Error)
	}

	validatedResult := results["validate"].Output.(map[string]any)
	if validatedResult["status"] != "aggregated" {
		t.Errorf("Validated result status wrong")
	}

	t.Logf("✓ Parallel workflow with DialogDecider validation passed")
}

func TestWorkflowEngine_ErrorRecoveryWithRetry(t *testing.T) {
	wf := aiwf.NewWorkflowDefinition("retry")

	attempts := 0
	step := aiwf.NewSimpleStep("flaky",
		func(ctx context.Context, input any) (any, error) {
			attempts++
			if attempts < 2 {
				return nil, fmt.Errorf("temporary failure")
			}
			return "success after retry", nil
		})

	wf.AddStep(step)

	executor, _ := aiwf.NewWorkflowExecutor(wf)
	executor.SetMaxRetries(3)

	results, err := executor.Execute(context.Background(), nil)

	if err != nil {
		t.Fatalf("Workflow failed: %v", err)
	}

	result := results["flaky"]
	if result.Output != "success after retry" {
		t.Errorf("Expected success after retry, got %v", result.Output)
	}

	if result.Attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", result.Attempts)
	}

	t.Logf("✓ Retry logic recovered from temporary failure")
}

func TestWorkflowEngine_DAGValidation(t *testing.T) {
	wf := aiwf.NewWorkflowDefinition("invalid")

	step1 := aiwf.NewSimpleStep("step1", nil)
	step2 := aiwf.NewSimpleStep("step2", nil)

	wf.AddStep(step1)
	wf.AddStep(step2)

	// Manually create a cycle
	wf.Dependencies["step1"] = []string{"step2"}
	wf.Dependencies["step2"] = []string{"step1"}

	_, err := aiwf.NewWorkflowExecutor(wf)
	if err == nil {
		t.Fatal("Expected DAG validation to fail")
	}

	t.Logf("✓ DAG validation caught cycle: %v", err)
}

func TestWorkflowEngine_TopologicalOrdering(t *testing.T) {
	wf := aiwf.NewWorkflowDefinition("topo")

	// Create complex DAG
	step1 := aiwf.NewSimpleStep("step1", nil)
	step2 := aiwf.NewSimpleStep("step2", nil).WithDependencies("step1")
	step3 := aiwf.NewSimpleStep("step3", nil).WithDependencies("step1")
	step4 := aiwf.NewSimpleStep("step4", nil).WithDependencies("step2", "step3")
	step5 := aiwf.NewSimpleStep("step5", nil).WithDependencies("step4")

	wf.AddStep(step1)
	wf.AddStep(step2)
	wf.AddStep(step3)
	wf.AddStep(step4)
	wf.AddStep(step5)

	order, err := wf.GetTopologicalOrder()
	if err != nil {
		t.Fatalf("Topological sort failed: %v", err)
	}

	// Verify constraints
	pos := make(map[string]int)
	for i, name := range order {
		pos[name] = i
	}

	constraints := [][2]string{
		{"step1", "step2"},
		{"step1", "step3"},
		{"step2", "step4"},
		{"step3", "step4"},
		{"step4", "step5"},
	}

	for _, constraint := range constraints {
		if pos[constraint[0]] >= pos[constraint[1]] {
			t.Errorf("%s should come before %s in order %v", constraint[0], constraint[1], order)
		}
	}

	t.Logf("✓ Topological ordering is correct: %v", order)
}

func contains(s string, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
