# Runtime

Runtime-библиотеки для выполнения сгенерированного SDK.

## Go Runtime (`runtime/go/aiwf`)

### Основные интерфейсы

- **`ModelClient`** - интерфейс для вызова LLM
  - `CallJSONSchema` - синхронный вызов с JSON Schema
  - `CallJSONSchemaStream` - потоковый вызов (в разработке)

- **`ThreadManager`** - управление состоянием диалогов
  - `Start` - начало нового треда
  - `Continue` - продолжение с обратной связью
  - `Close` - завершение треда

- **`ArtifactStore`** - хранение промежуточных результатов
  - Реализации: filesystem, S3

- **`TypeProvider`** - предоставление метаданных типов для провайдеров

- **`AgentBase`** - базовая реализация агента с CallModel

### Использование

```go
import "github.com/andranikuz/aiwf/runtime/go/aiwf"

// Создание клиента провайдера
client := openai.NewClient(config)

// Создание сервиса SDK
service := sdk.NewService(client)
    .WithThreadManager(threadManager)
    .WithArtifactStore(store)

// Вызов агента
result, trace, err := service.Agents().DataExtractor.Run(ctx, input)
```

### Контракты

- **`ModelCall`** - структура запроса к LLM
- **`Tokens`** - метрики использования токенов
- **`Trace`** - трассировка выполнения
- **`ThreadState`** - состояние диалогового треда

## Roadmap

- [ ] Python runtime
- [ ] TypeScript runtime
- [ ] Полная поддержка стриминга
- [ ] Batch операции
- [ ] Middleware система
