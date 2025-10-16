# YAML-спецификация AIWF

## Обязательные блоки

- `version`: семантическая версия спецификации (например, `0.3`).
- `assistants`: словарь агентов; минимум один.

## Необязательные блоки

- `schema_registry`: корень JSON Schema (если схемы хранятся в `.json`).
- `imports`: список YAML-файлов с секцией `types:` (используйте `$id` вида `aiwf://namespace/Type`).
- `threads`: политики работы с тредами (см. ниже).
- `workflows`: DAG воркфлоу (можно пропустить, если нужен только SDK агентов).

## Раздел `threads`

```yaml
threads:
  default:
    provider: openai_responses
    strategy: append      # append | reset_before_step
    create: true
    close_on_finish: false
    ttl_hours: 24
    metadata:
      project: book
```

## Описание ассистента

```yaml
assistants:
  premise:
    model: gpt-4.1-mini
    system_prompt: |
      Ты — автор идей.
    input_schema_ref: "aiwf://book/PremiseInput"
    output_schema_ref: "aiwf://book/PremiseOutput"
    thread: { use: default, strategy: append }
    dialog: { max_rounds: 3 }
```
- `system_prompt_ref` можно использовать вместо `system_prompt`.
- `thread` определяет политику; `dialog.max_rounds` задаёт лимит уточнений.

## DAG воркфлоу

```yaml
workflows:
  novel:
    thread: { use: default }
    dag:
      - step: premise
        assistant: premise
        approval:
          review:
            prompt: "Проверь"
            properties:
              "logline": { prompt: "≤160" }
          on_reject: { action: continue_dialog }
        next:
          step: outline
          input_binding:
            premise: "{{ .premise }}"
          input_contract_ref: "aiwf://book/OutlineInput"
      - step: outline
        assistant: outline
```

### `approval` действия

- `continue` — перейти к следующему шагу.
- `retry` — повторить в рамках воркфлоу (новый запуск шага).
- `continue_dialog` — повторить в рамках диалога (не переходя к следующему шагу).
- `goto` — перейти к указанному шагу.
- `stop` — прервать воркфлоу.

### `next.input_binding`

Использует Go `text/template`. Доступны переменные `input` (вход воркфлоу) и результаты шагов (`.premise`, `.outline` и др.).

## Импорт типов

```yaml
# common.types.yaml
types:
  Tone:
    $id: "aiwf://common/Tone"
    type: string
    enum: [dark, hopeful, playful]

# book.types.yaml
types:
  PremiseOutput:
    $id: "aiwf://book/PremiseOutput"
    type: object
    required: [logline, themes]
    properties:
      logline: { type: string, minLength: 10 }
      themes:
        type: array
        items: { type: string }
        maxItems: 6
```

## Генерация SDK

```bash
aiwf sdk \
  --file examples/book/sdk.yaml \
  --out examples/book/sdk \
  --package book
```
- При наличии `threads`/`dialog` SDK генерирует `dialog.go`, расширенные `service.go`, `workflows.go` и агенты.
- Без `workflows` создаются только агенты и контракты.

## Валидация

```bash
aiwf validate examples/book/sdk.yaml
```

Проверяется структура YAML, наличие схем по ссылкам, корректность политики тредов, диалоговых настроек и approval.
