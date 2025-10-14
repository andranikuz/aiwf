# Generator Core

## YAML Loader
- Функция `core.LoadSpec(path)` читает YAML, применяет `schema_registry.root` и гарантирует наличие всех ассистентов и их JSON Schema.
- Для каждого ассистента вычисляются абсолютные пути к схемам (`AssistantSpec.Resolved`) и проверяется, что JSON Schema корректна (парсится через gojsonschema).
- Workflow-шага проверяются на существование ссылок на ассистентов.

## Использование в CLI
- Ближайшая интеграция — команда `aiwf validate`, которая будет вызывать `LoadSpec` и поверх этого добавлять схемную валидацию и генерацию IR.
- Ошибки возвращаются в виде `ValidationError` с заполненными полями (`Field`, `Msg`), что облегчает подсветку проблем в CLI.

## Промежуточное представление (IR)
- `core.BuildIR(spec)` преобразует `Spec` в нормализованную структуру для генераторов (ассистенты, воркфлоу, DAG с `scatter` и `input_binding`).
- Проверяются уникальность шагов, корректность зависимостей (`needs` только к уже объявленным шагам) и валидность `scatter` (поля `from`, `as`, `concurrency`). Неиспользуемые ассистенты и пустые воркфлоу фиксируются как предупреждения.
- `IRAssistant`, `IRWorkflow`, `IRStep` используются при рендеринге Go/TS/Python шаблонов.

## Go backend
- Пакет `generator/backend-go` собирает фасад `Service`, интерфейсы агентов/воркфлоу и заглушки выполнения.
- Шаблоны хранятся в `generator/backend-go/templates/` и встраиваются через `embed`. Генерируются файлы `service.go`, `agents.go`, `workflows.go`, `contracts.go`.
- Голден-файлы лежат в `generator/backend-go/testdata`, обновляются командой `UPDATE_GOLDEN=1 go test ./generator/backend-go`.

## CLI `aiwf validate`
- Подкоманда `validate` использует `core.LoadSpec` и `core.BuildIR`, чтобы проверить YAML и вывести сводку по количеству ассистентов и воркфлоу.
- При ошибках `ValidationError` пробрасывается напрямую, что позволяет увидеть поле и сообщение.
