# Repository Guidelines
Reference these notes when contributing to AI Workflow (AIWF).

## Project Structure & Module Organization
Match `structure.md`: `cmd/aiwf` for CLI, `runtime/go/aiwf` for contracts, `providers/{openai,anthropic,local}` for model clients, `generator/{backend-go,backend-ts,backend-py}` for renderers, shared YAML assets in `presets/` and `templates/`, examples and golden SDKs in `examples/`.

## Communication Guidelines
Отвечайте пользователю **только на русском языке**, сохраняя техническую точность и краткость.

## Delivery Roadmap
The canonical checklist lives in `ROADMAP.md`; keep its boxes in sync with this summary.
### Foundation
- [x] Scaffold CLI core in `cmd/aiwf` with flags.
- [x] Define baseline interfaces and table tests in `runtime/go/aiwf`.
- [x] Sync `go.mod` naming and update the charter in `sdk.md`.

### Runtime & Providers
- [x] Ship OpenAI provider с системным промптом (instructions) и заглушкой стриминга.
- [x] Add Anthropic provider с поддержкой `system`-сообщений и shared retry.
- [x] Add local OpenAI-compatible provider-обёртку.
- [x] Provide filesystem artifact store с TTL и sweep, плюс S3 store для удалённых артефактов.

Смотри `docs/providers.md` для аргументов в пользу собственных HTTP-клиентов и маршрутизации системных промптов.

### Generator & Templates
- [x] Build the YAML→IR parser с генерацией промежуточной IR.
- [x] Render Go backend сервис и положи снепшоты в `generator/backend-go/testdata`.
- [ ] Port TS/Python renderers, share helpers, version presets/templates.

### Developer Experience
- [ ] Publish reference workflows and synced SDKs within `examples/`.
- [ ] Polish `aiwf validate|sdk|dry-run` UX text and docs.
- [ ] Wire CI to run `go test ./...` and `aiwf validate`; log schema notes.

## Build, Test, and Development Commands
- `go fmt ./...`
- `go test ./...`
- `go run ./cmd/aiwf --help`
- `aiwf validate presets/workflows/day-planner.yaml`
- `aiwf sdk --file presets/workflows/day-planner.yaml --out ./_generated --package dayplanner`

## Coding Style & Naming Conventions
Keep Go files `gofmt`-clean, exports `PascalCase`, privates `camelCase`, packages lowercase. YAML assistants, workflows, and steps stay `lower_snake_case`; prompts and schema filenames use kebab or snake (e.g., `generate-plan.md`, `daily_plan.json`).

## Testing Guidelines
Colocate table-driven tests with the code, use fake clients, rerun `go test ./...` plus targeted `aiwf validate` before review, and refresh generator golden files in `testdata/` when templates move.

## Commit & Pull Request Guidelines
Follow Conventional Commits, list touched modules, flag CLI-visible changes, and link context in `structure.md` or `sdk.md`; attach screenshots or transcripts for UX work and split refactors from features.

## Agent & Workflow Tips
Update schemas in `presets/assistants`, regenerate workflows, refresh `yaml.md` snippets, keep prompts/templates aligned, and stub new providers so optional integrations compile.

## `/develop` Command Workflow
`/develop` reads the roadmap, selects the next unchecked task by phase, prints the target, and drives the implementation. It finishes with `go fmt ./...`, `go test ./...`, any cited `aiwf` commands, stages files, creates a Conventional Commit (e.g., `feat: add openai provider client`), and either checks the task off or reports failures.
