package aiwf

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func TestSimpleStep(t *testing.T) {
	step := NewSimpleStep("test_step", func(ctx context.Context, input any) (any, error) {
		return fmt.Sprintf("processed: %v", input), nil
	})

	if step.GetName() != "test_step" {
		t.Errorf("Expected name 'test_step', got %q", step.GetName())
	}

	result, err := step.Execute(context.Background(), "input")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := "processed: input"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}

	t.Logf("✓ SimpleStep works correctly")
}

func TestWorkflowDefinition_BasicChain(t *testing.T) {
	wf := NewWorkflowDefinition("basic_chain")

	step1 := NewSimpleStep("step1", func(ctx context.Context, input any) (any, error) {
		return "result1", nil
	})

	step2 := NewSimpleStep("step2", func(ctx context.Context, input any) (any, error) {
		return fmt.Sprintf("%v -> result2", input), nil
	}).WithDependencies("step1")

	if err := wf.AddStep(step1); err != nil {
		t.Fatalf("Failed to add step1: %v", err)
	}

	if err := wf.AddStep(step2); err != nil {
		t.Fatalf("Failed to add step2: %v", err)
	}

	if err := wf.ValidateDAG(); err != nil {
		t.Fatalf("DAG validation failed: %v", err)
	}

	order, err := wf.GetTopologicalOrder()
	if err != nil {
		t.Fatalf("Failed to get topological order: %v", err)
	}

	if len(order) != 2 {
		t.Errorf("Expected 2 steps, got %d", len(order))
	}

	if order[0] != "step1" || order[1] != "step2" {
		t.Errorf("Wrong order: %v", order)
	}

	t.Logf("✓ Workflow DAG correctly ordered: %v", order)
}

func TestWorkflowDefinition_MultipleDependencies(t *testing.T) {
	wf := NewWorkflowDefinition("multi_dep")

	step1 := NewSimpleStep("step1", func(ctx context.Context, input any) (any, error) {
		return "result1", nil
	})

	step2 := NewSimpleStep("step2", func(ctx context.Context, input any) (any, error) {
		return "result2", nil
	})

	step3 := NewSimpleStep("step3", func(ctx context.Context, input any) (any, error) {
		return "result3", nil
	}).WithDependencies("step1", "step2")

	wf.AddStep(step1)
	wf.AddStep(step2)
	wf.AddStep(step3)

	order, err := wf.GetTopologicalOrder()
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}

	// step1 and step2 should be before step3
	step3Idx := -1
	for i, s := range order {
		if s == "step3" {
			step3Idx = i
			break
		}
	}

	if step3Idx != 2 {
		t.Errorf("step3 should be last, but order is %v", order)
	}

	t.Logf("✓ Multiple dependencies handled correctly: %v", order)
}

func TestWorkflowDefinition_DetectsCycle(t *testing.T) {
	wf := NewWorkflowDefinition("cyclic")

	// We'll manually create a cycle by manipulating dependencies
	step1 := NewSimpleStep("step1", nil)
	step2 := NewSimpleStep("step2", nil)

	wf.AddStep(step1)
	wf.AddStep(step2)

	// Manually create cycle: step1 depends on step2, step2 depends on step1
	wf.Dependencies["step1"] = []string{"step2"}
	wf.Dependencies["step2"] = []string{"step1"}

	err := wf.ValidateDAG()
	if err == nil {
		t.Fatal("Expected cycle detection to fail, but it passed")
	}

	if !contains(err.Error(), "cycle") {
		t.Errorf("Wrong error message: %q", err.Error())
	}

	t.Logf("✓ Cycle detection works: %v", err)
}

func TestWorkflowExecutor_SimpleExecution(t *testing.T) {
	wf := NewWorkflowDefinition("simple_exec")

	step1 := NewSimpleStep("step1", func(ctx context.Context, input any) (any, error) {
		return "result1", nil
	})

	step2 := NewSimpleStep("step2", func(ctx context.Context, input any) (any, error) {
		// Receives output of step1
		return fmt.Sprintf("step2: %v", input), nil
	}).WithDependencies("step1")

	wf.AddStep(step1)
	wf.AddStep(step2)

	executor, err := NewWorkflowExecutor(wf)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	ctx := context.Background()
	results, err := executor.Execute(ctx, "initial")
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	// Check step1 result
	result1, ok := results["step1"]
	if !ok {
		t.Fatal("step1 result not found")
	}
	if result1.Output != "result1" {
		t.Errorf("step1: expected 'result1', got %q", result1.Output)
	}

	// Check step2 result
	result2, ok := results["step2"]
	if !ok {
		t.Fatal("step2 result not found")
	}
	expected := "step2: result1"
	if result2.Output != expected {
		t.Errorf("step2: expected %q, got %q", expected, result2.Output)
	}

	t.Logf("✓ Sequential execution works correctly")
}

func TestWorkflowExecutor_ParallelBranches(t *testing.T) {
	wf := NewWorkflowDefinition("parallel")

	// Start step
	start := NewSimpleStep("start", func(ctx context.Context, input any) (any, error) {
		return "started", nil
	})

	// Two parallel branches
	branch1 := NewSimpleStep("branch1", func(ctx context.Context, input any) (any, error) {
		return "branch1_result", nil
	}).WithDependencies("start")

	branch2 := NewSimpleStep("branch2", func(ctx context.Context, input any) (any, error) {
		return "branch2_result", nil
	}).WithDependencies("start")

	// Join step that combines results
	join := NewSimpleStep("join", func(ctx context.Context, input any) (any, error) {
		results := input.(map[string]any)
		return fmt.Sprintf("joined: %v, %v", results["branch1"], results["branch2"]), nil
	}).WithDependencies("branch1", "branch2")

	wf.AddStep(start)
	wf.AddStep(branch1)
	wf.AddStep(branch2)
	wf.AddStep(join)

	executor, _ := NewWorkflowExecutor(wf)
	ctx := context.Background()
	results, err := executor.Execute(ctx, nil)

	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	joinResult := results["join"]
	if joinResult.Error != nil {
		t.Fatalf("Join failed: %v", joinResult.Error)
	}

	expected := "joined: branch1_result, branch2_result"
	if joinResult.Output != expected {
		t.Errorf("Expected %q, got %q", expected, joinResult.Output)
	}

	t.Logf("✓ Parallel branches with join work correctly")
}

func TestWorkflowExecutor_ErrorHandling(t *testing.T) {
	wf := NewWorkflowDefinition("error_handling")

	step1 := NewSimpleStep("step1", func(ctx context.Context, input any) (any, error) {
		return "result1", nil
	})

	step2 := NewSimpleStep("step2", func(ctx context.Context, input any) (any, error) {
		return nil, fmt.Errorf("step2 failed")
	}).WithDependencies("step1")

	step3 := NewSimpleStep("step3", func(ctx context.Context, input any) (any, error) {
		return "result3", nil
	}).WithDependencies("step2")

	wf.AddStep(step1)
	wf.AddStep(step2)
	wf.AddStep(step3)

	executor, _ := NewWorkflowExecutor(wf)
	ctx := context.Background()
	results, err := executor.Execute(ctx, nil)

	if err == nil {
		t.Fatal("Expected error, but execution succeeded")
	}

	// step1 should have succeeded
	if result1, ok := results["step1"]; !ok || result1.Error != nil {
		t.Fatal("step1 should have succeeded")
	}

	// step2 should have failed
	if result2, ok := results["step2"]; !ok || result2.Error == nil {
		t.Fatal("step2 should have failed")
	}

	// step3 should not have been executed (due to step2 failure)
	if _, ok := results["step3"]; ok {
		t.Fatal("step3 should not have been executed")
	}

	t.Logf("✓ Error handling stops workflow correctly")
}

func TestWorkflowExecutor_Retry(t *testing.T) {
	wf := NewWorkflowDefinition("retry_test")

	attempts := 0
	step := NewSimpleStep("retry_step", func(ctx context.Context, input any) (any, error) {
		attempts++
		if attempts < 2 {
			return nil, fmt.Errorf("retry me")
		}
		return "success", nil
	})

	wf.AddStep(step)

	executor, _ := NewWorkflowExecutor(wf)
	executor.SetMaxRetries(3)

	ctx := context.Background()
	results, err := executor.Execute(ctx, nil)

	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	result := results["retry_step"]
	if result.Output != "success" {
		t.Errorf("Expected success, got %q", result.Output)
	}

	if result.Attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", result.Attempts)
	}

	t.Logf("✓ Retry logic works (recovered after %d attempts)", result.Attempts)
}

func TestWorkflowExecutor_ComplexDAG(t *testing.T) {
	// Complex workflow:
	//        step1
	//       /      \
	//    step2    step3
	//      |   \   /
	//      |   step4
	//       \ /
	//      step5

	wf := NewWorkflowDefinition("complex")

	step1 := NewSimpleStep("step1", func(ctx context.Context, input any) (any, error) {
		return 1, nil
	})
	step2 := NewSimpleStep("step2", func(ctx context.Context, input any) (any, error) {
		return 2, nil
	}).WithDependencies("step1")
	step3 := NewSimpleStep("step3", func(ctx context.Context, input any) (any, error) {
		return 3, nil
	}).WithDependencies("step1")
	step4 := NewSimpleStep("step4", func(ctx context.Context, input any) (any, error) {
		results := input.(map[string]any)
		return fmt.Sprintf("%v+%v=7", results["step2"], results["step3"]), nil
	}).WithDependencies("step2", "step3")
	step5 := NewSimpleStep("step5", func(ctx context.Context, input any) (any, error) {
		results := input.(map[string]any)
		return fmt.Sprintf("final: %v, %v", results["step2"], results["step4"]), nil
	}).WithDependencies("step2", "step4")

	wf.AddStep(step1)
	wf.AddStep(step2)
	wf.AddStep(step3)
	wf.AddStep(step4)
	wf.AddStep(step5)

	order, _ := wf.GetTopologicalOrder()

	// Verify order constraints
	pos := make(map[string]int)
	for i, name := range order {
		pos[name] = i
	}

	checks := [][2]string{
		{"step1", "step2"},
		{"step1", "step3"},
		{"step2", "step4"},
		{"step3", "step4"},
		{"step2", "step5"},
		{"step4", "step5"},
	}

	for _, check := range checks {
		if pos[check[0]] >= pos[check[1]] {
			t.Errorf("%s should come before %s", check[0], check[1])
		}
	}

	executor, _ := NewWorkflowExecutor(wf)
	results, err := executor.Execute(context.Background(), nil)

	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	if len(results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(results))
	}

	t.Logf("✓ Complex DAG executed correctly")
	t.Logf("  Order: %v", order)
}

func BenchmarkWorkflowExecution(b *testing.B) {
	wf := NewWorkflowDefinition("bench")

	// Create 10-step linear workflow
	var prevStep string
	for i := 1; i <= 10; i++ {
		name := fmt.Sprintf("step%d", i)
		step := NewSimpleStep(name, func(ctx context.Context, input any) (any, error) {
			time.Sleep(1 * time.Millisecond) // Simulate work
			return input, nil
		})

		if i > 1 {
			step.WithDependencies(prevStep)
		}

		wf.AddStep(step)
		prevStep = name
	}

	executor, _ := NewWorkflowExecutor(wf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		executor.Execute(ctx, "input")
	}
}
