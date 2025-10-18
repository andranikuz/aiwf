# Providers

 50;870F88 LLM-?@>20945@>2 4;O AIWF.

## >ABC?=K5 ?@>20945@K

### OpenAI (`providers/openai`)

>445@68205B OpenAI Responses API A JSON Schema.

```go
import "github.com/andranikuz/aiwf/providers/openai"

client, err := openai.NewClient(openai.ClientConfig{
    APIKey:  os.Getenv("OPENAI_API_KEY"),
    BaseURL: "https://api.openai.com/v1",
})
```

**>7<>6=>AB8:**
-  Structured outputs G5@57 JSON Schema
-  >=25@B0F8O TypeDef ’ JSON Schema
-  Thread management
- ó Streaming (2 @07@01>B:5)

### Anthropic (`providers/anthropic`)

 @07@01>B:5.

### Local Stub (`providers/local`)

03;CH:0 4;O B5AB8@>20=8O 157 2K7>20 @50;L=KE API.

```go
import "github.com/andranikuz/aiwf/providers/local"

client := local.NewClient()
```

##  50;870F8O ?@>20945@0

@>20945@ 4>;65= @50;87>20BL 8=B5@D59A `ModelClient`:

```go
type ModelClient interface {
    CallJSONSchema(ctx context.Context, call ModelCall) ([]byte, Tokens, error)
    CallJSONSchemaStream(ctx context.Context, call ModelCall) (<-chan StreamChunk, Tokens, error)
}
```

### ModelCall

```go
type ModelCall struct {
    Model          string
    SystemPrompt   string
    UserPrompt     string
    MaxTokens      int
    Temperature    float64
    Payload        any    // E>4=K5 40==K5
    InputTypeName  string // <O 2E>4=>3> B8?0
    OutputTypeName string // <O 2KE>4=>3> B8?0
    TypeMetadata   any    // TypeDef 8;8 JSON Schema
}
```

### >=25@B0F8O B8?>2

@>20945@K 4>;6=K :>=25@B8@>20BL `TypeMetadata` 2 A2>9 D>@<0B AE5<K:

```go
// OpenAI ?@8<5@
converter := NewSchemaConverter()
jsonSchema, err := converter.ConvertTypeMetadata(call.TypeMetadata)
```

## "5AB8@>20=85

```bash
go test ./providers/...
```

## Roadmap

- [ ] Anthropic Claude provider
- [ ] Google Gemini provider
- [ ] Cohere provider
- [ ] Streaming 4;O 2A5E ?@>20945@>2
- [ ] Retry ;>38:0
- [ ] Rate limiting