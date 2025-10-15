# AI Blog Workflow Example

Этот пример демонстрирует, как описывать и генерировать SDK для блога с помощью AIWF.

## Сценарий

Воркфлоу `novel` состоит из двух ассистентов:
- `premise` формирует логлайн и список тем.
- `outline` строит план глав, используя результат первого шага.

Схемы типов вынесены в YAML с импортами, что позволяет переиспользовать контракты между шагами.

## Важные файлы

- `config/sdk.yaml` — основная спецификация AIWF.
- `config/common.types.yaml` — общие типы (например, `Tone`).
- `config/book.types.yaml` — типы блога (`PremiseOutput`, `OutlineOutput`).
- `example_usage.go` — Go-пример использования сгенерированного SDK.

## Структура YAML-спецификации

```yaml
version: 0.2
imports:
  - as: common
    path: ./common.types.yaml
  - as: book
    path: ./book.types.yaml
assistants:
  premise:
    model: gpt-4.1-mini
    system_prompt: |
      ...
    output_schema_ref: "aiwf://book/PremiseOutput"
  outline:
    model: gpt-4.1
    system_prompt_ref: prompts/outline.md
    input_schema_ref: "aiwf://book/OutlineInput"
    output_schema_ref: "aiwf://book/OutlineOutput"
workflows:
  novel:
    dag:
      - step: premise
        approval: ...
        next:
          step: outline
          input_binding:
            premise: "{{ .Premise }}"
          input_contract_ref: "aiwf://book/OutlineInput"
      - step: outline
```

### Разделы
- `imports` — подключение YAML-файлов с типами (`types:`).
- `assistants` — описание моделей и ссылок на схемы (`system_prompt`, `system_prompt_ref`, `input/output_schema_ref`).
- `workflows` — DAG шагов, правила переходов, approval, привязки входов.

## Генерация Go SDK

```bash
go run ./cmd/aiwf sdk \
  --file examples/blog/config/sdk.yaml \
  --out examples/blog/sdk \
  --package blog
```

## Использование SDK

Пример `example_usage.go` создаёт фейковый `ModelClient`, запускает `service.Workflows().Novel()` и печатает список глав.

```bash
cd examples/blog
go run ./example_usage.go
```

Для реальной интеграции замените `fakeClient` на провайдер (OpenAI/Anthropic) и передайте `aiwf.ModelCall` в SDK сервис.

## Полезно знать

- Enum значения из YAML превращаются в Go-константы (`ToneDark`, `ToneHopeful`, ...).
- Контракты из `imports` генерируются как структуры и alias-ы (`OutlineOutput`, `Tone`).
- Approval блога демонстрирует dot-path правила (`logline`, `themes[]`).
