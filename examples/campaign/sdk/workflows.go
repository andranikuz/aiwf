package blog

import (
    "context"
    "fmt"

    "github.com/andranikuz/aiwf/runtime/go/aiwf"
)

type Workflows interface {
    CampaignLaunch() CampaignLaunchWorkflow
}

type workflows struct {
    client aiwf.ModelClient
}
func (w *workflows) CampaignLaunch() CampaignLaunchWorkflow { return &campaignLaunchWorkflow{client: w.client} }

type CampaignLaunchWorkflow interface {
    Run(ctx context.Context, input CampaignLaunchInput) (CampaignLaunchOutput, *aiwf.Trace, error)
}

type campaignLaunchWorkflow struct {
    client aiwf.ModelClient
}

func (r *campaignLaunchWorkflow) Run(ctx context.Context, input CampaignLaunchInput) (CampaignLaunchOutput, *aiwf.Trace, error) {
    agents := &agents{client: r.client}
    result := CampaignLaunchOutput{}
    traces := make([]*aiwf.Trace, 0, 5)
    prev := map[string]any{"input": input}
    {
        stepInput := MarketResearchInput{}
        if candidate, ok := prev["input"].(MarketResearchInput); ok {
            stepInput = candidate
        }

        output, trace, err := agents.MarketResearch().Run(ctx, stepInput)
        if err != nil {
            return result, mergeTraces(traces...), fmt.Errorf("step research failed: %w", err)
        }
        prev["research"] = output
        traces = append(traces, trace)
    }
    {
        stepInput := ValuePropositionInput{}
        if candidate, ok := prev["research"].(ValuePropositionInput); ok {
            stepInput = candidate
        }

        output, trace, err := agents.ValueProposition().Run(ctx, stepInput)
        if err != nil {
            return result, mergeTraces(traces...), fmt.Errorf("step messaging failed: %w", err)
        }
        prev["messaging"] = output
        traces = append(traces, trace)
    }
    {
        stepInput := ContentStrategyInput{}
        if candidate, ok := prev["messaging"].(ContentStrategyInput); ok {
            stepInput = candidate
        }

        output, trace, err := agents.ContentStrategy().Run(ctx, stepInput)
        if err != nil {
            return result, mergeTraces(traces...), fmt.Errorf("step strategy failed: %w", err)
        }
        prev["strategy"] = output
        traces = append(traces, trace)
    }
    {
        stepInput := LaunchTimelineInput{}
        if candidate, ok := prev["strategy"].(LaunchTimelineInput); ok {
            stepInput = candidate
        }

        output, trace, err := agents.LaunchTimeline().Run(ctx, stepInput)
        if err != nil {
            return result, mergeTraces(traces...), fmt.Errorf("step timeline failed: %w", err)
        }
        prev["timeline"] = output
        traces = append(traces, trace)
    }
    {
        stepInput := RiskAssessmentInput{}
        if candidate, ok := prev["timeline"].(RiskAssessmentInput); ok {
            stepInput = candidate
        }

        output, trace, err := agents.RiskAssessment().Run(ctx, stepInput)
        if err != nil {
            return result, mergeTraces(traces...), fmt.Errorf("step risks failed: %w", err)
        }
        prev["risks"] = output
        traces = append(traces, trace)
        result = mergeWorkflowOutput(result, CampaignLaunchOutput(output))
    }

    return result, mergeTraces(traces...), nil
}
