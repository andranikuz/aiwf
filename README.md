# AIWF - AI Agent Framework

AIWF — фреймворк для построения типобезопасных AI-агентов с автоматической генерацией SDK.

## Возможности

- **Генерация типобезопасного SDK** из YAML-спецификаций
- **Упрощённая система типов** с ограничениями (`string(1..100)`, `enum(...)` и т.д.)
- **Мульти-агенты** со структурированными входами/выходами
- **Абстракция провайдеров** (OpenAI, Claude и др.)
- **Управление тредами** для контекста диалогов
- **Диалоговый режим** для интерактивных агентов

## Быстрый старт

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
types:
  UserRequest:
    text: string(1..1000)
    language: enum(en, es, fr, de)

  Translation:
    text: string
    confidence: number(0..1)

assistants:
  translator:
    model: gpt-4o-mini
    system_prompt: Переведи текст на указанный язык
    input_type: UserRequest
    output_type: Translation
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
