# AIWF - AI Agent Framework

AIWF — фреймворк для построения типобезопасных AI-агентов с автоматической генерацией SDK.

Разработайте свой агент один раз на простом YAML, получите типобезопасный Go SDK и работайте с любыми LLM провайдерами.

## Возможности

- **Генерация типобезопасного SDK** из YAML-спецификаций
- **Упрощённая система типов** с ограничениями (`string(1..100)`, `enum(...)` и т.д.)
- **Мульти-агенты** со структурированными входами/выходами
- **3 встроенных провайдера**: OpenAI, Grok (xAI), Anthropic (Claude)
- **Конфигурируемые параметры**: max_tokens, temperature для каждого агента
- **Опциональный вывод**: простой `string` или структурированный JSON
- **Управление тредами** для многораундных диалогов
- **Диалоговый режим** для интерактивных агентов

## 🚀 Быстрый старт

⏱️ **Если у вас есть 10 минут** → Следите [`docs/GETTING_STARTED.md`](./docs/GETTING_STARTED.md)

### Установка

```bash
go install ./cmd/aiwf
```

### Использование CLI

```bash
# Валидация YAML-конфигурации
aiwf validate -f config.yaml

# Генерация SDK
aiwf sdk -f config.yaml -o ./generated
```

### Пример YAML-конфигурации

```yaml
version: 0.3

types:
  UserRequest:
    text: string(1..1000)
    language: enum(en, es, fr, de)

  Translation:
    text: string
    confidence: number(0..1)

assistants:
  translator:
    use: openai                    # Провайдер (openai, grok, anthropic)
    model: gpt-4o-mini
    system_prompt: Переведи текст на указанный язык
    input_type: UserRequest
    output_type: Translation
    max_tokens: 1000               # Опционально
    temperature: 0.3               # Опционально
```

## Структура проекта

- **`cmd/aiwf`** - CLI-инструмент для валидации и генерации SDK
- **`generator/`** - Движок генерации кода ([подробнее](./generator/README.md))
- **`runtime/`** - Runtime-библиотеки для разных языков ([подробнее](./runtime/README.md))
- **`providers/`** - Реализации LLM-провайдеров ([подробнее](./providers/README.md))
- **`templates/`** - Готовые шаблоны агентов

## Документация

- [Документация генератора](./generator/README.md) - YAML-спецификация и система типов
- [Документация runtime](./runtime/README.md) - Runtime-контракты и интерфейсы
- [Документация провайдеров](./providers/README.md) - Руководство по реализации провайдеров
- [Шаблоны](./templates/README.md) - Примеры конфигураций

## Полные примеры

Смотрите директорию `examples/` для полностью работающих примеров:

- **DataAnalyst** - Структурированный анализ данных (сложные типы)
- **CreativeWriter** - Генерация текста (простой строковый вывод)
- **CustomerSupport** - Многораундный диалог с контекстом (threads)

Примеры используют OpenAI и Grok провайдеры - см. `.env.example` для настройки.

## Система типов

AIWF использует упрощённую нотацию типов:

### Базовые типы
- `string` - Строка
- `int` - Целое число
- `number` - Число с плавающей точкой
- `bool` - Булево значение
- `any` - Любой тип
- `datetime`, `date`, `uuid` - Специальные типы

### Ограничения
- `string(1..100)` - Ограничение длины строки
- `int(0..100)` - Числовой диапазон
- `enum(value1, value2)` - Перечисление
- `Type[]` - Массив типов
- `map(string, any)` - Словарь
- `$Reference` - Ссылка на другой тип

## Лицензия

MIT
