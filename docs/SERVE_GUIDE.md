# AIWF Serve - HTTP API Server

Команда `aiwf serve` позволяет запустить HTTP API сервер, который автоматически превращает ваших AI агентов в REST API endpoints.

## Обзор

С помощью `aiwf serve` вы можете:
- 🚀 Мгновенно развернуть AI агентов как REST API
- 🔄 Автоматически генерировать и компилировать сервер
- 🧹 Использовать ephemeral mode без создания лишних файлов
- 🐛 Сохранять сгенерированный код для дебага

## Быстрый старт

```bash
# 1. Создайте конфигурацию с агентами
cat > my-agents.yaml << EOF
version: 0.3

types:
  GreetingRequest:
    name: string(1..100)

  GreetingResponse:
    message: string

assistants:
  greeter:
    use: openai
    model: gpt-4o-mini
    system_prompt: "Generate a friendly greeting"
    input_type: GreetingRequest
    output_type: GreetingResponse
EOF

# 2. Установите API ключ
export OPENAI_API_KEY="sk-..."

# 3. Запустите сервер
./aiwf serve -f my-agents.yaml

# Сервер запущен на http://127.0.0.1:8080
```

## Использование

### Базовая команда

```bash
aiwf serve -f config.yaml
```

### Опции

```bash
aiwf serve [flags]

Flags:
  -f, --file string     Путь к YAML конфигурации (default "config.yaml")
  -o, --output string   Директория для сохранения сгенерированного SDK (persistent mode)
  -p, --port int        Порт сервера (default 8080)
      --host string     Host для биндинга (default "127.0.0.1")
```

### Примеры

```bash
# Быстрый старт (ephemeral mode)
./aiwf serve -f config.yaml

# Кастомный порт
./aiwf serve -f config.yaml --port 3000

# Persistent mode (сохраняет сгенерированные файлы)
./aiwf serve -f config.yaml --output ./generated

# Публичный доступ
./aiwf serve -f config.yaml --host 0.0.0.0 --port 8080
```

## Режимы работы

### Ephemeral Mode (по умолчанию)

Не создает лишних файлов - идеально для разработки и тестирования.

```bash
./aiwf serve -f config.yaml
```

**Что происходит:**
1. SDK генерируется во временную директорию
2. Сервер компилируется и запускается
3. При остановке (Ctrl+C) временные файлы удаляются автоматически

**Преимущества:**
- ✅ Нет мусора в файловой системе
- ✅ Быстрый старт для разработки
- ✅ Изменения в YAML сразу применяются при перезапуске

### Persistent Mode

Сохраняет сгенерированные файлы - полезно для дебага и кастомизации.

```bash
./aiwf serve -f config.yaml --output ./generated
```

**Что происходит:**
1. SDK генерируется в указанную директорию
2. Сервер компилируется и запускается
3. При остановке файлы **остаются** для дебага

**Структура сгенерированных файлов:**
```
generated/
├── go.mod          # Go module
├── types.go        # Типы данных
├── agents.go       # AI агенты
├── service.go      # Сервис
└── cmd/
    └── server/
        └── main.go # HTTP сервер
```

**Преимущества:**
- ✅ Можно изучить сгенерированный код
- ✅ Можно модифицировать для продакшена
- ✅ Легче дебажить проблемы

## API Endpoints

Сервер автоматически создает следующие endpoints:

### Health Check

```bash
GET /health
```

**Response:**
```json
{"status": "ok"}
```

### Список агентов

```bash
GET /agents
```

**Response:**
```json
{
  "agents": [
    {"name": "greeter", "endpoint": "/agent/greeter"},
    {"name": "translator", "endpoint": "/agent/translator"}
  ]
}
```

### Вызов агента

```bash
POST /agent/{agent_name}
Content-Type: application/json

{
  "field1": "value1",
  "field2": "value2"
}
```

**Response:**
```json
{
  "data": {
    // Результат агента
  },
  "trace": {
    "step_name": "agent_name",
    "usage": {
      "prompt": 15,
      "completion": 25,
      "total": 40
    },
    "duration": "1.2s"
  }
}
```

## Аутентификация

Для защиты API установите переменную `API_KEY`:

```bash
export API_KEY="your-secret-key"
./aiwf serve -f config.yaml
```

Используйте ключ в запросах:

```bash
curl -X POST http://127.0.0.1:8080/agent/greeter \
  -H "X-API-Key: your-secret-key" \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice"}'
```

## Мониторинг

### Логирование

Сервер автоматически логирует все запросы:

```
2025-11-06T19:00:00Z POST /agent/greeter - 1.2s
2025-11-06T19:00:01Z POST /agent/translator - 850ms
2025-11-06T19:00:02Z GET /health - 5ms
```

### Trace информация

Каждый ответ включает trace с метриками:

```json
{
  "trace": {
    "step_name": "greeter",
    "usage": {
      "prompt": 15,
      "completion": 25,
      "total": 40
    },
    "attempts": 1,
    "duration": "1.234s"
  }
}
```

## Production Deployment

### Docker

```dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY . .

RUN go build -o aiwf ./cmd/aiwf

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/aiwf .
COPY config.yaml .

EXPOSE 8080

ENV OPENAI_API_KEY=""
ENV API_KEY=""

CMD ["./aiwf", "serve", "-f", "config.yaml", "--host", "0.0.0.0"]
```

**Запуск:**
```bash
docker build -t my-ai-server .
docker run -p 8080:8080 \
  -e OPENAI_API_KEY="sk-..." \
  -e API_KEY="secret" \
  my-ai-server
```

### Kubernetes

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: aiwf-config
data:
  config.yaml: |
    version: 0.3
    types:
      # ваши типы
    assistants:
      # ваши агенты

---
apiVersion: v1
kind: Secret
metadata:
  name: aiwf-secrets
type: Opaque
stringData:
  openai-api-key: sk-...
  api-key: your-secret-key

---
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
      - name: server
        image: my-ai-server:latest
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
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: config
        configMap:
          name: aiwf-config

---
apiVersion: v1
kind: Service
metadata:
  name: aiwf-server
spec:
  selector:
    app: aiwf-server
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer
```

### Systemd Service

```ini
[Unit]
Description=AIWF AI Agent Server
After=network.target

[Service]
Type=simple
User=aiwf
WorkingDirectory=/opt/aiwf
Environment="OPENAI_API_KEY=sk-..."
Environment="API_KEY=secret"
ExecStart=/opt/aiwf/aiwf serve -f /opt/aiwf/config.yaml --host 0.0.0.0
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

## Примеры клиентов

### Python

```python
import requests

class AIWFClient:
    def __init__(self, base_url, api_key=None):
        self.base_url = base_url
        self.headers = {"Content-Type": "application/json"}
        if api_key:
            self.headers["X-API-Key"] = api_key

    def call_agent(self, agent_name, data):
        response = requests.post(
            f"{self.base_url}/agent/{agent_name}",
            json=data,
            headers=self.headers
        )
        response.raise_for_status()
        return response.json()

    def list_agents(self):
        response = requests.get(
            f"{self.base_url}/agents",
            headers=self.headers
        )
        return response.json()

# Использование
client = AIWFClient("http://127.0.0.1:8080", api_key="secret")

result = client.call_agent("greeter", {"name": "Alice"})
print(result["data"])
```

### JavaScript/TypeScript

```typescript
class AIWFClient {
  constructor(
    private baseUrl: string,
    private apiKey?: string
  ) {}

  async callAgent<T>(agentName: string, data: any): Promise<T> {
    const headers: HeadersInit = {
      "Content-Type": "application/json",
    };

    if (this.apiKey) {
      headers["X-API-Key"] = this.apiKey;
    }

    const response = await fetch(
      `${this.baseUrl}/agent/${agentName}`,
      {
        method: "POST",
        headers,
        body: JSON.stringify(data),
      }
    );

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    return await response.json();
  }

  async listAgents() {
    const response = await fetch(`${this.baseUrl}/agents`);
    return await response.json();
  }
}

// Использование
const client = new AIWFClient("http://127.0.0.1:8080", "secret");

const result = await client.callAgent("greeter", { name: "Alice" });
console.log(result.data);
```

### Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type AIWFClient struct {
    BaseURL string
    APIKey  string
}

func (c *AIWFClient) CallAgent(agentName string, data interface{}) (map[string]interface{}, error) {
    jsonData, err := json.Marshal(data)
    if err != nil {
        return nil, err
    }

    req, err := http.NewRequest(
        "POST",
        fmt.Sprintf("%s/agent/%s", c.BaseURL, agentName),
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        return nil, err
    }

    req.Header.Set("Content-Type", "application/json")
    if c.APIKey != "" {
        req.Header.Set("X-API-Key", c.APIKey)
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return result, nil
}

func main() {
    client := &AIWFClient{
        BaseURL: "http://127.0.0.1:8080",
        APIKey:  "secret",
    }

    result, err := client.CallAgent("greeter", map[string]string{
        "name": "Alice",
    })
    if err != nil {
        panic(err)
    }

    fmt.Printf("Result: %+v\n", result["data"])
}
```

## Troubleshooting

### Ошибка: "OPENAI_API_KEY not set"

```bash
export OPENAI_API_KEY="sk-..."
```

### Ошибка: "Port already in use"

```bash
# Используйте другой порт
./aiwf serve -f config.yaml --port 3000

# Или найдите процесс, использующий порт
lsof -i :8080
kill <PID>
```

### Ошибка компиляции

```bash
# Убедитесь что Go установлен
go version

# Пересоберите aiwf
go build -o aiwf ./cmd/aiwf
```

### Медленный старт

Первый запуск может быть медленным из-за:
- Загрузки зависимостей (`go mod tidy`)
- Компиляции сервера

Последующие запуски будут быстрее благодаря кешу Go.

### Сервер не отвечает

Проверьте логи и убедитесь что:
1. API ключ установлен правильно
2. Порт доступен
3. Firewall не блокирует соединение

## Best Practices

### 1. Используйте переменные окружения

```bash
# .env file
OPENAI_API_KEY=sk-...
API_KEY=secret
PORT=8080
HOST=0.0.0.0
```

```bash
# Загрузка из .env
export $(cat .env | xargs)
./aiwf serve -f config.yaml
```

### 2. Настройте rate limiting

TODO: Будет добавлено в следующей версии

### 3. Мониторинг и алерты

TODO: Prometheus metrics будут добавлены

### 4. Версионирование API

Используйте разные конфигурации для разных версий:

```
configs/
├── v1/
│   └── config.yaml
└── v2/
    └── config.yaml
```

### 5. Load Balancing

Запустите несколько инстансов за Nginx/HAProxy:

```nginx
upstream aiwf_backend {
    server 127.0.0.1:8080;
    server 127.0.0.1:8081;
    server 127.0.0.1:8082;
}

server {
    listen 80;

    location / {
        proxy_pass http://aiwf_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## Roadmap

- [ ] Streaming responses (Server-Sent Events)
- [ ] Batch processing endpoints
- [ ] Prometheus metrics (`/metrics`)
- [ ] OpenAPI/Swagger documentation
- [ ] Rate limiting configuration
- [ ] Webhooks для async processing
- [ ] CORS configuration в YAML
- [ ] Authentication providers (JWT, OAuth)

## См. также

- [GETTING_STARTED.md](./GETTING_STARTED.md) - Основы работы с AIWF
- [Generator README](../generator/README.md) - Система типов и генерация SDK
- [API Server Template](../templates/api-server/) - Готовый пример
