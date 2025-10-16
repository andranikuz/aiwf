# Рантайм AIWF

`runtime/` содержит контракты и утилиты, необходимые для исполнения сгенерированных воркфлоу.

## Основные пакеты

- `runtime/go/aiwf`
  - `ModelClient` — интерфейс LLM-клиента (`CallJSONSchema`, `CallJSONSchemaStream`).
  - `ThreadState`, `ThreadBinding`, `ThreadManager` — новая подсистема для управления диалоговыми тредами.
  - `DialogAction`, `DialogDecider`, `DefaultDialogDecider`, `NoopThreadManager` — механика диалогов.
  - `Workflow` — базовый интерфейс для раннеров (сгенерированный SDK его реализует).
  - `Tokens`, `Trace`, `ArtifactStore` — контракты наблюдаемости и артефактов.
- `runtime/go/aiwf/store` — реализации `ArtifactStore` (filesystem, s3).

## Диалоговые воркфлоу

- `ThreadManager` управляет `ThreadState` между вызовами агента (создание/продолжение/закрытие).
- `DialogDecider` решает, повторять ли шаг (`DialogActionRetry`), переходить к другому (`DialogActionGoto`), завершать.
- По умолчанию используется `NoopThreadManager` и `DefaultDialogDecider`, работающие как single-shot.

## Реализация провайдеров

- Провайдеры должны поддерживать `ThreadState` (например, OpenAI Responses API использует `thread_id`).
- При `DialogActionRetry` генератор вызывает `ThreadManager.Continue` с текстом feedback.

## Пример использования

Сгенерированный SDK (например, `examples/blog/sdk`) вызывает агента примерно так:

```go
state, _ := threadManager.Start(ctx, "premise", aiwf.ThreadBinding{Name: "default", Provider: "openai_responses"})
output, state, trace, err := service.Agents().Premise().Run(ctx, input, state)
```

Пользователь может передать свой менеджер тредов:

```go
svc := blog.NewService(client).
    WithThreadManager(myThreads).
    WithDialogDecider(myDecider).
    WithMaxDialogRounds(3)
```

## Тестирование

```bash
unset GOROOT && export GOTOOLCHAIN=local
go test ./runtime/...
```

`runtime/go/aiwf/dialog.go` и `runtime/go/aiwf/contracts.go` — отправные точки для расширения собственного рантайма.
