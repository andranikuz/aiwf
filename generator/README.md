# Generator - Генератор SDK

Генератор преобразует YAML-спецификации в типобезопасный код для различных языков программирования.

## Архитектура

```
YAML ──LoadSpec──▶ Spec ──ResolveTypes──▶ BuildIR ──▶ IR ──backend──▶ SDK
```

- `core/spec.go` — структуры YAML-спецификации
- `core/typedef.go` — система типов и парсер выражений
- `core/resolution.go` — резолюция типов и ссылок
- `core/ir.go` — построение промежуточного представления (IR)
- `backend-go/` — генератор Go-кода

## YAML Спецификация

### Структура файла

```yaml
# Определения типов
types:
  TypeName:
    field1: type_expression
    field2: type_expression

# Ассистенты (агенты)
assistants:
  assistant_name:
    model: string           # Модель LLM (gpt-4o, claude-3, etc.)
    system_prompt: string   # Системный промпт
    input_type: TypeName    # Тип входных данных
    output_type: TypeName   # Тип выходных данных
    thread:                 # Опционально: конфигурация треда
      use: thread_name
      strategy: string
    dialog:                 # Опционально: диалоговый режим
      max_rounds: int

# Воркфлоу
workflows:
  workflow_name:
    steps:
      - name: step_name
        assistant: assistant_name
        needs: [previous_step]

# Треды
threads:
  thread_name:
    provider: openai
    strategy: new|continue|append
```

## Система типов

### Базовые типы

- `string` - Строка
- `int` - Целое число
- `number` - Число с плавающей точкой
- `bool` - Булево значение
- `any` - Любой тип
- `datetime`, `date`, `uuid` - Специальные типы

### Выражения типов

#### Строки с ограничениями
```yaml
username: string(1..100)    # Длина от 1 до 100
password: string(10..)       # Минимум 10 символов
bio: string(..500)          # Максимум 500 символов
```

#### Числа с диапазонами
```yaml
age: int(0..150)            # Целое от 0 до 150
confidence: number(0..1)     # Число от 0.0 до 1.0
count: int(0..)             # От 0 без верхней границы
```

#### Перечисления
```yaml
status: enum(active, inactive, pending)
```

#### Массивы
```yaml
tags: string[]              # Массив строк
users: User[]               # Массив объектов
```

#### Словари
```yaml
metadata: map(string, any)   # Словарь с любыми значениями
scores: map(string, number)  # Словарь чисел
```

#### Ссылки на типы
```yaml
author: User                 # Ссылка на другой тип
manager: $User              # Альтернативный синтаксис
```

### Примеры типов

```yaml
types:
  # Простой объект
  User:
    id: uuid
    name: string(1..50)
    email: string
    age: int(0..150)
    role: enum(admin, user, guest)

  # Составной тип
  Article:
    title: string(1..200)
    content: string(10..10000)
    author: User
    tags: string[]
    metadata: map(string, any)
    status: enum(draft, published)
    published_at: datetime
```

## Генерация Go-кода

### Генерируемые файлы

1. **types.go** - Структуры и валидаторы
```go
type User struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age"`
    Role  string `json:"role"`
}

func ValidateUser(u *User) error {
    // Валидация ограничений
}
```

2. **agents.go** - Типизированные агенты
```go
func (a *TranslatorAgent) Run(ctx context.Context, input TranslationRequest) (*TranslationResult, *aiwf.Trace, error) {
    // Реализация
}
```

3. **service.go** - Сервис с агентами
```go
type Service struct {
    Agents *Agents
}
```

## Использование CLI

```bash
# Валидация
aiwf validate -f config.yaml

# Генерация SDK
aiwf sdk -f config.yaml -o ./generated --package myapp
```

## Совместимость с OpenAI API

### JSON Schema Requirements

При использовании сгенерированных типов с OpenAI Responses API (`response_format`):

- ✅ Все поля в `properties` автоматически добавляются в `required` массив
- ✅ Соответствует OpenAI Strict JSON Schema требованиям
- ✅ Генерируется `"additionalProperties": false` для type safety

Это обеспечивает полную совместимость с OpenAI's structured outputs.

## Текущие ограничения

- Валидация ограничений пока не генерируется полностью
- Workflows требуют доработки под новую систему типов
- Опциональные поля (помеченные `?`) требуют либо исключения из типа,
  либо добавления в `required` (для OpenAI совместимости)

## Roadmap

- [ ] Полная генерация валидаторов
- [ ] Поддержка workflows
- [ ] Опциональные поля
- [ ] Python/TypeScript генераторы
