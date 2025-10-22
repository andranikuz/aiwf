# AIWF Getting Started Guide

Этот гайд поможет вам создать свой первый AI-агент за 10 минут.

## 1. Установка

```bash
# Клонируйте репозиторий
git clone https://github.com/andranikuz/aiwf.git
cd aiwf

# Сборка CLI
go build -o aiwf ./cmd/aiwf

# Проверка установки
./aiwf validate -f examples/agents.yaml
```

## 2. Получение API ключей

Выберите провайдера:

### OpenAI
```bash
# 1. Создайте аккаунт: https://platform.openai.com
# 2. Получите API ключ: https://platform.openai.com/api-keys
export OPENAI_API_KEY="sk-..."
```

### Grok (xAI)
```bash
# 1. Создайте аккаунт: https://console.x.ai/
# 2. Получите API ключ
export GROK_API_KEY="xai-..."
```

### Anthropic (Claude)
```bash
# 1. Создайте аккаунт: https://console.anthropic.com/
# 2. Получите API ключ
export ANTHROPIC_API_KEY="sk-ant-..."
```

## 3. Создайте свой первый агент

### Шаг 1: Напишите YAML конфигурацию

Создайте файл `my-agent.yaml`:

```yaml
version: 0.3

types:
  # Входные данные
  TranslationRequest:
    text: string(1..1000)
    source_lang: enum(en, es, fr, de, ru)
    target_lang: enum(en, es, fr, de, ru)

  # Выходные данные
  TranslationResult:
    translated_text: string
    confidence: number(0.0..1.0)

assistants:
  translator:
    use: openai              # Или: grok, anthropic
    model: gpt-4o-mini       # Или: gpt-4, claude-3-sonnet, grok-beta
    input_type: TranslationRequest
    output_type: TranslationResult
    max_tokens: 500
    temperature: 0.3
    system_prompt: |
      You are a professional translator. Translate the provided text
      accurately while preserving meaning and tone.
```

### Шаг 2: Валидируйте конфигурацию

```bash
./aiwf validate -f my-agent.yaml
# ✓ YAML валиден. Assistants: 1
```

### Шаг 3: Сгенерируйте SDK

```bash
./aiwf sdk -f my-agent.yaml -o ./generated --package mypkg
# ✓ SDK сгенерирован в ./generated
```

### Шаг 4: Используйте агента в коде

Создайте файл `main.go`:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    sdk "./generated"  // или "your-module/generated" если SDK в отдельном модуле
    "github.com/andranikuz/aiwf/providers/openai"
)

func main() {
    // Создайте клиента провайдера
    client, err := openai.NewClient(openai.ClientConfig{
        APIKey: os.Getenv("OPENAI_API_KEY"),
    })
    if err != nil {
        log.Fatal(err)
    }

    // Создайте сервис с агентами
    service := sdk.NewService(client)

    // Вызовите агента
    ctx := context.Background()
    result, trace, err := service.Agents().Translator.Run(ctx, sdk.TranslationRequest{
        Text:       "Hello, world!",
        SourceLang: "en",
        TargetLang: "es",
    })

    if err != nil {
        log.Fatal(err)
    }

    // Используйте результат
    fmt.Printf("Translation: %s\n", result.TranslatedText)
    fmt.Printf("Confidence: %.2f\n", result.Confidence)
    fmt.Printf("Tokens used: %d\n", trace.Usage.Total)
}
```

### Шаг 5: Запустите код

```bash
export OPENAI_API_KEY="sk-..."  # Установите ваш API ключ!
go run main.go
# Translation: ¡Hola, mundo!
# Confidence: 0.98
# Tokens used: 42
```

## Обработка ошибок и контекст

### Контекст (context.Context)

В примере используется `context.Background()` для синхронных вызовов без timeout'а. Для более контролируемого использования:

```go
import "time"

// Вызов с timeout 30 секунд
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, trace, err := service.Agents().Translator.Run(ctx, request)
if err != nil {
    log.Printf("Error: %v", err)  // timeout или ошибка API
}
```

### Обработка ошибок

```go
result, trace, err := service.Agents().Translator.Run(ctx, request)
if err != nil {
    // Проверьте environment variable
    if os.Getenv("OPENAI_API_KEY") == "" {
        log.Fatal("OPENAI_API_KEY not set")
    }
    // Проверьте сетевое соединение
    log.Fatalf("API call failed: %v", err)
}

// Используйте результат и метрики
fmt.Printf("Tokens used: %d\n", trace.Usage.Total)
fmt.Printf("Time taken: %v\n", trace.Duration)
```

## 4. Понимание типов

### Базовые типы

```yaml
types:
  SimpleExample:
    # Скалярные типы
    name: string           # Любая строка
    count: int             # Целое число
    ratio: number          # Число с точкой
    active: bool           # true/false
    id: uuid               # UUID
    created: datetime      # ISO 8601

    # Ограничения
    short_text: string(1..50)      # 1-50 символов
    age: int(0..150)               # 0-150
    confidence: number(0.0..1.0)   # 0.0-1.0

    # Перечисления
    status: enum(active, inactive, pending)

    # Массивы
    tags: string[]                 # Массив строк
    scores: number[]               # Массив чисел

    # Словари
    metadata: map(string, any)     # map[string]interface{}
    scores_map: map(string, number) # map[string]float64

    # Ссылки на другие типы
    author: User                   # Вложенный тип User
    items: Item[]                  # Массив объектов Item
```

### Сложные примеры

```yaml
types:
  # Вложенная структура
  User:
    id: uuid
    name: string(1..100)
    email: string
    age: int(0..150)
    role: enum(admin, user, guest)

  BlogPost:
    id: uuid
    title: string(1..200)
    content: string(10..10000)
    author: User                   # Вложенный User
    tags: string[](max:5)          # До 5 тегов
    metadata: map(string, any)     # Дополнительные данные
    published: bool
    created: datetime
```

## 5. Конфигурация агентов

### Обязательные поля

```yaml
assistants:
  my_agent:
    use: openai              # Провайдер: openai, grok, anthropic
    model: gpt-4o-mini       # Модель LLM
    input_type: MyInput      # Тип входных данных
    system_prompt: "..."     # Инструкция для модели
```

### Опциональные поля

```yaml
assistants:
  my_agent:
    # Вывод типа (дефолт: string)
    output_type: MyOutput    # Опционально, дефолт: string

    # Параметры генерации (дефолты: max_tokens=2000, temperature=0.7)
    max_tokens: 1000         # Максимум токенов в ответе
    temperature: 0.5         # Случайность (0.0-1.0)

    # Поддержка многораундных диалогов
    thread:
      use: my_thread         # Название конфигурации треда
      strategy: append       # Как добавлять сообщения

    # Диалоговый режим
    dialog:
      max_rounds: 5          # Максимум раундов диалога
```

### Примеры конфигураций

**Простой текстовый агент (Grok)**
```yaml
assistants:
  writer:
    use: grok
    model: grok-beta
    system_prompt: "You are a creative writer"
    input_type: WritingPrompt
    # output_type не указан → дефолт string
    max_tokens: 2000
    temperature: 0.9
```

**Структурированный вывод (OpenAI)**
```yaml
assistants:
  analyzer:
    use: openai
    model: gpt-4o-mini
    system_prompt: "Analyze the data"
    input_type: DataInput
    output_type: AnalysisResult     # Структурированный JSON
    max_tokens: 1500
    temperature: 0.3                 # Более детерминированно
```

**Многораундный диалог (OpenAI + Threads)**
```yaml
assistants:
  support_bot:
    use: openai
    model: gpt-4o-mini
    system_prompt: "You are a helpful support agent"
    input_type: CustomerQuery
    output_type: SupportResponse
    thread:
      use: support_thread
      strategy: append              # Сохраняет контекст
    dialog:
      max_rounds: 10
    max_tokens: 1000
    temperature: 0.5
```

## 6. Структура проекта

После генерации SDK вы получите:

```
generated/
├── agents.go       # Типизированные агенты
├── service.go      # Сервис со всеми агентами
├── types.go        # Структуры данных и валидаторы
└── go.mod          # Зависимости
```

## 7. Следующие шаги

- 📚 [Документация Generator](../generator/README.md) - углубленное описание системы типов
- 🔧 [Документация Runtime](../runtime/README.md) - API интерфейсы
- 🔌 [Документация Providers](../providers/README.md) - поддерживаемые LLM
- 📋 [Примеры](../examples/) - полные рабочие примеры
- 🎯 [Шаблоны](../templates/) - готовые конфигурации

## 8. Часто задаваемые вопросы

### Как переключиться между провайдерами?

Просто измените `use:` в конфигурации и пересгенерируйте SDK:

```yaml
# Было
assistants:
  my_agent:
    use: openai

# Стало
assistants:
  my_agent:
    use: grok
```

### Какой провайдер выбрать?

- **OpenAI** (gpt-4o-mini) - универсальный, хороший баланс цены и качества
- **Grok** (grok-beta) - отличен для текстовой генерации, быстрый
- **Anthropic** (claude-3-sonnet) - мощный для анализа, хороша обработка ошибок

### Какие модели доступны?

Каждый провайдер предоставляет разные модели. Актуальные списки:

- **OpenAI модели**: https://platform.openai.com/docs/models
  - `gpt-4o` (новейшая и самая мощная)
  - `gpt-4o-mini` (быстрая и дешевая)
  - `gpt-4` (мощная, но медленнее)

- **Grok модели**: https://docs.x.ai/docs/models
  - `grok-beta` (основная модель)

- **Anthropic модели**: https://docs.anthropic.com/en/docs/about/models/overview
  - `claude-3-5-sonnet` (мощная)
  - `claude-3-opus` (самая мощная)
  - `claude-3-haiku` (быстрая)

Используйте более новые/мощные модели для сложных задач, но помните о цене!

### Как добавить новое поле в тип?

Отредактируйте YAML, пересгенерируйте и Go код будет обновлен автоматически.

### Как кастомизировать промпт?

Отредактируйте `system_prompt` в конфигурации агента.

### Как обрабатывать ошибки?

```go
result, trace, err := service.Agents().MyAgent.Run(ctx, input)
if err != nil {
    log.Printf("Error: %v", err)
    return
}
```

## Версионирование

Текущая версия: **v1.8.0**

- ✅ MaxTokens и Temperature конфигурация
- ✅ Опциональный output_type
- ✅ OpenAI, Grok, Anthropic провайдеры
- ✅ Thread поддержка для диалогов
