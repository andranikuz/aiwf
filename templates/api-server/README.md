# API Server Example

Этот пример демонстрирует, как создать REST API сервер с несколькими AI агентами.

## Возможности

- 🔍 **Текстовый анализ** - sentiment, summary, keywords, language detection
- 🌍 **Перевод** - перевод между 7 языками
- ✍️ **Генерация контента** - создание текста в разных стилях
- ❓ **Q&A система** - ответы на вопросы с контекстом

## Быстрый старт

### 1. Установите API ключ

```bash
export OPENAI_API_KEY="sk-..."
```

### 2. Запустите сервер

```bash
# Из корня проекта
./aiwf serve -f templates/api-server/config.yaml

# Или с кастомным портом
./aiwf serve -f templates/api-server/config.yaml --port 3000

# Для дебага (сохраняет сгенерированные файлы)
./aiwf serve -f templates/api-server/config.yaml --output ./generated
```

Сервер запустится на `http://127.0.0.1:8080`

### 3. Используйте API

#### Health Check

```bash
curl http://127.0.0.1:8080/health
```

**Response:**
```json
{"status": "ok"}
```

#### Список агентов

```bash
curl http://127.0.0.1:8080/agents
```

**Response:**
```json
{
  "agents": [
    {"name": "text_analyzer", "endpoint": "/agent/text_analyzer"},
    {"name": "translator", "endpoint": "/agent/translator"},
    {"name": "content_generator", "endpoint": "/agent/content_generator"},
    {"name": "qa_system", "endpoint": "/agent/qa_system"}
  ]
}
```

## API Endpoints

### 1. Текстовый анализ

**Endpoint:** `POST /agent/text_analyzer`

**Request:**
```bash
curl -X POST http://127.0.0.1:8080/agent/text_analyzer \
  -H "Content-Type: application/json" \
  -d '{
    "text": "This is an amazing product! I love it so much.",
    "analysis_type": "sentiment"
  }'
```

**Response:**
```json
{
  "data": {
    "analysis_type": "sentiment",
    "result": "positive",
    "confidence": 0.95,
    "metadata": {
      "sentiment_score": 0.9,
      "key_phrases": ["amazing product", "love it"]
    }
  },
  "trace": {
    "step_name": "text_analyzer",
    "usage": {
      "prompt": 45,
      "completion": 32,
      "total": 77
    },
    "duration": "1.2s"
  }
}
```

**Типы анализа:**
- `sentiment` - анализ тональности
- `summary` - краткое содержание
- `keywords` - ключевые слова
- `language` - определение языка

### 2. Перевод

**Endpoint:** `POST /agent/translator`

**Request:**
```bash
curl -X POST http://127.0.0.1:8080/agent/translator \
  -H "Content-Type: application/json" \
  -d '{
    "text": "Hello, world!",
    "source_lang": "en",
    "target_lang": "es"
  }'
```

**Response:**
```json
{
  "data": {
    "original_text": "Hello, world!",
    "translated_text": "¡Hola, mundo!",
    "source_lang": "en",
    "target_lang": "es",
    "confidence": 0.98
  },
  "trace": {
    "usage": {
      "total": 45
    }
  }
}
```

**Поддерживаемые языки:**
- `en` - English
- `es` - Español
- `fr` - Français
- `de` - Deutsch
- `ru` - Русский
- `zh` - 中文
- `ja` - 日本語

### 3. Генерация контента

**Endpoint:** `POST /agent/content_generator`

**Request:**
```bash
curl -X POST http://127.0.0.1:8080/agent/content_generator \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Write about artificial intelligence",
    "style": "professional",
    "length": "medium",
    "topic": "AI in Healthcare"
  }'
```

**Response:**
```json
{
  "data": {
    "content": "Artificial Intelligence is revolutionizing healthcare...",
    "word_count": 250,
    "style_used": "professional"
  },
  "trace": {
    "usage": {
      "total": 350
    }
  }
}
```

**Стили:**
- `professional` - профессиональный
- `casual` - неформальный
- `creative` - креативный
- `technical` - технический
- `marketing` - маркетинговый

**Длина:**
- `short` - короткий (100-200 слов)
- `medium` - средний (200-400 слов)
- `long` - длинный (400-600 слов)

### 4. Вопрос-ответ

**Endpoint:** `POST /agent/qa_system`

**Request:**
```bash
curl -X POST http://127.0.0.1:8080/agent/qa_system \
  -H "Content-Type: application/json" \
  -d '{
    "question": "What is machine learning?",
    "context": "Machine learning is a subset of artificial intelligence...",
    "max_answer_length": 200
  }'
```

**Response:**
```json
{
  "data": {
    "answer": "Machine learning is a branch of AI that enables systems to learn from data...",
    "confidence": 0.92,
    "sources": ["provided_context", "general_knowledge"]
  },
  "trace": {
    "usage": {
      "total": 120
    }
  }
}
```

## Аутентификация (опционально)

Добавьте API ключ для защиты endpoints:

```bash
export API_KEY="your-secret-key"
./aiwf serve -f templates/api-server/config.yaml
```

Затем используйте ключ в запросах:

```bash
curl -X POST http://127.0.0.1:8080/agent/text_analyzer \
  -H "X-API-Key: your-secret-key" \
  -H "Content-Type: application/json" \
  -d '{"text": "...", "analysis_type": "sentiment"}'
```

## Production Deployment

### Docker

```dockerfile
FROM golang:1.24-alpine

WORKDIR /app
COPY . .

RUN go build -o aiwf ./cmd/aiwf

EXPOSE 8080

CMD ["./aiwf", "serve", "-f", "templates/api-server/config.yaml", "--host", "0.0.0.0"]
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: aiwf-server
spec:
  replicas: 3
  selector:
    matchLabels:
      app: aiwf-server
  template:
    metadata:
      labels:
        app: aiwf-server
    spec:
      containers:
      - name: aiwf-server
        image: aiwf-server:latest
        ports:
        - containerPort: 8080
        env:
        - name: OPENAI_API_KEY
          valueFrom:
            secretKeyRef:
              name: aiwf-secrets
              key: openai-api-key
        - name: API_KEY
          valueFrom:
            secretKeyRef:
              name: aiwf-secrets
              key: api-key
```

## Мониторинг

### Prometheus Metrics

TODO: Добавить /metrics endpoint

### Логирование

Сервер логирует все запросы:

```
2025-11-06T19:00:00Z POST /agent/text_analyzer - 1.2s
2025-11-06T19:00:01Z POST /agent/translator - 850ms
```

## Кастомизация

Вы можете:
- Добавить новых агентов в `config.yaml`
- Изменить промпты для улучшения результатов
- Настроить `temperature` и `max_tokens`
- Использовать другие LLM провайдеры (Grok, Anthropic)

## Производительность

Типичные времена ответа:
- Текстовый анализ: 0.5-1.5s
- Перевод: 0.5-1s
- Генерация контента: 2-4s (зависит от длины)
- Q&A: 1-2s

## Troubleshooting

### Ошибка: "OPENAI_API_KEY not set"
```bash
export OPENAI_API_KEY="sk-..."
```

### Ошибка: "Port already in use"
```bash
./aiwf serve -f config.yaml --port 3000
```

### Ошибка компиляции
```bash
# Убедитесь что Go установлен
go version

# Пересоберите aiwf
go build -o aiwf ./cmd/aiwf
```

## Примеры использования

### Python клиент

```python
import requests

def analyze_text(text, analysis_type):
    response = requests.post(
        "http://127.0.0.1:8080/agent/text_analyzer",
        json={
            "text": text,
            "analysis_type": analysis_type
        }
    )
    return response.json()

result = analyze_text("I love this product!", "sentiment")
print(result["data"]["result"])  # positive
```

### JavaScript клиент

```javascript
async function translateText(text, sourceLang, targetLang) {
  const response = await fetch('http://127.0.0.1:8080/agent/translator', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
      text,
      source_lang: sourceLang,
      target_lang: targetLang
    })
  });
  return await response.json();
}

const result = await translateText('Hello!', 'en', 'es');
console.log(result.data.translated_text);  // ¡Hola!
```

## Лицензия

MIT
