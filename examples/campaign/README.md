# Launch Campaign Workflow Example

Этот пример показывает, как построить пошаговый план запуска продукта с помощью AIWF: от исследовательских инсайтов до таймлайна и управления рисками.

## Файлы

- `workflow.yaml` — декларативное описание ассистентов и воркфлоу `campaign_launch`.
- `schemas/` — JSON Schema для входов/выходов каждого шага.
- `sdk/` — сгенерированный Go SDK (перегенерируйте командой ниже).
- `main.go` — CLI, который прогоняет воркфлоу, используя OpenAI Responses API.

## Регенерация SDK

```bash
go run ./cmd/aiwf sdk \
  --file examples/campaign/workflow.yaml \
  --out examples/campaign/sdk \
  --package campaign
```

## Запуск примера

```bash
export OPENAI_API_KEY=... # Responses API ключ
cd examples/campaign
go run .
```
