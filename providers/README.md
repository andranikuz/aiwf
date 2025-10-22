# AIWF Providers

LLM провайдеры для AIWF Framework.

## Поддерживаемые провайдеры

### OpenAI
Полнофункциональный провайдер с JSON Schema и потоковым выводом.

```go
import "github.com/andranikuz/aiwf/providers/openai"

client, err := openai.NewClient(openai.ClientConfig{
    APIKey: "your-api-key",
})
service := sdk.NewService(client)
```

**Особенности:**
- Структурированный вывод через JSON Schema
- Потоковые ответы
- Управление тредами

### Grok (xAI)
Провайдер для Grok - новой LLM от xAI.

```go
import "github.com/andranikuz/aiwf/providers/grok"

client, err := grok.NewClient(grok.ClientConfig{
    APIKey: "xai-...",
})
service := sdk.NewService(client)
```

**Особенности:**
- Chat API для текстовых ответов
- JSON входные данные
- Отличная производительность

### Anthropic
Провайдер для Claude от Anthropic.

```go
import "github.com/andranikuz/aiwf/providers/anthropic"

client, err := anthropic.NewClient(anthropic.ClientConfig{
    APIKey: "sk-ant-...",
})
service := sdk.NewService(client)
```

**Особенности:**
- Messages API
- Системные промпты
- Гибкое контекстное окно

## Использование

```bash
# OpenAI
export OPENAI_API_KEY="sk-..."

# Grok
export GROK_API_KEY="xai-..."

# Anthropic
export ANTHROPIC_API_KEY="sk-ant-..."
```
