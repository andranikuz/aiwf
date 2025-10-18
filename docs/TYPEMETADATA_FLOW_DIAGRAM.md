# TypeMetadata Flow Diagram

## Complete Flow: From YAML Template to OpenAI API

```
┌─────────────────────────────────────────────────────────────────┐
│ 1. USER CODE - Digest Project                                   │
│                                                                   │
│  service := sdk.NewService(openaiClient)                        │
│  digestAgent := service.Agents().DigestAssistant               │
│  result, _, err := digestAgent.Run(ctx, input)                 │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│ 2. SDK CODE - service.go (GENERATED)                            │
│                                                                   │
│  func NewService(client aiwf.ModelClient) *Service {            │
│      s := &Service{client: client}                              │
│      agent := NewDigestAssistantAgent(client)                   │
│      agent.Types = s  ← TYPE PROVIDER INJECTED                │
│      return s                                                    │
│  }                                                               │
│                                                                   │
│  func (s *Service) GetTypeMetadata(typeName string) (any, err) {│
│      meta, ok := TypeMetadata[typeName]                        │
│      return meta, nil  ← RETURNS FULL JSON SCHEMA              │
│  }                                                               │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│ 3. SDK CODE - agents.go (GENERATED)                             │
│                                                                   │
│  type DigestAssistantAgent struct {                            │
│      aiwf.AgentBase                                             │
│  }                                                               │
│                                                                   │
│  func NewDigestAssistantAgent(...) *DigestAssistantAgent {     │
│      return &DigestAssistantAgent{                             │
│          AgentBase: aiwf.AgentBase{                             │
│              Config: aiwf.AgentConfig{                          │
│                  Name: "digest_assistant",                      │
│                  OutputTypeName: "DigestOutput"  ← SET          │
│              },                                                  │
│              Client: client,                                    │
│          },                                                      │
│      }                                                           │
│  }                                                               │
│                                                                   │
│  func (a *DigestAssistantAgent) Run(...) (..., error) {        │
│      result, trace, err := a.CallModel(ctx, input, nil)        │
│  }                                                               │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│ 4. RUNTIME - sdk_helpers.go                                     │
│                                                                   │
│  func (a *AgentBase) CallModel(...) (..., error) {             │
│      var typeMetadata any                                       │
│                                                                   │
│      if a.Types != nil && a.Config.OutputTypeName != "" {     │
│          meta, _ := a.Types.GetTypeMetadata(                   │
│              a.Config.OutputTypeName)  ← RETRIEVES "DigestOutput"
│          typeMetadata = meta                                    │
│      }                                                           │
│                                                                   │
│      call := ModelCall{                                         │
│          OutputTypeName: "DigestOutput",                        │
│          TypeMetadata: typeMetadata,  ← FULL JSON SCHEMA       │
│      }                                                           │
│                                                                   │
│      result := a.Client.CallJSONSchema(ctx, call)             │
│  }                                                               │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│ 5. PROVIDER - providers/openai/provider.go                      │
│                                                                   │
│  func (c *Client) buildJSONSchemaFormat(call) (textSection) {  │
│      var schema json.RawMessage                                 │
│                                                                   │
│      if call.TypeMetadata != nil {                             │
│          schema = c.converter.ConvertTypeMetadata(             │
│              call.TypeMetadata)  ← USE FULL SCHEMA            │
│      } else {                                                    │
│          schema = minimalSchema  ← FALLBACK (CAUSES ERROR!)    │
│      }                                                           │
│                                                                   │
│      return textSection{                                        │
│          Format: textFormat{                                    │
│              Type: "json_schema",                               │
│              Schema: schema,                                    │
│          },                                                      │
│      }                                                           │
│  }                                                               │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│ 6. SDK CODE - types.go (GENERATED)                              │
│                                                                   │
│  var TypeMetadata = map[string]interface{}{                    │
│      "DigestOutput": {                                          │
│          "type": "object",                                      │
│          "properties": {                                        │
│              "events": {...},                                   │
│              "summary": {...},                                  │
│              "funny_moments": {...},                            │
│          },                                                      │
│          "required": [...],                                     │
│          "additionalProperties": false,  ← REQUIRED BY OPENAI  │
│      },                                                          │
│  }                                                               │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│ 7. OPENAI API                                                   │
│                                                                   │
│  POST /v1/responses                                             │
│  {                                                               │
│      "model": "gpt-4o-mini",                                   │
│      "text": {                                                   │
│          "format": {                                            │
│              "type": "json_schema",                             │
│              "schema": {                                        │
│                  "type": "object",                              │
│                  "properties": {...},                           │
│                  "required": [...],                             │
│                  "additionalProperties": false  ← PRESENT!     │
│              }                                                   │
│          }                                                       │
│      }                                                           │
│  }                                                               │
│                                                                   │
│  RESPONSE: ✅ 200 OK (schema is valid)                         │
└─────────────────────────────────────────────────────────────────┘
```

## When TypeMetadata is Nil - The Error Path

```
┌─────────────────────────────────────────────────────────────────┐
│ PROBLEM: a.Types is nil in CallModel                            │
│                                                                   │
│  if a.Types != nil && a.Config.OutputTypeName != "" {         │
│      meta, _ := a.Types.GetTypeMetadata(...)                   │
│  }                                                               │
│  ↑                                                               │
│  FALSE - a.Types was never injected!                           │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│ RESULT: typeMetadata stays nil                                  │
│                                                                   │
│  call := ModelCall{                                             │
│      TypeMetadata: nil,  ← EMPTY                               │
│  }                                                               │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│ FALLBACK: Minimal schema is used                                │
│                                                                   │
│  if call.TypeMetadata != nil {                                 │
│      // Use full schema                                         │
│  } else {                                                        │
│      minimalSchema := {                                         │
│          "type": "object",                                      │
│          "properties": {},  ← EMPTY!                           │
│          "additionalProperties": false,                         │
│      }                                                           │
│  }                                                               │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│ ERROR: OpenAI rejects request                                   │
│                                                                   │
│  "error": {                                                      │
│      "message": "Invalid schema for response_format 'aiwf_output':│
│                  In context=(), 'additionalProperties' is        │
│                  required to be supplied and to be false.",      │
│      "code": "invalid_json_schema"                              │
│  }                                                               │
│                                                                   │
│  Because the schema has empty properties but minimal schema    │
│  structure doesn't match the actual expected output structure.  │
└─────────────────────────────────────────────────────────────────┘
```

## Why "a.Types is nil"

This happens when the digest project is using OLD SDK files where:

```go
// OLD service.go (WITHOUT TypeProvider injection)
func NewService(client aiwf.ModelClient) *Service {
    s := &Service{client: client}
    s.agents = &Agents{
        DigestAssistant: NewDigestAssistantAgent(client),
        // ❌ agent.Types = s  ← MISSING!
    }
    return s
}
```

vs

```go
// NEW service.go (WITH TypeProvider injection)
func NewService(client aiwf.ModelClient) *Service {
    s := &Service{client: client}

    agent := NewDigestAssistantAgent(client)
    agent.Types = s  // ✅ INJECTED

    s.agents = &Agents{
        DigestAssistant: agent,
    }
    return s
}
```

## The Fix

Regenerate the SDK in the digest project using the latest aiwf version:
```bash
cd digest_project
aiwf generate ./aiwf-digest.yaml --output ./pkg/sdk
# This will generate service.go WITH TypeProvider injection
```

Then deploy the updated digest project.
