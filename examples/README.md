# AIWF Examples

This directory contains examples demonstrating different ways to use AIWF.

## Examples Overview

### 1. Go SDK (Full SDK)

**Path:** `go-sdk/`

Full Go SDK with embedded runtime and providers. Works standalone without HTTP server.

```bash
cd examples/go-sdk
aiwf sdk -f agents.yaml -o ./generated
go run main.go
```

**Use cases:**
- CLI tools
- Edge computing
- Serverless functions
- Standalone applications

**Features:**
- ✅ No HTTP server needed
- ✅ Direct LLM API calls
- ✅ Type-safe Go code
- ⚠️ Requires API keys in application

### 2. PHP HTTP Client

**Path:** `php-client/`

Lightweight PHP HTTP client that connects to AIWF server.

```bash
# 1. Generate client
cd examples/php-client
aiwf sdk -f php-translator-example.yaml -l php -o ./generated

# 2. Start server (in another terminal)
export OPENAI_API_KEY="sk-..."
aiwf serve -f php-translator-example.yaml

# 3. Use client
php php-translator-usage.php
```

**Use cases:**
- Web applications
- WordPress plugins
- Laravel/Symfony apps
- PHP APIs

**Features:**
- ✅ Minimal dependencies (cURL only)
- ✅ Type-safe PHP 8+ code
- ✅ API keys on server (secure)
- ✅ ~100 lines of code
- ⚠️ Requires AIWF server

See [php-client/README_PHP.md](php-client/README_PHP.md) for detailed documentation.

## Comparison: Full SDK vs HTTP Client

| Feature | Full SDK (Go) | HTTP Client (PHP/Python/TS) |
|---------|---------------|------------------------------|
| **Dependencies** | LLM SDK + Runtime | HTTP library only |
| **Code Size** | ~1000+ lines | ~100 lines |
| **API Keys** | In application | On server |
| **Server Required** | No | Yes |
| **Latency** | Minimal | +50-200ms (network) |
| **Maintenance** | High (per language) | Low (one server) |
| **Monitoring** | Distributed | Centralized |
| **Caching** | Per instance | Shared |

## Quick Command Reference

### Generate Full Go SDK

```bash
aiwf sdk -f config.yaml -l go -o ./sdk
```

### Generate HTTP Client

```bash
# PHP
aiwf sdk -f config.yaml -l php -o ./client.php

# Python (coming soon)
aiwf sdk -f config.yaml -l python -o ./client.py

# TypeScript (coming soon)
aiwf sdk -f config.yaml -l typescript -o ./client.ts

# Go HTTP client (instead of full SDK)
aiwf sdk -f config.yaml -l go --type client -o ./client.go
```

### Start HTTP Server

```bash
export OPENAI_API_KEY="sk-..."
aiwf serve -f config.yaml
```

## Supported Languages

| Language | Full SDK | HTTP Client | Status |
|----------|----------|-------------|--------|
| Go | ✅ | ✅ | Ready |
| PHP | ❌ | ✅ | Ready |
| Python | ❌ | 🚧 | Coming soon |
| TypeScript | ❌ | 🚧 | Coming soon |
| Ruby | ❌ | 📋 | Planned |
| Java | ❌ | 📋 | Planned |
| C# | ❌ | 📋 | Planned |

## Creating Your Own Example

### 1. Define YAML Configuration

```yaml
version: 0.3

types:
  MyRequest:
    text: string(1..1000)

  MyResponse:
    result: string
    confidence: number(0..1)

assistants:
  my_agent:
    use: openai
    model: gpt-4o-mini
    system_prompt: Your prompt here
    input_type: $MyRequest
    output_type: $MyResponse
```

### 2. Generate Client

```bash
# Choose your language
aiwf sdk -f my-config.yaml -l php -o ./client
```

### 3. Start Server

```bash
export OPENAI_API_KEY="sk-..."
aiwf serve -f my-config.yaml
```

### 4. Use Client

See language-specific examples in subdirectories.

## Best Practices

### For Development

1. **Use HTTP clients** for most use cases
2. **Start with PHP/Python** - easy to test and iterate
3. **Use ephemeral mode** for `aiwf serve` (default)
4. **Keep YAML configs simple** - add complexity gradually

### For Production

1. **Deploy AIWF server** as a microservice
2. **Use Docker** for consistent deployments
3. **Set API keys** via environment variables
4. **Enable authentication** with `API_KEY` env var
5. **Monitor server** logs and metrics
6. **Scale horizontally** - run multiple server instances

## Troubleshooting

### "Connection refused"

Server not running. Start it:
```bash
aiwf serve -f config.yaml
```

### "Type not found"

Regenerate client after YAML changes:
```bash
aiwf sdk -f config.yaml -l php -o ./client
```

### "API key not set"

Set environment variable:
```bash
export OPENAI_API_KEY="sk-..."
# or
export GROK_API_KEY="..."
# or
export ANTHROPIC_API_KEY="..."
```

## Next Steps

- Read [Getting Started Guide](../docs/GETTING_STARTED.md)
- Learn about [HTTP Server](../docs/SERVE_GUIDE.md)
- Explore [YAML specification](../generator/README.md)
- Check [available templates](../templates/README.md)

## Contributing

To add a new example:

1. Create directory: `examples/my-example/`
2. Add YAML config
3. Add README with instructions
4. Add usage example
5. Update this README

See existing examples for reference.
