# HTTP Client Generation

AIWF can generate lightweight HTTP clients for multiple languages that connect to `aiwf serve` server.

## Why HTTP Clients?

Instead of porting the entire runtime + providers to each language, we:
1. Run one Go server with `aiwf serve`
2. Generate simple HTTP clients for any language
3. Keep API keys on the server (more secure)
4. Centralize monitoring and caching

**Benefits:**
- ✅ ~100 lines of code vs ~1000+ for full SDK
- ✅ Only HTTP library dependency (no LLM SDKs)
- ✅ API keys stay on server
- ✅ Easy to add new languages
- ✅ Centralized monitoring

## Quick Start

### 1. Start Server

```bash
export OPENAI_API_KEY="sk-..."
aiwf serve -f config.yaml
```

Server runs on `http://127.0.0.1:8080` by default.

### 2. Generate Client

```bash
# PHP
aiwf sdk -f config.yaml -l php -o ./client.php

# Python (coming soon)
aiwf sdk -f config.yaml -l python -o ./client.py

# TypeScript (coming soon)
aiwf sdk -f config.yaml -l typescript -o ./client.ts

# Go HTTP client
aiwf sdk -f config.yaml -l go --type client -o ./client.go
```

### 3. Use Client

See language-specific examples in `examples/` directory.

## Supported Languages

| Language | Status | Dependencies | Code Size |
|----------|--------|--------------|-----------|
| **PHP** | ✅ Ready | cURL | ~100 lines |
| **Python** | 🚧 Coming soon | requests | ~80 lines |
| **TypeScript** | 🚧 Coming soon | fetch | ~90 lines |
| **Go** | ✅ Ready | net/http | ~120 lines |
| **Ruby** | 📋 Planned | net/http | ~100 lines |
| **Java** | 📋 Planned | HttpClient | ~150 lines |
| **C#** | 📋 Planned | HttpClient | ~130 lines |

## CLI Options

```bash
aiwf sdk [flags]

Flags:
  -f, --file string       Path to YAML config (required)
  -l, --lang string       Target language (default "go")
  -o, --out string        Output directory (required)
      --type string       SDK type: "full" or "client" (auto-detected)
      --base-url string   Base URL for HTTP client (default "http://127.0.0.1:8080")
      --package string    Package/module name (default "aiwfgen")
```

### Auto-Detection

The `--type` flag is auto-detected based on language:

- **Go**: `full` by default (has runtime support)
- **PHP, Python, TS, etc**: `client` by default (HTTP only)

Override with `--type`:
```bash
# Force HTTP client for Go
aiwf sdk -f config.yaml -l go --type client -o ./client.go
```

## Generated Code Structure

### PHP Example

Given this YAML:
```yaml
version: 0.3

types:
  Request:
    text: string
    lang: enum(en, es, fr)

  Response:
    result: string
    confidence: number(0..1)

assistants:
  translator:
    use: openai
    model: gpt-4o-mini
    input_type: $Request
    output_type: $Response
```

Generates:
```php
<?php
namespace AIWFClient;

// Type classes
class Request {
    public function __construct(
        public string $lang,
        public string $text
    ) {}

    public function toArray(): array { ... }
    public static function fromArray(array $data): Request { ... }
}

class Response {
    public function __construct(
        public float $confidence,
        public string $result
    ) {}

    public function toArray(): array { ... }
    public static function fromArray(array $data): Response { ... }
}

// Client
class AIWFClient {
    public function __construct(
        string $baseURL = 'http://127.0.0.1:8080',
        ?string $apiKey = null
    ) { ... }

    public function translator(Request $request): Response {
        // HTTP POST to /agent/translator
    }
}
```

## Features

### Type Safety

All generated clients include:
- ✅ Type hints/annotations
- ✅ Input/output validation
- ✅ Enum support
- ✅ Nested object support
- ✅ Array types

### Authentication

```php
$client = new AIWFClient(
    baseURL: 'https://api.example.com',
    apiKey: 'your-secret-key'
);
```

Server checks `X-API-Key` header if `API_KEY` env var is set.

### Error Handling

All clients throw exceptions on HTTP errors:

```php
try {
    $response = $client->translator($request);
} catch (\Exception $e) {
    echo "Error: " . $e->getMessage();
}
```

### Serialization

Auto-generated serialization methods:
- `toArray()` - Convert object to array for JSON
- `fromArray($data)` - Create object from JSON response

## Comparison: Full SDK vs HTTP Client

| Aspect | Full SDK | HTTP Client |
|--------|----------|-------------|
| **Setup** | Complex | Simple |
| **Code Size** | ~1000+ lines | ~100 lines |
| **Dependencies** | LLM SDK + Runtime | HTTP lib only |
| **API Keys** | In application ⚠️ | On server ✅ |
| **Latency** | Direct (~500ms) | +50-200ms network |
| **Maintenance** | High (per lang) | Low (one server) |
| **Deployment** | Standalone | Needs server |
| **Monitoring** | Distributed | Centralized ✅ |
| **Caching** | Per instance | Shared ✅ |
| **Updates** | Redeploy app | Just restart server |

## Use Cases

### HTTP Client (Recommended)

✅ **Web applications** - PHP/Python/Node.js backends
✅ **Mobile backends** - Any language
✅ **Microservices** - Service-to-service calls
✅ **Internal tools** - Admin dashboards
✅ **Prototyping** - Fast iteration

### Full SDK

✅ **CLI tools** - Standalone binaries
✅ **Edge computing** - No server access
✅ **Serverless** - AWS Lambda, etc.
✅ **Embedded systems** - Minimal network
✅ **Desktop apps** - Offline support

## Production Deployment

### Docker

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o aiwf ./cmd/aiwf

FROM alpine:latest
COPY --from=builder /app/aiwf .
COPY config.yaml .
EXPOSE 8080
ENV OPENAI_API_KEY=""
CMD ["./aiwf", "serve", "-f", "config.yaml", "--host", "0.0.0.0"]
```

### Kubernetes

See [SERVE_GUIDE.md](./SERVE_GUIDE.md) for complete Kubernetes manifests.

### Environment Variables

```bash
# LLM Provider (one required)
OPENAI_API_KEY=sk-...
GROK_API_KEY=xai-...
ANTHROPIC_API_KEY=sk-ant-...

# Server config
PORT=8080
HOST=0.0.0.0
API_KEY=your-secret-key  # Optional authentication

# Custom base URL for clients
AIWF_BASE_URL=https://api.yourcompany.com
```

## Custom Base URL

Generate clients with custom base URL:

```bash
aiwf sdk -f config.yaml -l php \
  --base-url https://api.yourcompany.com \
  -o ./client.php
```

## Troubleshooting

### Connection Refused

Server not running:
```bash
aiwf serve -f config.yaml
```

### Type Errors

Regenerate client after YAML changes:
```bash
aiwf sdk -f config.yaml -l php -o ./client.php
```

### Authentication Failed

Set API key on server:
```bash
export API_KEY="secret"
aiwf serve -f config.yaml
```

And in client:
```php
$client = new AIWFClient(apiKey: 'secret');
```

## Examples

See `examples/` directory for complete examples:
- `examples/php-client/` - PHP translator
- `examples/python-client/` - Coming soon
- `examples/typescript-client/` - Coming soon

## See Also

- [HTTP Server Guide](./SERVE_GUIDE.md) - Server deployment
- [Getting Started](./GETTING_STARTED.md) - Basics
- [Examples](../examples/README.md) - Complete examples
- [YAML Specification](../generator/README.md) - Type system

## Roadmap

### v0.5 (Current)
- ✅ PHP client generator
- 🚧 Python client generator
- 🚧 TypeScript client generator

### v0.6 (Next)
- 📋 Go HTTP client
- 📋 Ruby client generator
- 📋 Streaming support

### v0.7 (Future)
- 📋 Java client generator
- 📋 C# client generator
- 📋 Rust client generator
- 📋 OpenAPI spec generation
