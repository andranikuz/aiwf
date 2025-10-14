package blog

import (
    "context"
    "encoding/json"

    "github.com/andranikuz/aiwf/runtime/go/aiwf"
)
var draftOutputSchemaJSON = json.RawMessage("{\n  \"$schema\": \"http://json-schema.org/draft-07/schema#\",\n  \"type\": \"object\",\n  \"properties\": {\n    \"title\": { \"type\": \"string\" },\n    \"content\": { \"type\": \"string\" }\n  },\n  \"required\": [\"title\", \"content\"],\n  \"additionalProperties\": false\n}\n")
var outlineOutputSchemaJSON = json.RawMessage("{\n  \"$schema\": \"http://json-schema.org/draft-07/schema#\",\n  \"type\": \"object\",\n  \"properties\": {\n    \"title\": { \"type\": \"string\" },\n    \"sections\": {\n      \"type\": \"array\",\n      \"items\": { \"type\": \"string\" }\n    }\n  },\n  \"required\": [\"title\", \"sections\"],\n  \"additionalProperties\": false\n}\n")

type Agents interface {
    Draft() DraftAgent
    Outline() OutlineAgent
}

type agents struct {
    client aiwf.ModelClient
}
func (a *agents) Draft() DraftAgent { return &draftAgent{client: a.client} }

type DraftAgent interface {
    Run(ctx context.Context, input DraftInput) (DraftOutput, *aiwf.Trace, error)
}

type draftAgent struct {
    client aiwf.ModelClient
}

func (a *draftAgent) Run(ctx context.Context, input DraftInput) (DraftOutput, *aiwf.Trace, error) {
    call := aiwf.ModelCall{
        Model:           "gpt-4o",
        SystemPrompt:    "Write a concise blog post section by section.\n",
        InputSchemaRef:  "draft_input.json",
        OutputSchemaRef: "draft_output.json",
        OutputSchema:    draftOutputSchemaJSON,
        Payload:         input,
    }

    raw, usage, err := a.client.CallJSONSchema(ctx, call)
    if err != nil {
        return DraftOutput{}, nil, err
    }

    var output DraftOutput
    if err := json.Unmarshal(raw, &output); err != nil {
        return DraftOutput{}, nil, err
    }

    trace := &aiwf.Trace{StepName: "draft", Usage: usage}
    return output, trace, nil
}
func (a *agents) Outline() OutlineAgent { return &outlineAgent{client: a.client} }

type OutlineAgent interface {
    Run(ctx context.Context, input OutlineInput) (OutlineOutput, *aiwf.Trace, error)
}

type outlineAgent struct {
    client aiwf.ModelClient
}

func (a *outlineAgent) Run(ctx context.Context, input OutlineInput) (OutlineOutput, *aiwf.Trace, error) {
    call := aiwf.ModelCall{
        Model:           "gpt-4o-mini",
        SystemPrompt:    "You are an editorial planner. Produce a catchy title and a list of sections for a blog post.\n",
        InputSchemaRef:  "outline_input.json",
        OutputSchemaRef: "outline_output.json",
        OutputSchema:    outlineOutputSchemaJSON,
        Payload:         input,
    }

    raw, usage, err := a.client.CallJSONSchema(ctx, call)
    if err != nil {
        return OutlineOutput{}, nil, err
    }

    var output OutlineOutput
    if err := json.Unmarshal(raw, &output); err != nil {
        return OutlineOutput{}, nil, err
    }

    trace := &aiwf.Trace{StepName: "outline", Usage: usage}
    return output, trace, nil
}
