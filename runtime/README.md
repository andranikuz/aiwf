# Рантайм AIWF

`runtime/` содержит контракты и сторы, необходимые для запуска сгенерированных воркфлоу.

## Пакеты

- `go/aiwf` — базовые интерфейсы и типы:
  - `ModelClient` — абстракция над LLM (методы `CallJSONSchema`, `CallJSONSchemaStream`).
  - `Workflow`, `Trace`, `Tokens` — единый контракт для SDK.
  - `ArtifactStore` — хранение промежуточных артефактов (промпты, JSON).
- `go/aiwf/store` — реализации `ArtifactStore`:
  - `filesystem` — локальное хранение с TTL и очисткой;
  - `s3` — сохранение артефактов в S3;
  - вспомогательный `hashutil`.

## Как использовать

1. Реализуйте `ModelClient`, который конвертирует `aiwf.ModelCall` в запрос к провайдеру.
2. Создайте `aiwf.ArtifactStore` (например, `store.NewFilesystemStore`).
3. Передайте клиент в сгенерированный SDK: `sdk.NewService(modelClient)`.
4. Запускайте воркфлоу `service.Workflows().<Name>().Run(ctx, input)`.

## Пример

См. `examples/blog/example_usage.go` — минимальный `ModelClient`, который имитирует ответы и печатает результат воркфлоу.

## Тесты

```bash
go test ./runtime/...
```

Для хранилища S3 нужны переменные окружения `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, либо используйте `-run TestS3` с `testing.Short()` во время разработки.

