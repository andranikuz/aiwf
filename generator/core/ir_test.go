package core

import "testing"

func TestBuildIRSuccess(t *testing.T) {
	spec := &Spec{
		Assistants: map[string]AssistantSpec{
			"writer": {
				Model:        "gpt-4",
				SystemPrompt: "Be creative",
				Resolved: AssistantResolution{
					InputSchemaPath:  "/tmp/input.json",
					OutputSchemaPath: "/tmp/output.json",
				},
			},
		},
		Workflows: map[string]WorkflowSpec{
			"novel": {
				Description: "Novel workflow",
				DAG: []WorkflowDAG{
					{
						Step:      "draft",
						Assistant: "writer",
					},
				},
			},
		},
	}

	ir, err := BuildIR(spec)
	if err != nil {
		t.Fatalf("BuildIR: %v", err)
	}

	if len(ir.Assistants) != 1 {
		t.Fatalf("expected 1 assistant, got %d", len(ir.Assistants))
	}
	writer := ir.Assistants["writer"]
	if writer.OutputSchemaPath != "/tmp/output.json" {
		t.Fatalf("unexpected output schema path: %s", writer.OutputSchemaPath)
	}

	wf, ok := ir.Workflows["novel"]
	if !ok {
		t.Fatalf("workflow novel not found")
	}
	if len(wf.Steps) != 1 || wf.Steps[0].Name != "draft" {
		t.Fatalf("unexpected steps: %+v", wf.Steps)
	}
}

func TestBuildIRDuplicateStep(t *testing.T) {
	spec := &Spec{
		Assistants: map[string]AssistantSpec{
			"writer": {
				Resolved: AssistantResolution{OutputSchemaPath: "/tmp/output.json"},
			},
		},
		Workflows: map[string]WorkflowSpec{
			"novel": {
				DAG: []WorkflowDAG{
					{Step: "draft", Assistant: "writer"},
					{Step: "draft", Assistant: "writer"},
				},
			},
		},
	}

	if _, err := BuildIR(spec); err == nil {
		t.Fatalf("expected duplicate step error")
	}
}

func TestBuildIRNeedsUnknownStep(t *testing.T) {
	spec := &Spec{
		Assistants: map[string]AssistantSpec{
			"writer": {
				Resolved: AssistantResolution{OutputSchemaPath: "/tmp/output.json"},
			},
		},
		Workflows: map[string]WorkflowSpec{
			"novel": {
				DAG: []WorkflowDAG{
					{Step: "draft", Assistant: "writer", Needs: []string{"outline"}},
				},
			},
		},
	}

	if _, err := BuildIR(spec); err == nil {
		t.Fatalf("expected missing dependency error")
	}
}

func TestBuildIRScatterValidation(t *testing.T) {
	spec := &Spec{
		Assistants: map[string]AssistantSpec{
			"writer": {
				Resolved: AssistantResolution{OutputSchemaPath: "/tmp/output.json"},
			},
		},
		Workflows: map[string]WorkflowSpec{
			"novel": {
				DAG: []WorkflowDAG{
					{Step: "draft", Assistant: "writer", Scatter: &ScatterSpec{}},
				},
			},
		},
	}

	if _, err := BuildIR(spec); err == nil {
		t.Fatalf("expected scatter validation error")
	}
}

func TestBuildIRReturnsMultiError(t *testing.T) {
	spec := &Spec{
		Assistants: map[string]AssistantSpec{
			"writer": {Resolved: AssistantResolution{OutputSchemaPath: "/tmp/output.json"}},
		},
		Workflows: map[string]WorkflowSpec{
			"novel": {
				DAG: []WorkflowDAG{
					{Step: "draft", Assistant: "writer"},
					{Step: "draft", Assistant: "writer"},
					{Step: "review", Assistant: "missing"},
				},
			},
		},
	}

	_, err := BuildIR(spec)
	if err == nil {
		t.Fatalf("expected multi error")
	}
	me, ok := err.(*MultiError)
	if !ok {
		t.Fatalf("expected MultiError, got %T", err)
	}
	if len(me.Errors) < 2 {
		t.Fatalf("expected multiple errors, got %d", len(me.Errors))
	}
}
