# AIWF Integration Tests

Набор интеграционных тестов для проверки работы сгенерированных SDK с реальным OpenAI API.

Тесты организованы в 3 категории:
- **assistants** - простые одношаговые ассистенты
- **dialogs** - многошаговые диалоговые системы
- **workflows** - сложные многошаговые воркфлоу

## Структура

```
test/integration/
├── assistants/                          # Простые ассистенты
│   ├── setup_test.go                   # Инициализация и загрузка конфигурации
│   ├── data_extractor_integration_test.go
│   ├── code_analyzer_integration_test.go
│   ├── translator_integration_test.go
│   └── generated/                       # Сгенерированные SDK
│       ├── data_extractor/
│       ├── code_analyzer/
│       └── translator/
│
├── dialogs/                             # Диалоговые системы
│   ├── setup_test.go
│   ├── customer_support_integration_test.go
│   └── generated/
│       └── customer_support/
│
├── workflows/                           # Многошаговые воркфлоу
│   ├── setup_test.go
│   ├── blog_pipeline_integration_test.go
│   └── generated/
│       └── blog_pipeline/
│
└── README.md                            # Этот файл
```

## Требования

- Go 1.24 или выше
- Действительный OpenAI API ключ

## Подготовка

### 1. Создание файла `.env`

Скопируйте `.env.example` в корне проекта в `.env` и заполните свой OpenAI API ключ:

```bash
cp .env.example .env
```

Отредактируйте `.env`:

```env
OPENAI_API_KEY=sk-your-actual-key-here
ANTHROPIC_API_KEY=sk-ant-your-key-if-needed
AWS_REGION=us-east-1
TEST_TIMEOUT_SECONDS=60
TEST_MAX_RETRIES=3
```

### 2. Загрузка переменных окружения

Перед запуском тестов установите переменные окружения:

**На macOS/Linux:**
```bash
export $(cat .env | grep -v '^#' | xargs)
```

**Или передайте API ключ напрямую:**
```bash
export OPENAI_API_KEY=sk-your-actual-key-here
```

## Запуск тестов

### Запуск всех тестов

```bash
go test -v ./test/integration/...
```

### Запуск тестов по категориям

**Assistants (простые ассистенты):**
```bash
go test -v ./test/integration/assistants
```

**Dialogs (диалоговые системы):**
```bash
go test -v ./test/integration/dialogs
```

**Workflows (многошаговые воркфлоу):**
```bash
go test -v ./test/integration/workflows
```

### Запуск конкретного теста

```bash
# Data Extractor тесты
go test -v -run TestDataExtractor ./test/integration/assistants

# Code Analyzer тесты
go test -v -run TestCodeAnalyzer ./test/integration/assistants

# Translator тесты
go test -v -run TestTranslator ./test/integration/assistants

# Customer Support тесты
go test -v -run TestCustomerSupport ./test/integration/dialogs

# Blog Pipeline тесты
go test -v -run TestBlogPipeline ./test/integration/workflows
```

### Запуск конкретного подтеста

```bash
# Только основной тест Data Extractor
go test -v -run TestDataExtractor_Integration ./test/integration/assistants

# Тест с несколькими режимами
go test -v -run TestDataExtractor_MultipleExtractionModes ./test/integration/assistants
```

### Запуск с временным лимитом

```bash
go test -v -timeout 5m ./test/integration/...
```

### Запуск с подробным логированием

```bash
go test -v -run TestDataExtractor -args -test.v ./test/integration/assistants
```

## Описание тестов

### assistants/data_extractor_integration_test.go

Тесты для агента извлечения структурированной информации из текста.

- **TestDataExtractor_Integration**: Базовый тест извлечения сущностей
- **TestDataExtractor_MultipleExtractionModes**: Тест разных режимов работы (entities, relationships, full)
- **TestDataExtractor_EdgeCases**: Граничные случаи (пустые строки, сложный текст)

### assistants/code_analyzer_integration_test.go

Тесты для агента анализа кода на качество и безопасность.

- **TestCodeAnalyzer_Integration**: Базовый анализ кода на Go
- **TestCodeAnalyzer_MultipleLanguages**: Анализ кода на разных языках (Python, JavaScript, Rust)
- **TestCodeAnalyzer_ComplexCode**: Анализ реалистичного кода с множественными проблемами

### assistants/translator_integration_test.go

Тесты для агента перевода текста.

- **TestTranslator_Integration**: Базовый перевод английского текста
- **TestTranslator_MultipleDomains**: Перевод в разных доменах (technical, legal, medical)
- **TestTranslator_LongForm**: Перевод многострочного текста
- **TestTranslator_EdgeCases**: Граничные случаи (код, URLs, спецсимволы)

### dialogs/customer_support_integration_test.go

Тесты для диалогового агента поддержки клиентов.

- **TestCustomerSupport_Integration**: Базовое взаимодействие с клиентом
- **TestCustomerSupport_ComplexDialog**: Многошаговый диалог
- **TestCustomerSupport_DifferentSubscriptions**: Проверка обработки разных уровней подписки
- **TestCustomerSupport_WithAttachments**: Обработка сообщений с вложениями

### workflows/blog_pipeline_integration_test.go

Тесты для воркфлоу создания блог-поста.

- **TestBlogPipeline_Integration**: Инициализация пайплайна
- **TestBlogPipeline_ResearchAgent**: Агент исследования
- **TestBlogPipeline_OutlineAgent**: Агент создания плана
- **TestBlogPipeline_WriterAgent**: Агент написания контента
- **TestBlogPipeline_EditorAgent**: Агент редактирования
- **TestBlogPipeline_AllAgents**: Проверка доступности всех агентов

## Метрики и отслеживание

Каждый тест выводит:

- ✓ Успешное выполнение операции
- Токены использования (InputTokens, OutputTokens)
- Время выполнения (через context timeout)
- Результаты анализа/обработки

Пример вывода:

```
✓ Translation completed
  Original: The quick brown fox jumps over the lazy dog
  Translated: El rápido zorro marrón salta sobre el perro perezoso
  Confidence: 0.95
  Alternative translations (2):
    [0] El ágil zorro pardo salta sobre el perro indolente (preference: 0.92)
  Tokens (in/out): 45/38
```

## Решение проблем

### Ошибка: OPENAI_API_KEY not set

```bash
# Убедитесь, что переменная окружения установлена
echo $OPENAI_API_KEY

# Если пусто, экспортируйте API ключ
export OPENAI_API_KEY=sk-your-key
```

### Ошибка: connection timeout

- Проверьте интернет соединение
- Проверьте актуальность API ключа
- Увеличьте timeout: `go test -timeout 2m ./test/integration/...`

### Ошибка: invalid API key

```bash
# Проверьте что ключ корректный
echo $OPENAI_API_KEY
# Должен начинаться с "sk-"

# Обновите в .env и переэкспортируйте
export OPENAI_API_KEY=sk-new-key
```

### Ошибка: rate limit exceeded

OpenAI API имеет ограничения по частоте запросов. Если вы получаете ошибки 429:

```bash
# Используйте -parallel flag для ограничения одновременных тестов
go test -v -parallel 1 ./test/integration/...
```

## Регенерация SDK

Если вы обновите YAML спецификации в `templates/`, нужно пересоздать SDK:

```bash
# Assistants
go run ./cmd/aiwf sdk -f templates/assistant/data_extractor.yaml -o test/integration/assistants/generated/data_extractor --package data_extractor_sdk
go run ./cmd/aiwf sdk -f templates/assistant/code_analyzer.yaml -o test/integration/assistants/generated/code_analyzer --package code_analyzer_sdk
go run ./cmd/aiwf sdk -f templates/assistant/translator.yaml -o test/integration/assistants/generated/translator --package translator_sdk

# Dialogs
go run ./cmd/aiwf sdk -f templates/dialog/customer_support.yaml -o test/integration/dialogs/generated/customer_support --package customer_support_sdk

# Workflows
go run ./cmd/aiwf sdk -f templates/workflow/blog_pipeline.yaml -o test/integration/workflows/generated/blog_pipeline --package blog_pipeline_sdk
```

## Примечания

- Тесты пропускаются автоматически если OPENAI_API_KEY не установлен
- Каждый тест имеет timeout 30 секунд (можно изменить в коде)
- Используется OpenAI API, поэтому каждый тест потребляет токены и средства с аккаунта
- Логирование использует стандартный Go testing логер
