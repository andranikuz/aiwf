# AI Workflow Framework (AIWF)

AIWF — это фреймворк для описания и генерации SDK под AI-воркфлоу. Конфигурации задаются в YAML, после чего CLI `aiwf` генерирует типизированный код для Go (другие языки в работе).

## Структура проекта

- `cmd/aiwf` — CLI (`aiwf sdk`, `aiwf validate`).
- `generator/` — парсер спецификации и backend’ы генерации (Go/TS/Py).
- `runtime/go/aiwf` — контракты для исполнения воркфлоу (ModelClient, ThreadManager, ArtifactStore и т.д.).
- `providers/` — реализации провайдеров (OpenAI, Anthropic, local stub).
- `examples/` — примеры конфигураций и сгенерированных SDK.
- `docs/` — документация по провайдерам, ТЗ, roadmap.

## Быстрый старт

1. Установите CLI:
   ```bash
   go install ./cmd/aiwf
   ```
2. Сгенерируйте SDK:
   ```bash
   aiwf sdk \
     --file examples/blog/config/sdk.yaml \
     --out examples/blog/sdk \
     --package blog
   ```
3. Ознакомьтесь с `README` конкретного примера (например, `examples/blog/README.md`).

Подробные инструкции по YAML-спецификации и диалоговым воркфлоу смотрите в `docs/dialog-workflows.md`.
