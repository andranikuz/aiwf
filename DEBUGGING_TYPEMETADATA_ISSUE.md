# TypeMetadata Not Being Passed to OpenAI API - Debugging Guide

## The Problem

On Fly.io, the digest application is getting this error:
```
Invalid schema for response_format 'aiwf_output':
In context=(), 'additionalProperties' is required to be supplied and to be false.
```

The OpenAI API is receiving a **minimal schema with empty properties**:
```json
{
  "properties": {},
  "required": [...],
  "type": "object"
}
```

This means **TypeMetadata is not being passed** to the OpenAI API call.

## Root Cause Analysis

The flow should be:

1. **Service Creation** (`NewService()` in `sdk/service.go`)
   - Creates agents
   - **Injects itself as TypeProvider** into each agent: `agent.Types = s`

2. **Agent Execution** (`CallModel()` in `runtime/go/aiwf/sdk_helpers.go`)
   - Checks if `a.Types != nil` and `a.Config.OutputTypeName != ""`
   - If yes, retrieves metadata: `a.Types.GetTypeMetadata(a.Config.OutputTypeName)`
   - Passes metadata to ModelCall

3. **OpenAI Provider** (`buildJSONSchemaFormat()` in `providers/openai/provider.go`)
   - If `call.TypeMetadata != nil`, converts it to JSON Schema
   - If `call.TypeMetadata == nil`, uses minimal schema (which fails OpenAI validation)

## The Issue

**TypeMetadata is nil**, which means one of these is true:
- ❌ `a.Types` is nil (Service not injected into agent)
- ❌ `a.Config.OutputTypeName` is empty (Agent config incomplete)
- ❌ `GetTypeMetadata()` is failing to retrieve data

## How to Debug

### 1. Check Generated Code

The generated `sdk/service.go` should contain TypeProvider injection:
```go
func NewService(client aiwf.ModelClient) *Service {
    s := &Service{client: client}

    cringe_detectorAgent := NewCringeDetectorAgent(client)
    cringe_detectorAgent.Types = s // ← Inject TypeProvider

    digest_assistantAgent := NewDigestAssistantAgent(client)
    digest_assistantAgent.Types = s // ← Inject TypeProvider

    s.agents = &Agents{...}
    return s
}
```

The generated `sdk/agents.go` should contain OutputTypeName in config:
```go
Config: aiwf.AgentConfig{
    Name:           "digest_assistant",
    OutputTypeName: "DigestOutput", // ← Must be set
    // ...
}
```

### 2. Check TypeMetadata Content

The generated `sdk/types.go` should have TypeMetadata map with all types:
```go
var TypeMetadata = map[string]interface{}{
    "DigestOutput": map[string]interface{}{
        "type": "object",
        "properties": {...},
        "additionalProperties": false, // ← Must be present everywhere
    },
    // ...
}
```

### 3. Deploy Debugging Version

The latest version includes debug logging:

**In `runtime/go/aiwf/sdk_helpers.go`:**
```
[DEBUG] CallModel: Agent=digest_assistant, Types=true, OutputTypeName=DigestOutput
[DEBUG] CallModel: Got TypeMetadata for DigestOutput
```

**In `providers/openai/provider.go`:**
```
openai: buildJSONSchemaFormat - TypeMetadata is not nil
```

If you see instead:
```
[DEBUG] CallModel: No TypeProvider or OutputTypeName empty
openai: buildJSONSchemaFormat - WARNING: TypeMetadata is nil, using minimal schema
```

Then you have found the problem!

## The Likely Issue

The user has stated:
> "так я скопировал их из своего проекта сюда, они там и используются"
> (I copied them from my project here, they are used there)

This suggests **the digest project has OLD SDK files** that don't have TypeProvider injection. The solution:

### Solution Steps

1. **Do NOT copy SDK files between projects**
   - Each project should generate its own SDK from the YAML template
   - Generated code should only be committed to the specific project that needs it

2. **In the digest project:**
   ```bash
   # Regenerate SDK with latest aiwf version
   aiwf generate /path/to/aiwf-digest.yaml --output ./pkg/sdk
   ```

3. **Verify the generated files have:**
   - ✅ TypeProvider injection in `service.go` (lines with `agent.Types = s`)
   - ✅ OutputTypeName in agent configs in `agents.go`
   - ✅ Complete TypeMetadata in `types.go`

4. **Deploy the digest project** with updated SDK

## Prevention

**Never** manually copy SDK files between projects. Instead:
- Each project maintains its own `.yaml` template
- Each project generates its own SDK from that template
- The AIWF framework ensures compatibility across projects

This prevents version mismatches and ensures TypeProvider injection is always present.
