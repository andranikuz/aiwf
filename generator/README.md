# Генератор SDK AIWF

Папка `generator/` содержит пайплайн превращения YAML-спецификации AIWF в типизированный SDK.

## Архитектура

```
YAML ──LoadSpec──▶ Spec ──BuildIR──▶ IR ──backend──▶ SDK
```

- `core/spec.go` — структуры YAML (`Spec`, `Assistants`, `Workflows`, `Threads`).
- `core/loader.go` — `LoadSpec`: чтение YAML, импорты типов (`imports`), валидация `threads`, диалоговых настроек, approval, проверка JSON Schema.
- `core/ir.go` — `BuildIR`: нормализация данных, подготовка реестра типов, тредов, approval для backend’а.
- `backend-go/` — генерация Go-кода (`service.go`, `agents.go`, `workflows.go`, `dialog.go`, `contracts.go`). TS/Py backend’ы в работе.

## YAML-спецификация

Пример набора блоков:

```yaml
version: 0.3
imports:
  - as: common
    path: ./common.types.yaml
threads:
  default:
    provider: openai_responses
    strategy: append
    ttl_hours: 24
assistants:
  premise:
    model: gpt-4.1-mini
    system_prompt: "..."
    output_schema_ref: "aiwf://book/PremiseOutput"
    thread: { use: default }
    dialog: { max_rounds: 3 }
workflows:
  novel:
    thread: { use: default }
    dag:
      - step: premise
        assistant: premise
        approval:
          review:
            prompt: "Проверь целостность"
            properties:
              "logline": { prompt: "≤160" }
          on_reject: { action: continue_dialog }
        next:
          step: outline
          input_binding:
            premise: "{{ .premise }}"
          input_contract_ref: "aiwf://book/OutlineInput"
```

Ключевые блоки:

- `imports` — YAML с `types:` и `$id`, формирующие `aiwf://`-ссылки.
- `threads` — политики тредов (`provider`, `strategy`, `ttl_hours`, `metadata`).
- `assistants` — модели, промпты, ссылки на схемы, диалоговые настройки.
- `workflows` — DAG шагов, `thread`/`dialog`, `approval` (действия: `continue`, `retry`, `continue_dialog`, `goto`, `stop`).

## Генерация Go SDK

```bash
aiwf sdk \
  --file examples/book/sdk.yaml \
  --out examples/book/sdk \
  --package book
```

Результат:

- `service.go` — фабрика `NewService`, методы настройки `WithThreadManager`, `WithDialogDecider`, `WithMaxDialogRounds`.
- `agents.go` — агенты принимают/возвращают `*aiwf.ThreadState`.
- `workflows.go` — пошаговый раннер с диалоговым циклом (вызовы агента → решение `DialogDecider`).
- `dialog.go` — вспомогательные типы и опции выполнения.
- `contracts.go` — структуры и enum-константы из схем.

## Тесты

```bash
unset GOROOT && export GOTOOLCHAIN=local
UPDATE_GOLDEN=1 go test ./generator/backend-go
```

- `generator/backend-go/generator_test.go` — снапшотные тесты (папка `testdata/`).
- Для IR/loader — `go test ./generator/core`.

## Поддержка диалогов

IR хранит `Threads` и `Dialog` настройки, Go backend использует их для генерации run-опций и thread-binding. README по диалогам и YAML см. `docs/dialog-workflows.md`.
