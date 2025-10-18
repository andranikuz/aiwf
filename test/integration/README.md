# AIWF Integration Tests

Набор интеграционных тестов для проверки работы сгенерированных SDK с реальным OpenAI API.

## Структура

```
integration/
├── README.md                              # Этот файл
├── setup_test.go                          # Инициализация и загрузка конфигурации
├── data_extractor_integration_test.go     # Тесты для data_extractor агента
├── code_analyzer_integration_test.go      # Тесты для code_analyzer агента
├── translator_integration_test.go         # Тесты для translator агента
├── customer_support_integration_test.go   # Тесты для customer_support диалога
├── blog_pipeline_integration_test.go      # Тесты для blog_pipeline воркфлоу
└── generated/                              # Сгенерированные SDK
    ├── data_extractor/
    ├── code_analyzer/
    ├── translator/
    ├── customer_support/
    └── blog_pipeline/
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
go test -v ./test/integration
```

### Запуск конкретного теста

```bash
# Data Extractor тесты
go test -v -run TestDataExtractor ./test/integration

# Code Analyzer тесты
go test -v -run TestCodeAnalyzer ./test/integration

# Translator тесты
go test -v -run TestTranslator ./test/integration

# Customer Support тесты
go test -v -run TestCustomerSupport ./test/integration

# Blog Pipeline тесты
go test -v -run TestBlogPipeline ./test/integration
```

### Запуск конкретного подтеста

```bash
# Только основной тест Data Extractor
go test -v -run TestDataExtractor_Integration ./test/integration

# Тест с несколькими режимами
go test -v -run TestDataExtractor_MultipleExtractionModes ./test/integration
```

### Запуск с временным лимитом

```bash
go test -v -timeout 5m ./test/integration
```

### Запуск с подробным логированием

```bash
go test -v -run TestDataExtractor -args -test.v ./test/integration
```

## Описание тестов

### data_extractor_integration_test.go

Тесты для агента извлечения структурированной информации из текста.

- **TestDataExtractor_Integration**: Базовый тест извлечения сущностей
- **TestDataExtractor_MultipleExtractionModes**: Тест разных режимов работы (entities, relationships, full)
- **TestDataExtractor_EdgeCases**: Граничные случаи (пустые строки, сложный текст)

### code_analyzer_integration_test.go

Тесты для агента анализа кода на качество и безопасность.

- **TestCodeAnalyzer_Integration**: Базовый анализ кода на Go
- **TestCodeAnalyzer_MultipleLanguages**: Анализ кода на разных языках (Python, JavaScript, Rust)
- **TestCodeAnalyzer_ComplexCode**: Анализ реалистичного кода с множественными проблемами

### translator_integration_test.go

Тесты для агента перевода текста.

- **TestTranslator_Integration**: Базовый перевод английского текста
- **TestTranslator_MultipleDomains**: Перевод в разных доменах (technical, legal, medical)
- **TestTranslator_LongForm**: Перевод многострочного текста
- **TestTranslator_EdgeCases**: Граничные случаи (код, URLs, спецсимволы)

### customer_support_integration_test.go

Тесты для диалогового агента поддержки клиентов.

- **TestCustomerSupport_Integration**: Базовое взаимодействие с клиентом
- **TestCustomerSupport_ComplexDialog**: Многошаговый диалог
- **TestCustomerSupport_DifferentSubscriptions**: Проверка обработки разных уровней подписки
- **TestCustomerSupport_WithAttachments**: Обработка сообщений с вложениями

### blog_pipeline_integration_test.go

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
- Увеличьте timeout: `go test -timeout 2m ./test/integration`

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
go test -v -parallel 1 ./test/integration
```

## Регенерация SDK

Если вы обновите YAML спецификации в `templates/`, нужно пересоздать SDK:

```bash
# Сгенерировать SDK для data_extractor
go run ./cmd/aiwf sdk -f templates/assistant/data_extractor.yaml -o test/integration/generated/data_extractor --package data_extractor_sdk

# Сгенерировать SDK для code_analyzer
go run ./cmd/aiwf sdk -f templates/assistant/code_analyzer.yaml -o test/integration/generated/code_analyzer --package code_analyzer_sdk

# Сгенерировать SDK для translator
go run ./cmd/aiwf sdk -f templates/assistant/translator.yaml -o test/integration/generated/translator --package translator_sdk

# Сгенерировать SDK для customer_support
go run ./cmd/aiwf sdk -f templates/dialog/customer_support.yaml -o test/integration/generated/customer_support --package customer_support_sdk

# Сгенерировать SDK для blog_pipeline
go run ./cmd/aiwf sdk -f templates/workflow/blog_pipeline.yaml -o test/integration/generated/blog_pipeline --package blog_pipeline_sdk
```

## Примечания

- Тесты пропускаются автоматически если OPENAI_API_KEY не установлен
- Каждый тест имеет timeout 30 секунд (можно изменить в коде)
- Используется OpenAI API, поэтому каждый тест потребляет токены и средства с аккаунта
- Логирование использует стандартный Go testing логер
