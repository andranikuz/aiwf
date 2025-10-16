# Техническое задание: Диалоговые воркфлоу AIWF

## 1. Цель
- Добавить поддержку многошаговых (диалоговых) взаимодействий внутри каждого шага воркфлоу.
- Позволить уточнять требования через чат и проводить approval с возможностью доработок.
- Сохранить обратную совместимость со сценариями «один шаг — один вызов».

## 2. Требования к YAML-спецификации
### 2.1 Новый раздел `threads`
```yaml
threads:
  default:
    provider: openai_responses
    create: true
    close_on_finish: false
    strategy: append   # append | reset_before_step
    ttl_hours: 24
    metadata:
      project: "novel"
```
- Описывает стратегии хранения истории (thread policy).
- Параметры: `provider`, `strategy`, `create`, `close_on_finish`, `ttl_hours`, `metadata`.

### 2.2 Ассистенты и шаги
- Добавить опциональный блок `thread` в описании ассистента и шага:
```yaml
assistants:
  outline:
    thread:
      use: default
      strategy: append
    dialog:
      max_rounds: 5
```
- Параметр `dialog.max_rounds` задаёт лимит уточнений до перехода к следующему шагу.

### 2.3 Approval
- Расширить структуру `approval`:
```yaml
approval:
  review:
    prompt: "..."
    properties:
      "logline": { prompt: "≤160" }
  feedback_template: "Обнови поле {{ .Path }} и пришли JSON ещё раз."
  require_for_retry: true
  max_retries: 3
  on_approve: { action: goto, step: outline }
  on_reject: { action: continue_dialog }
```
- Ввести новое действие `continue_dialog` для повторного запроса модификации внутри текущего шага.
- Поле `feedback_template` формирует сообщение ассистенту при доработке.
- Флаг `auto_continue_on_approve` (по умолчанию true) позволяет оставаться в диалоге после approve.

### 2.4 Workflows
- Разрешить отсутствие блока `workflows` (для генерации только агентов).
- При наличии DAG шагов — поддерживать диалоговый цикл на каждом этапе.

## 3. IR и генератор
- `core.Spec`: хранить `threads`, `dialog` настройки, расширенный `approval`.
- `LoadSpec`: провалидировать thread-политики, пре-создавать пустой `workflows` при отсутствии.
- `BuildIR`: перенести `threads`, `dialog.max_rounds`, расширенный approval.
- `backend-go`:
  - `service.go`: генерировать интерфейсы `ThreadManager`, `ThreadSession`; методы `Workflows` создавать только при наличии DAG.
  - `agents.go`: `Run` должен принимать/возвращать `*ThreadState`.
  - `workflows.go`: реализовать цикл диалога (run → approval → continue_dialog/goto/stop).
  - `contracts.go`: без изменений.
  - Добавить вспомогательные типы (`DialogStepResult`, `DialogAction`).

## 4. Рантайм
- Расширить `aiwf.ModelClient` или ввести отдельный интерфейс для thread-операций (`CreateThread`, `AppendMessage`, `CloseThread`).
- Добавить `ThreadState` (thread_id, history snapshot).
- Расширить `Trace`: `DialogRounds`, цепочка трейсов шагов.
- Разработать интерфейс `ThreadStore` (in-memory по умолчанию, готовность для внешних стор).

## 5. Провайдеры
- `providers/openai`:
  - Поддержка `thread_id`, хранение истории.
  - Методы для создания/закрытия тредов.
  - Поддержать `continue_dialog` — переслать feedback-сообщение в тред.
- `providers/anthropic`, `local`: stub-интерфейс с предупреждением (не поддерживает диалоговые треды, работает single-shot).

## 6. CLI
- `aiwf sdk`: обновить шаблоны, поддерживающие диалоги.
- `aiwf validate`: проверки thread-политик, `dialog.max_rounds >= 1`, корректность `feedback_template`.
- Логи генерации/валидации дополнять информацией о включенных диалогах.

## 7. SDK API (Go)
- Интерфейсы:
```go
type ThreadManager interface {
    Start(ctx context.Context, assistant string) (*ThreadState, error)
    Close(ctx context.Context, state *ThreadState) error
}

type Agent interface {
    Run(ctx context.Context, input InputType, state *ThreadState) (OutputType, *ThreadState, *aiwf.Trace, error)
}

type Workflow interface {
    Run(ctx context.Context, input InputType, opts ...RunOption) (OutputType, *aiwf.Trace, error)
}
```
- Опции: `WithMaxDialogRounds(n int)`, `WithThreadManager(tm ThreadManager)`.

## 8. Тесты
- Юнит-тесты loader/IR для `threads` и `dialog`.
- Тесты генератора: без воркфлоу → не генерировать `workflows.go`; с диалогом → наличие цикла.
- Рантайм: имитация нескольких раундов, проверка `continue_dialog`.
- Провайдер OpenAI: мок HTTP-сервер, проверка `thread_id`, формата `text.format.name`.
- E2E: пример `examples/book` с диалогом и approval (инициализация state, несколько уточнений).

## 9. Документация
- Обновить `generator/README.md`, `runtime/README.md`, `docs/providers.md`.
- Обновить пример (`examples/blog` или `examples/book`) демонстрирующий диалог с аппрувами.

## 10. Релиз
- После реализации: bump версии модуля (например, `v0.4.0`), подготовить changelog.
- Обновить README/ROADMAP, описать миграцию.

