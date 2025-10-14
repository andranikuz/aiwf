# Команды CLI

## aiwf validate
- Проверяет YAML-конфигурацию: `go run ./cmd/aiwf validate --file path/to/workflow.yaml`.
- Использует `core.LoadSpec` (валидирует JSON Schema через gojsonschema), затем `core.BuildIR` (проверяет DAG, scatter и зависимости).
- При успешной проверке выводит `✓ YAML валиден` с количеством ассистентов и воркфлоу.
- При ошибках печатает список строк вида `✗ field — message` (несколько строк на stderr) и завершает команду с ошибкой.
- Предупреждения выводятся строками `⚠ field — message`, но валидация остаётся успешной.

## aiwf sdk
- Генерирует Go SDK: `go run ./cmd/aiwf sdk --file workflows/novel.yaml --out ./sdk --package novelgen`.
- Перед генерацией повторно использует `validate`-проверки; ошибки блокируют процесс, предупреждения только печатаются.
- Результат сохраняется в указанном каталоге (`service.go`, интерфейсы агентов/воркфлоу).
*** End Patch
