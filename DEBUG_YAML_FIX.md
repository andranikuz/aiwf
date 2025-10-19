# Исправление debug/sdk.yaml

## Что было не так

Оригинальный файл `debug/sdk.yaml` содержал **JSON Schema** (стандарт JSON Schema), но парсер AIWF ожидает **AIWF Type expressions** — собственный формат AIWF для описания типов данных.

### Пример неправильного формата (JSON Schema):
```yaml
types:
  DraftsInput:
    $id: aiwf://bookforge/DraftsInput
    type: object
    required:
      - premise_json
      - characters_json
    properties:
      premise_json:
        type: string
        minLength: 50
        description: JSON BookPremise в виде строки
      characters_json:
        type: string
        minLength: 50
        description: JSON BookCharacters в виде строки
```

**Ошибка парсера:**
```
failed to parse field premise_json: failed to parse field minLength:
unexpected field type for minLength: int
```

## Правильный формат (AIWF Type Expressions)

AIWF использует свой компактный формат для описания типов:

```yaml
types:
  DraftsInput:
    premise_json: string(50..50000)          # string от 50 до 50000 символов
    characters_json: string(50..50000)
    continuity_keys: string[](max:120)       # array строк, максимум 120 элементов
    chapter_id: string
    scenes_json: string(50..50000)
    style_guidelines: string[](max:10)       # array строк, максимум 10 элементов
    writing_constraints: string[](max:10)
```

## Основные синтаксические правила AIWF

### Простые типы:
- `string` — строка без ограничений
- `int` — целое число
- `bool` — логическое значение
- `datetime` — дата-время
- `uuid` — UUID идентификатор

### Типы со строковыми ограничениями:
- `string(1..100)` — от 1 до 100 символов
- `string(email)` — email формат
- `string(url)` — URL формат
- `string(phone)` — телефонный номер

### Массивы:
- `string[]` — массив строк без ограничений
- `string[](max:10)` — максимум 10 элементов
- `string[](min:1, max:5)` — от 1 до 5 элементов
- `$TypeName[]` — массив типов (со ссылкой)

### Перечисления:
- `enum(value1, value2, value3)` — перечисление значений

### Ссылки на другие типы:
- `$TypeName` — ссылка на другой определенный тип
- `$TypeName[](max:10)` — массив другого типа

## OpenAI Strict JSON Schema Requirements

### Важное ограничение для OpenAI Responses API

При использовании OpenAI Responses API с `response_format`, все поля в `properties` должны быть либо:

1. **Обязательными** - включены в `required` array
2. **Не существующими** - полностью удалены из типа

**НЕЛЬЗЯ** использовать опциональные поля (например, `field?: type`).

### ❌ Неправильно:

```yaml
BookPremise:
  title: string(1..120)
  tone?: string(1..80)          # ❌ OpenAI не поддерживает опциональные поля
  setting?: string(1..300)
```

**Ошибка OpenAI:**
```
Invalid schema for response_format: 'required' is required to be supplied
and to be an array including every key in properties. Missing 'tone'.
```

### ✅ Правильно:

**Вариант 1:** Сделать все поля обязательными

```yaml
BookPremise:
  title: string(1..120)
  tone: string(1..80)              # ✅ Обязательное поле
  setting: string(1..300)
```

**Вариант 2:** Удалить опциональные поля совсем

```yaml
BookPremise:
  title: string(1..120)
  genre: string(1..60)
  # tone, setting, и другие опциональные поля удалены
```

### Как исправить свой YAML:

1. **Найти все `?` маркеры:**
   ```bash
   grep "?" your_file.yaml
   ```

2. **Удалить `?` для обязательных полей:**
   ```bash
   sed -i 's/?\(:\)/\1/g' your_file.yaml
   ```

3. **Или удалить полностью опциональные поля:**
   ```yaml
   field_name: string  # Удалить эту строку если опционально
   ```

4. **Валидировать:**
   ```bash
   go run ./cmd/aiwf validate -f your_file.yaml
   ```

## Пример правильного файла

Исправленный файл `debug/sdk.yaml` теперь содержит:

```yaml
version: 0.3-proposal

types:
  # Вспомогательные типы
  Chapter:
    id: string
    number: int
    title: string(1..120)
    summary: string(20..400)
    goal: string(5..160)

  # Основной тип с вложенными структурами
  BookOutline:
    version: string(1..32)
    accepted: bool
    chapters: $Chapter[](min:10, max:15)      # Массив Chapter от 10 до 15
    continuity_keys: string[](max:40)

  # Входные типы для ассистентов
  PremiseInput:
    idea: string(20..800)
    title: string(1..120)
    genre: string(1..60)
    subgenre: string(1..60)
    tone: string(1..80)
    target_audience: string(1..120)
    pov: string(1..80)
    setting_notes: string(1..400)
    inspiration: string[](max:6)               # Массив до 6 элементов
    constraints: string[](max:10)
    must_avoid: string[](max:8)
```

## Команды для работы

### Валидировать YAML файл:
```bash
go run ./cmd/aiwf validate -f debug/sdk.yaml
```

### Сгенерировать SDK из YAML:
```bash
go run ./cmd/aiwf sdk \
  -f debug/sdk.yaml \
  -o debug/generated \
  --package bookforge
```

### Локальный путь к CLI:
```bash
go run ./cmd/aiwf sdk --help
```

## Результат

После исправления:
- ✅ YAML валидирует успешно
- ✅ SDK генерируется в `debug/generated/`
- ✅ Создаются типы Go: `types.go`, `agents.go`, `service.go`
- ✅ Все типы с правильными JSON tags и структурами
- ✅ Совместимо с OpenAI Responses API

## Дополнительные ресурсы

- Примеры: `templates/dialog/customer_support.yaml`
- Документация: `WORKFLOW_ENGINE_GUIDE.md`
- Type expressions: `generator/core/type_parser.go`
- OpenAI API: https://platform.openai.com/docs/guides/structured-outputs
