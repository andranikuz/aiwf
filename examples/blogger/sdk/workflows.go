package blog

import (
    "context"
    "fmt"

    "github.com/andranikuz/aiwf/runtime/go/aiwf"
)

type Workflows interface {
    BlogPost() BlogPostWorkflow
}

type workflows struct {
    client aiwf.ModelClient
}
func (w *workflows) BlogPost() BlogPostWorkflow { return &blogPostWorkflow{client: w.client} }

type BlogPostWorkflow interface {
    Run(ctx context.Context, input BlogPostInput) (BlogPostOutput, *aiwf.Trace, error)
}

type blogPostWorkflow struct {
    client aiwf.ModelClient
}

func (r *blogPostWorkflow) Run(ctx context.Context, input BlogPostInput) (BlogPostOutput, *aiwf.Trace, error) {
    agents := &agents{client: r.client}
    result := BlogPostOutput{}
    traces := make([]*aiwf.Trace, 0, 2)
    prev := map[string]any{"input": input}
    {
        stepInput := OutlineInput{}
        if candidate, ok := prev["input"].(OutlineInput); ok {
            stepInput = candidate
        }

        output, trace, err := agents.Outline().Run(ctx, stepInput)
        if err != nil {
            return result, mergeTraces(traces...), fmt.Errorf("step outline failed: %w", err)
        }
        prev["outline"] = output
        traces = append(traces, trace)
    }
    {
        stepInput := DraftInput{}
        if candidate, ok := prev["outline"].(DraftInput); ok {
            stepInput = candidate
        }

        output, trace, err := agents.Draft().Run(ctx, stepInput)
        if err != nil {
            return result, mergeTraces(traces...), fmt.Errorf("step draft failed: %w", err)
        }
        prev["draft"] = output
        traces = append(traces, trace)
        result = mergeWorkflowOutput(result, BlogPostOutput(output))
    }

    return result, mergeTraces(traces...), nil
}
