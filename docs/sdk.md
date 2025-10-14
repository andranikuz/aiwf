# AIWF (AI Workflow) — Архитектура, интерфейсы и назначение

## 🌐 Общая идея

**AIWF** — это декларативный оркестратор для запуска многошаговых LLM-пайплайнов по YAML-конфигу. Он обеспечивает детерминизм, строгую типизацию и возможность генерации SDK для разных языков (Go, TypeScript, Python и др.) из одного описания.

Ключевая цель — дать разработчикам универсальный слой между моделями и бизнес-логикой: все схемы, промпты и воркфлоу описаны декларативно, SDK создаётся автоматически.

## 🎯 Основные задачи

* Обеспечить **детерминизм поверх LLM**: строгие JSON Schema, ретраи, автокоррекция.
* Разделить **контент и управление**: схемы и промпты — в YAML, код — чистая инфраструктура.
* Сгенерировать SDK для разных языков с едиными контрактами и интерфейсами.

## ⚙️ Текущее состояние реализации

* CLI основан на `spf13/cobra`: корневая команда `aiwf` уже обрабатывает глобальные флаги `--config`, `--env`, `--verbose` (`cmd/aiwf/main.go`).
* Базовые интерфейсы рантайма определены в `runtime/go/aiwf/contracts.go` и покрыты фиктивными тестами.
* Дорожная карта следующих слоёв зафиксирована в `ROADMAP.md` и синхронизируется с `AGENTS.md`.

## 🧱 Базовые интерфейсы (Go)

### ModelClient

```go
// Клиент модели (OpenAI/Anthropic/vLLM и т.д.)
type ModelClient interface{
    CallJSONSchema(ctx context.Context, in ModelCall) (raw []byte, usage Tokens, err error)
    CallJSONSchemaStream(ctx context.Context, in ModelCall) (chunks <-chan StreamChunk, usage Tokens, err error)
}
```

### Validator

```go
type Validator interface{
    Validate(schemaRef string, raw []byte) error
    ValidateBusiness(schemaRef string, raw []byte) error // доменные проверки
}
```

### TemplateEngine

```go
type TemplateEngine interface{
    RenderPrompt(pathOrInline string, data any) (string, error)
    RenderInput(templatePath string, data any) (any, error)
}
```

### ArtifactStore

```go
type ArtifactStore interface{
    Put(ctx context.Context, key string, bytes []byte) error
    Get(ctx context.Context, key string) ([]byte, bool, error)
    Key(workflow, step, itemKey, inputHash string) string
}
```

В репозитории уже есть реализации:

- `store.NewFSStore` — локальное файловое хранилище с TTL и периодической очисткой.
- `store.NewS3Store` — удалённое S3-совместимое хранилище с поддержкой префиксов.

### RetryPolicy

```go
type RetryPolicy interface{
    ShouldRetry(err error, attempt int) (retry bool, backoff time.Duration)
}
```

## 💬 Универсальные типы

```go
type Tokens struct{
    Prompt     int
    Completion int
    Total      int
}

type Trace struct{
    StepName   string
    Usage      Tokens
    Attempts   int
    Duration   time.Duration // требует import "time"
    ArtifactID string
}

type StreamChunk struct{
    Data       []byte
    Done       bool
    Partial    any
    Timestamps map[string]any
}
```

---

## 🤖 Обобщённые интерфейсы агентов, воркфлоу и тредов

### Agent

```go
type Agent[I any, O any] interface{
    Run(ctx context.Context, in I) (out O, tr *Trace, err error)
    RunStream(ctx context.Context, in I) (chunks <-chan StreamChunk, done <-chan Result[O,*Trace], err error)
}
```

### Workflow

```go
type Workflow[I any, O any] interface{
    Run(ctx context.Context, in I) (out O, tr *Trace, err error)
    RunStep(ctx context.Context, stepName string, payload any) (raw []byte, tr *Trace, err error)
}
```

### Thread

```go
type Thread interface{
    ID() string
    SendUser(ctx context.Context, msg any) (tr *Trace, err error)
    SendToolResult(ctx context.Context, tool string, payload any) (tr *Trace, err error)
    RunUntilIdle(ctx context.Context) (events []any, tr *Trace, err error)
    Snapshot(ctx context.Context) ([]byte, error)
}
```

---

## 📘 Реестр схем и описание ассистентов

```go
type SchemaRegistry interface{
    Get(schemaRef string) ([]byte, error)
    Describe(schemaRef string) (SchemaMeta, error)
}

type SchemaMeta struct{
    Name          string
    Discriminator string
    Enums         map[string][]string
    NullableProps map[string]bool
    OneOf         bool
}

// Ассистент (определён в YAML)
type AssistantSpec struct{
    Name            string
    Model           string
    SystemPromptRef string
    InputTemplate   string
    OutputSchemaRef string
    Temperature     *float32
    MaxOutputTokens *int
    Tools           []ToolSpec
}

type ToolSpec struct{
    Name string
    InputSchemaRef  string
    OutputSchemaRef string
}

// Scatter-конфигурация

type ScatterSpec struct{
    FromExpr    string
    As          string
    Concurrency int
}

// Узел DAG

type StepSpec struct{
    Name       string
    Assistant  string
    Needs      []string
    Scatter    *ScatterSpec
    InputBinding map[string]string
}

// Воркфлоу
type WorkflowSpec struct{
    Name  string
    Steps []StepSpec
}
```

---

## 🧩 Service — точка входа SDK

```go
type Service interface{
    Agents()   Agents
    Workflows() Workflows
    Threads()  Threads
    Schemas()  SchemaRegistry
}

type Agents interface{
    Premise() PremiseAgent
    Outline() OutlineAgent
    Draft()   DraftAgent
}

type Workflows interface{
    Novel() NovelRunner
    DayPlanner() DayPlannerRunner
}

type Threads interface{
    NewNovelThread(seed any) (Thread, error)
}
```

Пример типизированных агентов и воркфлоу:

```go
type PremiseAgent interface{ Agent[PremiseInput, PremiseOutput] }
type OutlineAgent interface{ Agent[OutlineInput, OutlineOutput] }
type NovelRunner interface{ Workflow[NovelInput, NovelOutput] }
```

---

## ⚙️ Контракты данных (пример)

```go
type Tone string
const (
    ToneDark    Tone = "dark"
    ToneHopeful Tone = "hopeful"
    TonePlayful Tone = "playful"
)

type PremiseOutput struct{
    Logline string   `json:"logline"`
    Themes  []string `json:"themes"`
    Tone    Tone     `json:"tone"`
}

type PremiseInput struct{
    Idea string `json:"idea"`
}
```

---

## 💡 Ошибки и константы

```go
var (
    ErrSchemaViolation = errors.New("aiwf: schema violation")
    ErrBusinessRule    = errors.New("aiwf: business rule failed")
    ErrModelCall       = errors.New("aiwf: model call failed")
    ErrIdempotentHit   = errors.New("aiwf: artifact cache hit")
)

type SchemaError struct{
    SchemaRef string
    Path      string
    Msg       string
}
func (e *SchemaError) Error() string { return fmt.Sprintf("%s %s: %s", e.SchemaRef, e.Path, e.Msg) }
```

---

## 🧠 Роль генератора SDK

Генератор `aiwf workflow sdk`:

* парсит YAML (assistants, workflows, schema_registry);
* генерирует SDK-пакет для выбранных языков;
* создаёт типизированные контракты и фасады `Service`;
* обеспечивает валидацию, ретраи, артефакты и наблюдаемость.

### Пример CLI

```bash
aiwf workflow sdk \
  --in internal/workflows/novel.yaml \
  --lang go,ts,py \
  --out ./_generated \
  --module-go github.com/you/aiwf_novel \
  --package-ts @you/aiwf-novel \
  --package-py aiwf-novel
```

---

## 🚀 Пример использования (Go)

```go
svc := aiwfgen.NewService(aiwf.Options{Client: openAI, Store: fsStore, Retry: backoff})
premise, tr, _ := svc.Agents().Premise().Run(ctx, contracts.PremiseInput{Idea: "Путешествие во сне"})
outline, _, _ := svc.Agents().Outline().Run(ctx, contracts.OutlineInput{Premise: premise})
novel, _, _ := svc.Workflows().Novel().Run(ctx, contracts.NovelInput{Idea: premise.Idea})
```

---

## 📚 Резюме

**AIWF** — это генератор SDK для LLM-пайплайнов по декларативному описанию:

* один YAML → несколько SDK (Go, TS, Python);
* строгие JSON Schema → типизированные контракты;
* фасад `Service` для удобного доступа к агентам и воркфлоу;
* интерфейсы `ModelClient`, `Validator`, `ArtifactStore`, `RetryPolicy` обеспечивают расширяемость.

Применимо в проектах: генерация книг, планировщик дня, ассистенты с DAG-логикой, AI-инструменты и агенты с детерминированными выходами.
