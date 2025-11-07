# PHP HTTP Client Example

This example demonstrates how to generate and use a PHP HTTP client for AIWF agents.

## Quick Start

### 1. Generate PHP Client

```bash
# Generate client from YAML config
aiwf sdk -f php-translator-example.yaml -l php -o ./generated

# This creates:
# php-client/client.php - Complete PHP client with types and methods
```

### 2. Start AIWF Server

```bash
export OPENAI_API_KEY="sk-..."
aiwf serve -f php-translator-example.yaml
```

### 3. Use the Client

```php
<?php
require_once 'php-client/client.php';

use AIWFClient\AIWFClient;
use AIWFClient\TranslateRequest;

// Create client
$client = new AIWFClient('http://127.0.0.1:8080');

// Make request
$request = new TranslateRequest(
    target_lang: 'es',
    text: 'Hello, world!'
);

$response = $client->translator($request);
echo $response->translated; // "Hola, mundo!"
```

## Generated Code Structure

### Type Classes

```php
class TranslateRequest {
    public function __construct(
        public string $text,
        public string $target_lang
    ) {}

    public function toArray(): array { ... }
    public static function fromArray(array $data): TranslateRequest { ... }
}

class TranslateResponse {
    public function __construct(
        public string $translated,
        public float $confidence,
        public string $source_lang
    ) {}

    public function toArray(): array { ... }
    public static function fromArray(array $data): TranslateResponse { ... }
}
```

### Client Class

```php
class AIWFClient {
    public function __construct(
        string $baseURL = 'http://127.0.0.1:8080',
        ?string $apiKey = null
    ) { ... }

    public function translator(TranslateRequest $request): TranslateResponse {
        // Sends HTTP POST to /agent/translator
        // Handles serialization/deserialization
    }
}
```

## Features

✅ **Type Safety** - Full PHP 8+ type hints
✅ **Zero Dependencies** - Uses built-in cURL
✅ **Auto Serialization** - `toArray()` / `fromArray()` methods
✅ **Error Handling** - Throws exceptions on HTTP errors
✅ **Authentication** - Optional API key support

## Requirements

- PHP 8.0+ (for constructor property promotion)
- cURL extension (usually enabled by default)

## Advanced Usage

### Custom Base URL

```php
$client = new AIWFClient('https://api.example.com');
```

### With Authentication

```php
$client = new AIWFClient(
    baseURL: 'https://api.example.com',
    apiKey: 'your-secret-key'
);
```

### Error Handling

```php
try {
    $response = $client->translator($request);
} catch (\Exception $e) {
    echo "Error: " . $e->getMessage();
}
```

## Generation Options

```bash
# Basic generation
aiwf sdk -f config.yaml -l php

# Custom output path
aiwf sdk -f config.yaml -l php -o ./src/AI

# Custom base URL
aiwf sdk -f config.yaml -l php --base-url https://api.example.com

# Force client type (default for PHP)
aiwf sdk -f config.yaml -l php --type client
```

## Comparison: HTTP Client vs Full SDK

| Feature | HTTP Client (This) | Full SDK |
|---------|-------------------|----------|
| Dependencies | cURL only | LLM SDK + Runtime |
| Code Size | ~100 lines | ~1000+ lines |
| API Keys | Server-side ✅ | Client-side ⚠️ |
| Maintenance | Minimal | High |
| Use Case | Web apps, APIs | Edge computing |

## Next Steps

- Add more agents to your YAML config
- Deploy server to production
- Generate clients for other languages (Python, TypeScript)
- Add authentication middleware

## Troubleshooting

### "Connection refused"
Make sure AIWF server is running:
```bash
aiwf serve -f config.yaml
```

### "HTTP Error 401"
Server requires authentication - set API key:
```php
$client = new AIWFClient(apiKey: getenv('AIWF_API_KEY'));
```

### "Type not found"
Regenerate client after changing YAML config:
```bash
aiwf sdk -f config.yaml -l php -o ./php-client
```
