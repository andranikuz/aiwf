# AIWF Delivery Roadmap
Track these milestones sequentially to build out the AI Workflow (AIWF) platform. Tackle sections in order; move to the next phase only after all tasks are checked.

## Foundation
- [x] Choose CLI framework (Cobra vs urfave), document decision, and scaffold the root command in `cmd/aiwf`.
- [x] Implement shared option parsing (config path, env, verbosity) and expose it to subcommands.
- [x] Define interfaces (`ModelClient`, `Workflow`, `Trace`, `Retry`, `ArtifactStore`) in `runtime/go/aiwf` with doc comments referencing `sdk.md`.
- [x] Add table-driven placeholder tests ensuring the interfaces compile against simple fakes.
- [x] Align `go.mod` module path, tidy dependencies, and refresh architectural overview in `sdk.md` to match the new structure.

## Runtime & Providers
- [x] Deliver OpenAI provider: request/response structs, placeholder JSON Schema check, streaming stub, и системный промпт через `instructions`.
- [x] Create Anthropic provider mirroring OpenAI surface, wiring shared retry/backoff utilities и систему сообщений (`system` + `user`).
- [x] Build local provider adapter supporting OpenAI-compatible endpoints (model router + configuration file).
- [x] Implement файловое хранилище с TTL и S3 store для удалённых артефактов.
- [ ] Stub S3 artifact store (interface + integration tests with localstack or fake client).
- [ ] Introduce basic trace/metrics hooks (structured logging, counters for token usage, retry count).

## Generator & Templates
- [x] Implement YAML loader с проверкой JSON Schema.
- [x] Construct intermediate representation (IR) structs capturing assistants, workflows, dag edges, scatter metadata.
- [x] Validate IR consistency (schema existence, dependency cycles) and report actionable errors.
- [x] Render Go backend templates: сервис, агенты, воркфлоу и контракты с golden-файлами.
- [x] Generate Go agents/workflows stubs with ModelClient вызовами и merge trace helper.
- [x] Derive контрактные типы из JSON Schema (простые свойства, массивы, bool/number).
- [x] Map workflow inputs/outputs (сохраняем результаты шагов и передаём их дальше).
- [ ] Подставлять реальные аргументы для шагов (преобразование `prev` и `input` в `stepInput`).
- [ ] Собрать финальный `result` и реализовать полноценный `mergeWorkflowOutput/mergeTraces`.
- [ ] Extract template helper library shared across Go/TS/Python backends.
- [ ] Port renderer to TypeScript with buildable package.json scaffold and golden snapshots.
- [ ] Port renderer to Python with packaging metadata and snapshot fixtures.
- [ ] Version presets/templates, add changelog entries, and define upgrade notes for generated SDKs.

## Developer Experience & Delivery
- [ ] Publish reference workflows in `presets/workflows` and regenerate matching SDKs under `examples/`.
- [ ] Document tutorial-style walkthrough for `aiwf validate`, `aiwf sdk`, and `aiwf dry-run` (include CLI examples).
- [ ] Wire CI to run `go test ./...`, lint Go (`go vet ./...`), and execute `aiwf validate` across presets.
- [ ] Provide contribution templates (PR checklist, issue template) referencing mandatory commands.
- [ ] Draft schema migration notes and automate a changelog snippet per release.

## Release Readiness
- [ ] Finalize versioning strategy (SemVer per module) and document release cadence.
- [ ] Prepare initial release notes highlighting supported providers and generators.
- [ ] Tag `v0.1.0` once core CLI, runtime, and Go generator are stable; publish binaries or install instructions.
