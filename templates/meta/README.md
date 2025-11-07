# Meta-агенты для генерации конфигураций

Это специальные агенты, которые генерируют YAML конфигурации для AIWF на основе описания задач на естественном языке.

## Обзор

Meta-агенты используют двухэтапный подход:

1. **task_analyzer** - анализирует задачу и проектирует архитектуру
2. **yaml_generator** - генерирует валидную YAML конфигурацию

Оба агента работают в едином thread контексте для сохранения информации между шагами.

## Файлы

- `config_generator.yaml` - конфигурация meta-агентов

## Использование

### Через CLI (рекомендуется)

```bash
# Интерактивный режим
aiwf generate --interactive

# Быстрая генерация
aiwf generate -t "Your task description"
```

### Прямое использование (advanced)

```bash
# 1. Сгенерировать SDK для meta-агентов
aiwf sdk -f templates/meta/config_generator.yaml -o ./meta-sdk

# 2. Использовать в коде
cd meta-sdk
go mod tidy
```

**Пример кода:**

```go
package main

import (
    "context"
    "fmt"
    "os"

    meta "path/to/meta-sdk"
    "github.com/andranikuz/aiwf/providers/openai"
)

func main() {
    ctx := context.Background()

    // Initialize provider
    provider, _ := openai.NewClient(openai.ClientConfig{
        APIKey: os.Getenv("OPENAI_API_KEY"),
    })

    // Create service
    service := meta.NewService(provider)

    // Step 1: Analyze task
    taskDesc := "Create a sentiment analyzer for customer reviews"
    analysis, _, _ := service.Agents().TaskAnalyzer.Run(ctx, taskDesc)

    fmt.Printf("Complexity: %s\n", analysis.Complexity)
    fmt.Printf("Agents needed: %d\n", analysis.AgentCount)

    // Check if clarification needed
    if analysis.NeedsClarification {
        for _, q := range analysis.Questions {
            fmt.Printf("Question: %s\n", q.Question)
            // Handle user input...
        }
    }

    // Step 2: Generate YAML
    input := meta.GenerationInput{
        Analysis: formatAnalysis(analysis),
        RefinementInstructions: "Use Anthropic for all agents",
        UserAnswers: map[string]string{
            "output_format": "JSON",
        },
    }

    config, _, _ := service.Agents().YamlGenerator.Run(ctx, input)

    // Save generated YAML
    os.WriteFile("generated.yaml", []byte(config.YamlContent), 0644)

    fmt.Println("✓ Configuration generated!")
    fmt.Printf("  Types: %d\n", config.TypeCount)
    fmt.Printf("  Agents: %d\n", config.AgentCount)
}
```

## Типы данных

### Входные типы

**task_analyzer:**
- **Вход:** `string` - описание задачи на естественном языке
- **Выход:** `TaskAnalysis` - детальный анализ с архитектурой

**yaml_generator:**
- **Вход:** `GenerationInput` - анализ + уточнения
- **Выход:** `GeneratedConfig` - готовая YAML конфигурация

### Структуры

#### TaskAnalysis

```go
type TaskAnalysis struct {
    // Основной анализ
    Summary          string
    Complexity       string  // simple, medium, complex, very_complex
    ComplexityScore  float64 // 1.0-10.0

    // Агенты
    AgentCount       int
    Agents           []AgentSpec

    // Структуры данных
    DataTypesNeeded  []string
    RequiresThread   bool
    RequiresDialog   bool

    // Уточнения
    NeedsClarification bool
    Questions         []ClarificationQuestion

    // Дополнительно
    ArchitectureNotes    string
    ImplementationHints  []string
}
```

#### AgentSpec

```go
type AgentSpec struct {
    Name              string
    Role              string
    Provider          string  // openai, grok, anthropic
    Model             string
    InputDescription  string
    OutputDescription string
    NeedsDialog       bool
    MaxDialogRounds   int
    Reasoning         string
}
```

#### ClarificationQuestion

```go
type ClarificationQuestion struct {
    Question    string
    Reason      string
    Suggestions []string
}
```

#### GenerationInput

```go
type GenerationInput struct {
    Analysis               string            // JSON или текст
    RefinementInstructions string            // Опционально
    UserAnswers            map[string]string // Опционально
}
```

#### GeneratedConfig

```go
type GeneratedConfig struct {
    // Метаданные
    Version       string
    GeneratedAt   string

    // Структуры
    TypeCount     int
    AgentCount    int
    Types         []TypeDefinition
    Assistants    []AssistantConfig

    // Финальный YAML
    YamlContent   string  // Готовый YAML

    // Валидация
    ValidationStatus  string  // valid, invalid, warning
    ValidationNotes   []ValidationNote
    Notes             []string

    // Улучшения
    ImprovementSuggestions []string
}
```

## Примеры задач

### Простая задача

**Input:**
```
"Classify text into positive, negative, or neutral sentiment"
```

**Output (анализ):**
- Complexity: simple
- Agents: 1 (sentiment_classifier)
- No thread needed
- No clarification needed

### Средняя задача

**Input:**
```
"Create a content moderation system that checks for toxic language,
spam, and personal information in social media posts"
```

**Output (анализ):**
- Complexity: medium
- Agents: 3 (toxicity_detector, spam_classifier, pii_detector)
- Optional thread for context
- Possible clarification: "What should be done when multiple issues detected?"

### Сложная задача

**Input:**
```
"Build an interactive tutoring system that adapts difficulty based on
student performance, provides hints, and tracks progress across sessions"
```

**Output (анализ):**
- Complexity: complex
- Agents: 4-5 (difficulty_adjuster, hint_generator, progress_tracker, question_generator, feedback_provider)
- Thread required: yes (for session context)
- Dialog required: yes (up to 20 rounds)
- Clarification: "What subjects should be supported?"

## Настройка промптов

Meta-агенты используют детальные промпты для качественной генерации.

### task_analyzer промпт

Фокус на:
- Понимание требований
- Оценка сложности
- Проектирование архитектуры
- Выбор провайдеров и моделей
- Генерация уточняющих вопросов при необходимости

### yaml_generator промпт

Фокус на:
- Соответствие AIWF v0.3 спецификации
- Валидность типов и ограничений
- Качество system prompts
- Правильная конфигурация threads/dialogs
- Готовность к production

## Механизм уточнений

### Когда задаются вопросы

task_analyzer задает вопросы когда:
1. Описание задачи слишком общее
2. Неясны входные/выходные данные
3. Множество возможных решений
4. Нужна доменная информация

### Обработка ответов

```go
// Получаем анализ с вопросами
analysis, _, _ := service.Agents().TaskAnalyzer.Run(ctx, taskDesc)

if analysis.NeedsClarification {
    // Собираем ответы от пользователя
    answers := make(map[string]string)
    for _, q := range analysis.Questions {
        answer := getUserInput(q.Question, q.Suggestions)
        answers[q.Question] = answer
    }

    // Генерируем с учетом ответов
    input := meta.GenerationInput{
        Analysis:    marshalAnalysis(analysis),
        UserAnswers: answers,
    }

    config, _, _ := service.Agents().YamlGenerator.Run(ctx, input)
}
```

### Refinement instructions

На любом этапе можно внести уточнения:

```go
input := meta.GenerationInput{
    Analysis: marshalAnalysis(analysis),
    RefinementInstructions: `
        - Use Anthropic Claude for all agents
        - Add input validation
        - Support batch processing
    `,
}

config, _, _ := service.Agents().YamlGenerator.Run(ctx, input)
```

## Thread контекст

Оба агента работают в едином thread:

```yaml
threads:
  generation_context:
    provider: openai
    strategy: append  # Сохраняет весь контекст
    create: true
    ttl_hours: 2
```

**Преимущества:**
- yaml_generator видит оригинальную задачу
- yaml_generator видит весь анализ
- При refinement контекст сохраняется
- Можно итеративно улучшать

## Валидация

Сгенерированная конфигурация включает валидацию:

```go
if config.ValidationStatus != "valid" {
    for _, note := range config.ValidationNotes {
        fmt.Printf("[%s] %s: %s\n",
            note.Severity, note.Field, note.Message)
    }
}
```

## Best Practices

### 1. Детальные описания задач

**Хорошо:**
```
Create a spam detection system for emails.
Input: email subject and body (up to 10KB)
Output: spam score (0-1), classification (spam/ham), reasoning
Should handle English emails, prioritize precision over recall
```

**Плохо:**
```
Spam detector
```

### 2. Использование refinements

Начинайте с базового описания, затем уточняйте:

```go
// 1. Первая итерация
analysis1, _, _ := service.Agents().TaskAnalyzer.Run(ctx, "Create translator")

// 2. Смотрим результат, добавляем уточнения
input := meta.GenerationInput{
    Analysis: marshalAnalysis(analysis1),
    RefinementInstructions: "Add support for 10 languages, include confidence scores",
}

config, _, _ := service.Agents().YamlGenerator.Run(ctx, input)

// 3. Если нужно, еще одна итерация
if !satisfied {
    input.RefinementInstructions = "Change to use Anthropic Claude"
    config, _, _ = service.Agents().YamlGenerator.Run(ctx, input)
}
```

### 3. Проверка сложности

Используйте complexity score для принятия решений:

```go
if analysis.ComplexityScore > 7.0 {
    fmt.Println("⚠️  Complex task detected")
    fmt.Println("   Consider breaking into smaller subtasks")
}

if analysis.AgentCount > 5 {
    fmt.Println("⚠️  Many agents suggested")
    fmt.Println("   Review architecture for optimization")
}
```

## HTTP Server интеграция

Meta-агенты можно развернуть как HTTP API:

```bash
# Запускаем сервер
aiwf serve -f templates/meta/config_generator.yaml

# Анализируем задачу
curl -X POST http://127.0.0.1:8080/agent/task_analyzer \
  -H "Content-Type: application/json" \
  -d '"Create a sentiment analyzer"'

# Генерируем YAML
curl -X POST http://127.0.0.1:8080/agent/yaml_generator \
  -H "Content-Type: application/json" \
  -d '{
    "analysis": "...",
    "refinement_instructions": "Use gpt-4o-mini",
    "user_answers": {}
  }'
```

## Troubleshooting

### "Type not found" при генерации SDK

Убедитесь что все типы определены в конфигурации:

```bash
aiwf validate -f templates/meta/config_generator.yaml
```

### Некачественная генерация

Попробуйте:
1. Более детальное описание задачи
2. Использовать refinement instructions
3. Ответить на уточняющие вопросы
4. Итеративно улучшать результат

### Невалидный YAML

Даже при невалидном YAML он сохраняется для ручного исправления:

```bash
# Проверяем ошибки
aiwf validate -f generated-config.yaml

# Исправляем вручную
vim generated-config.yaml
```

## Расширение

Можно создать собственные meta-агенты:

```yaml
assistants:
  # Ваш кастомный анализатор
  custom_analyzer:
    use: anthropic
    model: claude-sonnet-4-5
    system_prompt: "Your custom analysis logic"
    input_type: string
    output_type: CustomAnalysis

  # Специализированный генератор
  specialized_generator:
    use: openai
    model: gpt-4o
    system_prompt: "Generate configs for specific domain"
    input_type: CustomAnalysis
    output_type: GeneratedConfig
```

## См. также

- [Generate Guide](../../docs/GENERATE_GUIDE.md) - CLI документация
- [Getting Started](../../docs/GETTING_STARTED.md) - Основы AIWF
- [Generator Documentation](../../generator/README.md) - Система типов
