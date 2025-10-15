# Генератор SDK AIWF

Директория `generator/` содержит общий пайплайн перехода от YAML-конфигурации воркфлоу к структурированному SDK.

## Архитектура

```
YAML (spec) ──LoadSpec──▶ Spec ──BuildIR──▶ IR ──backend──▶ SDK (Go/TS/Py)
```

- `core/spec.go` — структура данных YAML (`Spec`, `AssistantSpec`, `WorkflowSpec`).
- `core/loader.go` — функция `LoadSpec(path)`:
  - читает YAML;
  - резолвит `imports` (YAML с секцией `types:`);
  - подгружает JSON Schema по `aiwf://` и валидирует их;
  - возвращает `Spec` + накопленный реестр типов.
- `core/ir.go` — `BuildIR(spec)` нормализует данные, чтобы backend не зависел от оригинального YAML.
- `backend-*/` — рендер файлов (Go реализован полностью, TS/Py в работе).

## YAML-спецификация

Минимальная структура:

```yaml
version: 0.2
imports:
  - as: common
    path: ./common.types.yaml
assistants:
  premise:
    model: gpt-4.1-mini
    system_prompt: |
      ...
    output_schema_ref: "aiwf://book/PremiseOutput"
workflows:
  novel:
    dag:
      - step: premise
        assistant: premise
```

### Корневые поля

| Поле                | Описание                                                                 |
|---------------------|---------------------------------------------------------------------------|
| `version`           | Семантическая версия спецификации (для будущей эволюции).                 |
| `imports`           | Массив YAML-файлов с типами (`types:`). Каждый тип должен иметь `$id`.     |
| `assistants`        | Набор ассистентов (LLM-агентов). Обязательный блок.                        |
| `workflows`         | (Опционально) описание DAG воркфлоу. Можно пропустить для генерации только агентов. |
| `schema_registry`   | (Опционально) путь к директории с JSON Schema; если используете `imports`, можно опустить. |

### Описание ассистента

```yaml
assistants:
  outline:
    use: openai_responses      # alias провайдера (необязательно)
    model: gpt-4.1             # имя модели
    system_prompt: |
      Ты — редактор...
    system_prompt_ref: prompts/outline.md  # внешний файл (если указан, имеет приоритет)
    input_schema_ref: "aiwf://book/OutlineInput"   # может быть пустым
    output_schema_ref: "aiwf://book/OutlineOutput" # обязательно
    depends_on: [premise]  # выражает requirement между ассистентами
```

- `system_prompt` vs `system_prompt_ref`: если заданы оба, loader подставит содержимое файла.
- `input_schema_ref` / `output_schema_ref` указывают либо на `aiwf://` тип из `imports`, либо относительный путь в `schema_registry`.
- Схема должна валидироваться как JSON Schema Draft 2020-12 (через `gojsonschema`).

### YAML c типами

```yaml
# common.types.yaml
types:
  Tone:
    $id: "aiwf://common/Tone"
    type: string
    enum: [dark, hopeful, playful]

  DateRange:
    $id: "aiwf://common/DateRange"
    type: object
    required: [start, end]
    properties:
      start: { type: string, format: date }
      end: { type: string, format: date }
```

Правила:
- Раздел `types` содержит словарь схем.
- `$id` обязателен и служит ключом для `aiwf://` ссылок.
- Допустимо использовать `$ref` внутри схем — они резолвятся из реестра импортов.

### Workflow DAG

```yaml
workflows:
  novel:
    description: "Сценарий написания романа"
    dag:
      - step: premise
        assistant: premise
        approval:
          review:
            strictness: 7
            prompt: "..."
            properties:
              "logline": { prompt: "≤160 символов" }
          on_approve: { action: goto, step: outline }
          on_reject: { action: retry }
        next:
          step: outline
          input_binding:
            premise: "{{ .Premise }}"
          input_contract_ref: "aiwf://book/OutlineInput"
      - step: outline
        assistant: outline
```

Ключевые элементы:
- `dag` — упорядоченный список шагов.
- `input_binding` — шаблон формирования входа следующего шага (поддерживает `text/template` синтаксис).
- `input_contract_ref` — ссылка на тип для валидации входа.
- `approval` — правила ревью (указание dot-path, `on_approve/on_reject`).
- Если раздел `workflows` отсутствует, можно генерировать только агентов/контракты.

## Генерация SDK

Команда CLI находится в `cmd/aiwf/sdk`.

```bash
go run ./cmd/aiwf sdk \
  --file examples/blog/config/sdk.yaml \
  --out examples/blog/sdk \
  --package blog
```

В итоге появятся файлы:
- `service.go` — фабрика `NewService` с доступом к агентам/воркфлоу;
- `agents.go` — методы для вызова ассистентов;
- `contracts.go` — структуры, alias-ы, enum-константы, полученные из YAML;
- `workflows.go` — раннеры DAG.

## Тесты

```bash
go test ./generator/...
```

Обратите внимание на совместимость версии toolchain (`go env GOVERSION`). Для обновления снепшотов Go backend используйте `UPDATE_GOLDEN=1 go test ./generator/backend-go`.

## Настройка Backend-ов

- Go backend: шаблоны в `backend-go/templates/*.tmpl`, вспомогательные функции в `schema.go`, `generator.go`.
- TS/Py — после дополнения IR будут использовать тот же `BuildIR`.

