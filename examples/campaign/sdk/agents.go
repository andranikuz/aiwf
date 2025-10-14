package blog

import (
    "context"
    "encoding/json"

    "github.com/andranikuz/aiwf/runtime/go/aiwf"
)
var contentStrategyOutputSchemaJSON = json.RawMessage("{\n  \"$schema\": \"http://json-schema.org/draft-07/schema#\",\n  \"type\": \"object\",\n  \"properties\": {\n    \"channels\": {\n      \"type\": \"array\",\n      \"items\": {\n        \"type\": \"object\",\n        \"properties\": {\n          \"name\": { \"type\": \"string\" },\n          \"format\": { \"type\": \"string\" },\n          \"cadence\": { \"type\": \"string\" },\n          \"audience_segment\": { \"type\": \"string\" },\n          \"kpi\": { \"type\": \"string\" }\n        },\n        \"required\": [\"name\", \"format\", \"audience_segment\", \"cadence\", \"kpi\"],\n        \"additionalProperties\": false\n      }\n    },\n    \"budget_notes\": { \"type\": \"string\" }\n  },\n  \"required\": [\"channels\", \"budget_notes\"],\n  \"additionalProperties\": false\n}\n")
var launchTimelineOutputSchemaJSON = json.RawMessage("{\n  \"$schema\": \"http://json-schema.org/draft-07/schema#\",\n  \"type\": \"object\",\n  \"properties\": {\n    \"milestones\": {\n      \"type\": \"array\",\n      \"items\": {\n        \"type\": \"object\",\n        \"properties\": {\n          \"week\": { \"type\": \"integer\", \"minimum\": 1 },\n          \"objective\": { \"type\": \"string\" },\n          \"owner\": { \"type\": \"string\" },\n          \"deliverables\": {\n            \"type\": \"array\",\n            \"items\": { \"type\": \"string\" }\n          }\n        },\n        \"required\": [\"week\", \"objective\", \"owner\", \"deliverables\"],\n        \"additionalProperties\": false\n      }\n    }\n  },\n  \"required\": [\"milestones\"],\n  \"additionalProperties\": false\n}\n")
var marketResearchOutputSchemaJSON = json.RawMessage("{\n  \"$schema\": \"http://json-schema.org/draft-07/schema#\",\n  \"type\": \"object\",\n  \"properties\": {\n    \"summary\": { \"type\": \"string\" },\n    \"segments\": {\n      \"type\": \"array\",\n      \"items\": {\n        \"type\": \"object\",\n        \"properties\": {\n          \"name\": { \"type\": \"string\" },\n          \"persona\": { \"type\": \"string\" },\n          \"pain_points\": {\n            \"type\": \"array\",\n            \"items\": { \"type\": \"string\" }\n          },\n          \"preferred_channels\": {\n            \"type\": \"array\",\n            \"items\": { \"type\": \"string\" }\n          }\n        },\n        \"required\": [\"name\", \"persona\", \"pain_points\", \"preferred_channels\"],\n        \"additionalProperties\": false\n      }\n    }\n  },\n  \"required\": [\"summary\", \"segments\"],\n  \"additionalProperties\": false\n}\n")
var riskAssessmentOutputSchemaJSON = json.RawMessage("{\n  \"$schema\": \"http://json-schema.org/draft-07/schema#\",\n  \"type\": \"object\",\n  \"properties\": {\n    \"risks\": {\n      \"type\": \"array\",\n      \"items\": {\n        \"type\": \"object\",\n        \"properties\": {\n          \"label\": { \"type\": \"string\" },\n          \"impact\": { \"type\": \"string\" },\n          \"likelihood\": { \"type\": \"string\" },\n          \"mitigation\": { \"type\": \"string\" },\n          \"monitoring_metric\": { \"type\": \"string\" }\n        },\n        \"required\": [\"label\", \"impact\", \"mitigation\", \"likelihood\", \"monitoring_metric\"],\n        \"additionalProperties\": false\n      }\n    }\n  },\n  \"required\": [\"risks\"],\n  \"additionalProperties\": false\n}\n")
var valuePropositionOutputSchemaJSON = json.RawMessage("{\n  \"$schema\": \"http://json-schema.org/draft-07/schema#\",\n  \"type\": \"object\",\n  \"properties\": {\n    \"narrative\": { \"type\": \"string\" },\n    \"value_map\": {\n      \"type\": \"array\",\n      \"items\": {\n        \"type\": \"object\",\n        \"properties\": {\n          \"segment\": { \"type\": \"string\" },\n          \"core_message\": { \"type\": \"string\" },\n          \"proof_points\": {\n            \"type\": \"array\",\n            \"items\": { \"type\": \"string\" }\n          }\n        },\n        \"required\": [\"segment\", \"core_message\", \"proof_points\"],\n        \"additionalProperties\": false\n      }\n    }\n  },\n  \"required\": [\"narrative\", \"value_map\"],\n  \"additionalProperties\": false\n}\n")

type Agents interface {
    ContentStrategy() ContentStrategyAgent
    LaunchTimeline() LaunchTimelineAgent
    MarketResearch() MarketResearchAgent
    RiskAssessment() RiskAssessmentAgent
    ValueProposition() ValuePropositionAgent
}

type agents struct {
    client aiwf.ModelClient
}
func (a *agents) ContentStrategy() ContentStrategyAgent { return &contentStrategyAgent{client: a.client} }

type ContentStrategyAgent interface {
    Run(ctx context.Context, input ContentStrategyInput) (ContentStrategyOutput, *aiwf.Trace, error)
}

type contentStrategyAgent struct {
    client aiwf.ModelClient
}

func (a *contentStrategyAgent) Run(ctx context.Context, input ContentStrategyInput) (ContentStrategyOutput, *aiwf.Trace, error) {
    call := aiwf.ModelCall{
        Model:           "gpt-4o",
        SystemPrompt:    "You design integrated launch strategies. Transform value propositions into concrete channel plays with cadence and KPIs.\n",
        InputSchemaRef:  "strategy_input.json",
        OutputSchemaRef: "strategy_output.json",
        OutputSchema:    contentStrategyOutputSchemaJSON,
        Payload:         input,
    }

    raw, usage, err := a.client.CallJSONSchema(ctx, call)
    if err != nil {
        return ContentStrategyOutput{}, nil, err
    }

    var output ContentStrategyOutput
    if err := json.Unmarshal(raw, &output); err != nil {
        return ContentStrategyOutput{}, nil, err
    }

    trace := &aiwf.Trace{StepName: "content_strategy", Usage: usage}
    return output, trace, nil
}
func (a *agents) LaunchTimeline() LaunchTimelineAgent { return &launchTimelineAgent{client: a.client} }

type LaunchTimelineAgent interface {
    Run(ctx context.Context, input LaunchTimelineInput) (LaunchTimelineOutput, *aiwf.Trace, error)
}

type launchTimelineAgent struct {
    client aiwf.ModelClient
}

func (a *launchTimelineAgent) Run(ctx context.Context, input LaunchTimelineInput) (LaunchTimelineOutput, *aiwf.Trace, error) {
    call := aiwf.ModelCall{
        Model:           "gpt-4o",
        SystemPrompt:    "You are an operations lead. Map the launch plan into weekly milestones with owners and deliverables.\n",
        InputSchemaRef:  "timeline_input.json",
        OutputSchemaRef: "timeline_output.json",
        OutputSchema:    launchTimelineOutputSchemaJSON,
        Payload:         input,
    }

    raw, usage, err := a.client.CallJSONSchema(ctx, call)
    if err != nil {
        return LaunchTimelineOutput{}, nil, err
    }

    var output LaunchTimelineOutput
    if err := json.Unmarshal(raw, &output); err != nil {
        return LaunchTimelineOutput{}, nil, err
    }

    trace := &aiwf.Trace{StepName: "launch_timeline", Usage: usage}
    return output, trace, nil
}
func (a *agents) MarketResearch() MarketResearchAgent { return &marketResearchAgent{client: a.client} }

type MarketResearchAgent interface {
    Run(ctx context.Context, input MarketResearchInput) (MarketResearchOutput, *aiwf.Trace, error)
}

type marketResearchAgent struct {
    client aiwf.ModelClient
}

func (a *marketResearchAgent) Run(ctx context.Context, input MarketResearchInput) (MarketResearchOutput, *aiwf.Trace, error) {
    call := aiwf.ModelCall{
        Model:           "gpt-4o-mini",
        SystemPrompt:    "You are a research strategist. Synthesize audience insights, surface 2-3 segments, and highlight pains plus preferred channels.\n",
        InputSchemaRef:  "research_input.json",
        OutputSchemaRef: "research_output.json",
        OutputSchema:    marketResearchOutputSchemaJSON,
        Payload:         input,
    }

    raw, usage, err := a.client.CallJSONSchema(ctx, call)
    if err != nil {
        return MarketResearchOutput{}, nil, err
    }

    var output MarketResearchOutput
    if err := json.Unmarshal(raw, &output); err != nil {
        return MarketResearchOutput{}, nil, err
    }

    trace := &aiwf.Trace{StepName: "market_research", Usage: usage}
    return output, trace, nil
}
func (a *agents) RiskAssessment() RiskAssessmentAgent { return &riskAssessmentAgent{client: a.client} }

type RiskAssessmentAgent interface {
    Run(ctx context.Context, input RiskAssessmentInput) (RiskAssessmentOutput, *aiwf.Trace, error)
}

type riskAssessmentAgent struct {
    client aiwf.ModelClient
}

func (a *riskAssessmentAgent) Run(ctx context.Context, input RiskAssessmentInput) (RiskAssessmentOutput, *aiwf.Trace, error) {
    call := aiwf.ModelCall{
        Model:           "gpt-4o-mini",
        SystemPrompt:    "You analyze plans for operational and market risks. Provide mitigations and metrics to monitor.\n",
        InputSchemaRef:  "risk_input.json",
        OutputSchemaRef: "risk_output.json",
        OutputSchema:    riskAssessmentOutputSchemaJSON,
        Payload:         input,
    }

    raw, usage, err := a.client.CallJSONSchema(ctx, call)
    if err != nil {
        return RiskAssessmentOutput{}, nil, err
    }

    var output RiskAssessmentOutput
    if err := json.Unmarshal(raw, &output); err != nil {
        return RiskAssessmentOutput{}, nil, err
    }

    trace := &aiwf.Trace{StepName: "risk_assessment", Usage: usage}
    return output, trace, nil
}
func (a *agents) ValueProposition() ValuePropositionAgent { return &valuePropositionAgent{client: a.client} }

type ValuePropositionAgent interface {
    Run(ctx context.Context, input ValuePropositionInput) (ValuePropositionOutput, *aiwf.Trace, error)
}

type valuePropositionAgent struct {
    client aiwf.ModelClient
}

func (a *valuePropositionAgent) Run(ctx context.Context, input ValuePropositionInput) (ValuePropositionOutput, *aiwf.Trace, error) {
    call := aiwf.ModelCall{
        Model:           "gpt-4o-mini",
        SystemPrompt:    "You convert research findings into differentiated value propositions. Draft crisp positioning for each segment.\n",
        InputSchemaRef:  "proposition_input.json",
        OutputSchemaRef: "proposition_output.json",
        OutputSchema:    valuePropositionOutputSchemaJSON,
        Payload:         input,
    }

    raw, usage, err := a.client.CallJSONSchema(ctx, call)
    if err != nil {
        return ValuePropositionOutput{}, nil, err
    }

    var output ValuePropositionOutput
    if err := json.Unmarshal(raw, &output); err != nil {
        return ValuePropositionOutput{}, nil, err
    }

    trace := &aiwf.Trace{StepName: "value_proposition", Usage: usage}
    return output, trace, nil
}
