# AIWF Generate - AI-Powered Configuration Generator

Команда `aiwf generate` использует AI для автоматической генерации YAML конфигураций на основе описания задачи на естественном языке.

## Обзор

С помощью `aiwf generate` вы можете:
- 🤖 Генерировать конфигурации из текстового описания
- 💬 Интерактивно уточнять требования
- ✅ Получать готовые к использованию YAML файлы
- 🔄 Итеративно улучшать конфигурацию

## Быстрый старт

### Интерактивный режим (рекомендуется)

```bash
aiwf generate --interactive
```

Вас проведут через пошаговый процесс:
1. Описание задачи
2. Анализ требований (с уточняющими вопросами при необходимости)
3. Просмотр и утверждение архитектуры
4. Генерация YAML
5. Просмотр и возможность внесения правок
6. Сохранение финального файла

### Быстрая генерация

```bash
# Из командной строки
aiwf generate -t "Create a sentiment analyzer for customer reviews"

# Из файла
aiwf generate --task-file task.txt -o my-config.yaml
```

## Использование

### Базовая команда

```bash
aiwf generate [flags]
```

### Флаги

| Флаг | Краткая форма | Описание | По умолчанию |
|------|---------------|----------|--------------|
| `--interactive` | `-i` | Интерактивный режим с подтверждениями | `false` |
| `--task` | `-t` | Описание задачи (строка) | - |
| `--task-file` | - | Файл с описанием задачи | - |
| `--output` | `-o` | Выходной YAML файл | `generated-config.yaml` |
| `--provider` | - | LLM провайдер | `openai` |
| `--api-key` | - | API ключ | из ENV |

### Переменные окружения

```bash
export OPENAI_API_KEY="sk-..."
export GROK_API_KEY="..."
export ANTHROPIC_API_KEY="..."
```

## Примеры

### 1. Создание системы модерации контента

```bash
aiwf generate --interactive
```

```
📝 Describe your task:
> Create a content moderation system for social media posts.
  Need to detect toxic language, spam, and personal information.

📊 Analysis:
   Complexity: MEDIUM
   Suggested agents: 3

   1. toxicity_detector (openai/gpt-4o-mini)
      Role: Detect toxic or offensive language

   2. spam_classifier (openai/gpt-4o-mini)
      Role: Identify spam and promotional content

   3. pii_detector (openai/gpt-4o)
      Role: Detect personal identifiable information

✅ Continue to generation? [Y/n]
```

### 2. Multi-language translator

```bash
aiwf generate -t "Create a translator that supports English, Spanish, French, German, and Russian. Should detect source language automatically and provide confidence scores."
```

**Результат:**
```yaml
version: 0.3

types:
  TranslationRequest:
    text: string(1..5000)
    target_language: enum(en, es, fr, de, ru)

  TranslationResult:
    original_text: string
    detected_language: string
    translated_text: string
    target_language: string
    confidence: number(0.0..1.0)

assistants:
  translator:
    use: openai
    model: gpt-4o
    system_prompt: |
      You are a professional translator.
      Automatically detect the source language and translate to the target.
      Provide confidence scores for translation quality.
    input_type: TranslationRequest
    output_type: TranslationResult
    max_tokens: 2000
    temperature: 0.3
```

### 3. Customer support chatbot

```bash
aiwf generate --interactive
```

```
📝 Task:
> Customer support bot that can handle common queries,
  escalate complex issues, and maintain conversation context

❓ Clarification needed:
   1. What type of queries? (e.g., technical, billing, general)
   Your answer: Technical support for software products

   2. How many conversation rounds expected?
   Your answer: Up to 10 rounds per conversation

✓ Generating configuration with dialog support and thread management...
```

**Результат:**
```yaml
version: 0.3

threads:
  support_conversation:
    provider: openai
    strategy: append
    create: true
    ttl_hours: 24

types:
  CustomerQuery:
    message: string(1..2000)
    user_id: string
    session_id: string

  SupportResponse:
    reply: string
    escalate: bool
    resolved: bool
    next_steps: string[]

assistants:
  support_agent:
    use: openai
    model: gpt-4o
    thread:
      use: support_conversation
      strategy: append
    dialog:
      max_rounds: 10
    system_prompt: |
      You are a technical support agent for software products.
      Help users with technical issues, provide clear solutions,
      and escalate complex problems when needed.
    input_type: CustomerQuery
    output_type: SupportResponse
```

### 4. Data analysis pipeline

```bash
aiwf generate -t "Create a data analysis pipeline: first validate CSV data, then analyze for anomalies, finally generate a summary report"
```

**Результат:**
```yaml
version: 0.3

threads:
  analysis_pipeline:
    provider: openai
    strategy: append
    create: true

types:
  CSVData:
    content: string(10..100000)
    column_names: string[]

  ValidationResult:
    is_valid: bool
    errors: string[]
    row_count: int

  AnomalyReport:
    anomalies_found: int
    anomaly_details: $Anomaly[]

  Anomaly:
    row_number: int
    column: string
    description: string
    severity: enum(low, medium, high)

  SummaryReport:
    total_rows: int
    validation_status: string
    anomalies_summary: string
    recommendations: string[]

assistants:
  validator:
    use: openai
    model: gpt-4o-mini
    thread:
      use: analysis_pipeline
    system_prompt: "Validate CSV data structure and format"
    input_type: CSVData
    output_type: ValidationResult
    temperature: 0.2

  anomaly_detector:
    use: openai
    model: gpt-4o
    thread:
      use: analysis_pipeline
    system_prompt: "Detect anomalies and outliers in validated data"
    input_type: ValidationResult
    output_type: AnomalyReport
    temperature: 0.3

  report_generator:
    use: openai
    model: gpt-4o
    thread:
      use: analysis_pipeline
    system_prompt: "Generate comprehensive summary report"
    input_type: AnomalyReport
    output_type: SummaryReport
    temperature: 0.4
```

## Интерактивный процесс

### Шаг 1: Описание задачи

```
📝 Describe your task in natural language.
   What do you want your AI agent(s) to do?

> _
```

Опишите задачу как можно подробнее:
- Что должна делать система
- Какие входные данные
- Какие выходные данные
- Особые требования

### Шаг 2: Анализ

```
📊 Task Analysis

   Complexity: MEDIUM
   Suggested agents: 2

   1. input_validator (openai/gpt-4o-mini)
      Role: Validate and sanitize user input
      Model: gpt-4o-mini
      Why: Fast and cost-effective for validation

   2. main_processor (openai/gpt-4o)
      Role: Process validated input and generate results
      Model: gpt-4o
      Why: More capable for complex processing

   💡 Architecture: Two-stage pipeline with validation
   💡 Implementation hints:
      • Use thread for context sharing
      • Consider adding error handling types
```

#### Уточняющие вопросы (если нужны)

```
❓ The agent has some questions:

1. What format should the output be?
   Suggestions: JSON, Plain text, Markdown
   Your answer: _

2. Should the system handle multiple languages?
   Your answer: _
```

### Шаг 3: Подтверждение или уточнение

```
✅ Review the analysis above.

Options:
  [c]ontinue - Proceed to YAML generation
  [r]efine   - Add additional instructions
  [q]uit     - Cancel generation

Your choice: _
```

#### Если выбрали refine:

```
📝 Enter your refinements (press Enter twice to finish):
> Add support for batch processing
> Use Anthropic for the main processor
>
```

### Шаг 4: Генерация YAML

```
⚙️ YAML Generation

Generating configuration...
✓ Configuration generated!

📄 Generated Configuration
======================================================================
version: 0.3

types:
  InputData:
    text: string(1..5000)
...
======================================================================

Options:
  [s]ave     - Save to file and exit
  [e]dit     - Request changes to the configuration
  [q]uit     - Cancel without saving

Your choice: _
```

#### Если выбрали edit:

```
📝 What changes would you like? (press Enter twice to finish)
> Change temperature for main_processor to 0.7
> Add a confidence score to the output
>

⚙️ Regenerating with your changes...
✓ Updated configuration ready!
```

### Шаг 5: Сохранение

```
💾 Configuration saved to: generated-config.yaml

✅ Validating configuration...
✓ Valid configuration!
   - 3 types
   - 2 assistants

🚀 Next Steps

1. Review the generated configuration:
   cat generated-config.yaml

2. Generate SDK:
   aiwf sdk -f generated-config.yaml -o ./generated

3. Or start HTTP server:
   aiwf serve -f generated-config.yaml
```

## Механизм уточнений

Агенты могут задавать уточняющие вопросы когда:
- Описание задачи слишком общее
- Неясны требования к входным/выходным данным
- Множество возможных архитектурных решений
- Нужна дополнительная информация о домене

### Типы вопросов

**1. Выбор из вариантов:**
```
Question: What format should the output be?
Suggestions: JSON, Plain text, Markdown, HTML
Your answer: JSON
```

**2. Открытый вопрос:**
```
Question: What is the expected volume of requests per day?
Reason: To choose appropriate model and configuration
Your answer: Around 10,000 requests
```

**3. Да/Нет:**
```
Question: Should the system cache results?
Your answer: yes
```

## Refinement Instructions

На любом этапе можно внести уточнения:

### В анализе задачи:
```
[r]efine - Add additional instructions
> Use Anthropic Claude for all agents
> Add input validation
> Support streaming responses
```

### В генерации YAML:
```
[e]dit - Request changes
> Increase max_tokens to 2000 for main agent
> Change temperature to 0.8 for creative_writer
> Add a new type for error handling
```

Агент применит изменения и регенерирует конфигурацию.

## Архитектура meta-агента

Система использует двухэтапный подход:

### 1. task_analyzer
- **Вход:** Описание задачи (string)
- **Выход:** TaskAnalysis
- **Роль:** Анализ задачи, определение архитектуры

**Возможности:**
- Оценка сложности (1-10)
- Определение количества агентов (1-10)
- Спецификация каждого агента
- Идентификация типов данных
- Генерация уточняющих вопросов

### 2. yaml_generator
- **Вход:** GenerationInput (analysis + refinements)
- **Выход:** GeneratedConfig
- **Роль:** Генерация валидного YAML

**Возможности:**
- Создание типов с правильными ограничениями
- Генерация агентов с промптами
- Добавление thread/dialog конфигурации
- Валидация сгенерированного YAML
- Предложения по улучшению

### Thread контекст

Оба агента работают в одном thread:
- task_analyzer видит исходную задачу
- yaml_generator видит весь контекст (задача + анализ + уточнения)
- При refinement контекст сохраняется

## Best Practices

### 1. Подробное описание задачи

**Хорошо:**
```
Create a sentiment analysis system for product reviews.
Input: review text (up to 5000 chars)
Output: sentiment (positive/negative/neutral), confidence score, key phrases
Should handle English and Spanish
```

**Плохо:**
```
Sentiment analyzer
```

### 2. Указание домена

Добавьте контекст о домене:
- E-commerce
- Healthcare
- Finance
- Education
- Customer service

Это помогает агенту выбрать правильные модели и параметры.

### 3. Уточнение объема

Если известно:
- Объем запросов в день
- Требования к скорости ответа
- Бюджетные ограничения

### 4. Итеративное улучшение

Не бойтесь использовать refinement:
1. Начните с базового описания
2. Посмотрите на анализ
3. Добавьте уточнения
4. Посмотрите на YAML
5. Запросите изменения

## Валидация

После генерации конфигурация автоматически валидируется:

```
✅ Validating configuration...
✓ Valid configuration!
   - 3 types
   - 2 assistants
```

Если есть ошибки:
```
⚠️  Validation warning: assistant 'analyzer': unknown type 'Result'
   The configuration was saved but may need manual fixes.
```

## HTTP Server поддержка

Сгенерированные конфигурации можно сразу использовать с `aiwf serve`:

```bash
# Генерируем конфигурацию
aiwf generate -i -o my-agents.yaml

# Запускаем сервер
aiwf serve -f my-agents.yaml

# Агенты доступны по HTTP
curl -X POST http://127.0.0.1:8080/agent/analyzer \
  -H "Content-Type: application/json" \
  -d '{"text": "Hello world"}'
```

## Примеры задач

### Простые задачи (1 агент)

- Text classification
- Language detection
- Sentiment analysis
- Text summarization
- Translation
- Keyword extraction

### Средние задачи (2-3 агента)

- Content moderation (toxicity + spam + PII)
- Data validation pipeline
- Document analysis with summarization
- Customer support with routing
- Multi-step content generation

### Сложные задачи (3-5 агентов)

- E-commerce recommendation system
- Multi-language support system
- Complex data analysis pipeline
- Interactive tutoring system
- Multi-stage content processing

### Очень сложные задачи (5+ агентов)

- Enterprise workflow automation
- Multi-domain knowledge system
- Advanced decision support system
- Comprehensive data processing pipeline

## Troubleshooting

### "Meta-config not found"

```bash
# Убедитесь что файл существует
ls templates/meta/config_generator.yaml

# Запускайте из корня проекта
cd /path/to/aiwf
aiwf generate -i
```

### "Failed to generate meta SDK"

Проверьте что meta-конфигурация валидна:
```bash
aiwf validate -f templates/meta/config_generator.yaml
```

### "API key not set"

```bash
export OPENAI_API_KEY="sk-..."
# или
aiwf generate -i --api-key "sk-..."
```

### "Invalid YAML generated"

Сгенерированный YAML всегда сохраняется, даже если есть ошибки валидации.
Можно вручную исправить и проверить:

```bash
# Исправьте вручную
vim generated-config.yaml

# Проверьте
aiwf validate -f generated-config.yaml
```

## Roadmap

Планируемые улучшения:

- [ ] Полная интеграция с meta-агентами (сейчас simulation)
- [ ] Режим streaming для real-time генерации
- [ ] Поддержка шаблонов задач
- [ ] База знаний с примерами
- [ ] Автоматическое тестирование сгенерированных конфигов
- [ ] Web UI для генерации
- [ ] Экспорт в другие форматы (OpenAPI, AsyncAPI)
- [ ] Version control integration

## См. также

- [Getting Started Guide](./GETTING_STARTED.md) - Основы AIWF
- [Generator Documentation](../generator/README.md) - Система типов
- [Serve Guide](./SERVE_GUIDE.md) - HTTP сервер
- [Templates](../templates/README.md) - Примеры конфигураций
