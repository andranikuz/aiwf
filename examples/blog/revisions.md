# Конфигурация AIWF Workflow — Необходимые доработки (v0.2)

## Обзор

Этот документ описывает **необходимые улучшения и уточнения**, чтобы финализировать YAML-конфигурацию воркфлоу для AIWF (AI Workflow Framework). Он основан на текущей версии воркфлоу (`version: 0.2`) и охватывает структурные, семантические и рантайм-требования.
Пример yaml с доработками в папке config/sdk.yaml

---

## 1. Контракты в YAML и их повторное использование

Новое правило:
Контракты (Input, Output, Review, и т.д.) теперь хранятся в YAML-файлах и могут переиспользоваться между воркфлоу и ассистентами через импорт. 
Для type: object необходимо генерировать структуры всегда, внутри указывть не просто []map[string]any а настоящий тип

Преимущества:

Единый источник истины для всех схем данных.

Возможность шаринга типов между ассистентами и шагами разных воркфлоу.

Упрощённая валидация и генерация SDK — типы подгружаются напрямую из YAML.

```yaml
# common.types.yaml
PremiseOutput:
  $id: "aiwf://book/PremiseOutput"
  type: object
  properties:
    logline: { type: string, maxLength: 160 }
    themes: { type: array, items: { type: string } }
    tone: { type: string }

OutlineInput:
  $id: "aiwf://book/OutlineInput"
  type: object
  required: [premise]
  properties:
    premise:
      $ref: "aiwf://book/PremiseOutput"
```
```yaml
# main workflow.yaml
imports:
  - as: common
    path: ./common.types.yaml
  - as: book
    path: ./book.types.yaml
```

## 2 Контракты входов для следующих шагов

**Проблема:** Переход от `premise` к `outline` не имеет явного контракта входных данных.

**Решение:** Определить `OutlineInput` в `book.types.yaml` и сослаться на него в воркфлоу.

```yaml
# book.types.yaml
OutlineInput:
  $id: "aiwf://book/OutlineInput"
  type: object
  required: [premise]
  properties:
    premise:
      $ref: "aiwf://book/PremiseOutput"
```

```yaml
next:
  step: outline
  input_binding:
    premise: "{{ .Premise }}"
  input_contract_ref: "aiwf://book/OutlineInput"
```

---

## 3. Определить финальный или следующий шаг для `outline`

**Проблема:** У шага `outline` нет правила завершения или перехода.

**Решение:** Либо:

* Сделать шаг `outline` финальным (без `next`), либо
* Явно указать `next`, если поток продолжается.

```yaml
- step: outline
  assistant: outline
  thread: { use: default, strategy: append }
  # финальный шаг
```

---

## 4. Приоритет между `system_prompt` и `system_prompt_ref`

**Уточнение:** Если оба поля указаны, приоритет имеет `system_prompt_ref`.

```yaml
# Правило выбора промпта
# - Если указаны оба поля, используется system_prompt_ref.
# - Если указано только одно — использовать его.
```

---

## 5. Формальная спецификация `approval.review`

**Цель:** Обеспечить согласованность логики ревью (LLM-проверки).

### Поля схемы

* `strictness`: целое число [0–10]
* `prompt`: общий промпт для проверки
* `properties`: объект — dot-path → `{ prompt, strictness }`

### Правила

* Поддерживаются пути: `a.b`, `array[]`, `array[*].field`.
* Более специфичные пути перекрывают глобальные.
* Рантайм формирует стандартную структуру результата:

```json
{
  "approved": true,
  "issues": [ {"path":"logline","message":"слишком длинный","severity":8} ],
  "score": 9,
  "edits": { }
}
```

---

## 6. Поведение `on_approve` и `on_reject`

**Улучшение:** Разрешить указывать конкретные шаги для перехода.

```yaml
approval:
  on_approve: { action: continue | retry | goto | stop, step?: <stepName> }
  on_reject:  { action: retry | goto | stop, step?: <stepName> }
```

### Правила валидации

* Поле `step` обязательно, если `action: goto`.
* `goto` имеет приоритет над `next`.
* Соблюдать `require_for_retry` и `max_retries`.

---

## 7. Проверка dot-path ссылок в `approval.review.properties`

**Решение:** Добавить семантическую валидацию, чтобы каждый dot-path соответствовал полю в `output_schema_ref` шага.

**Действие:** Интегрировать проверку путей против JSON Schema с отчётом об ошибке при несоответствии.

---

## 8. Управление тредами и сохранение состояния

**Улучшение:** Добавить явное поле для хранения идентификаторов тредов.

```yaml
threads:
  default:
    provider: openai_assistants
    strategy: append
    external: { thread_id: null }  # заполняется рантаймом
    metadata: { project: "novel", ttl_hours: 24 }
```

* `reset_before_step`: эмулировать через локальные секции, если провайдер не поддерживает.
* Хранить состояние тредов по каждому запуску (`threads_state/<run_id>.json`).

---

## 9. Политика артефактов и идемпотентности

**Цель:** Поддержка ретраев и аудита.

```yaml
runtime:
  artifacts:
    kind: fs
    root: ./artifacts
  idempotency:
    key_strategy: hash(input)
    include_thread: true
```

---

## 10. Стандартизированные коды ошибок

**Добавить для унификации CLI/SDK:**

* `SCHEMA_VIOLATION`
* `BUSINESS_RULE_FAILED`
* `REVIEW_REJECTED`
* `RETRY_LIMIT`
* `APPROVAL_REQUIRED`
* `THREAD_ERROR`
* `PROVIDER_ERROR`
* `CONFIG_ERROR`
* `TIMEOUT`

Также - добавить возможность сгенерировать только агентов, без создания воркфлоу (сейчсас при их отсутствии валидационная ошибка)

---

## 11. CLI-команды для управления аппрувами

**Улучшение:** Добавить команды для интерактивного ревью.

```bash
aiwf approvals ls --run <id>
aiwf approve --run <id> --step <name> --payload '{"approved":true,"comment":"ok"}'
aiwf retry --run <id> --step <name>
```

При `retry` рантайм должен автоматически добавлять комментарий и список ошибок (`issues`) в промпт LLM.

---

## 12. Рекомендации по тестированию

| Тип           | Область                                                      |
| ------------- | ------------------------------------------------------------ |
| Структурные   | Проверка YAML-схемы, обязательных полей, диапазонов          |
| Семантические | Целостность DAG, наличие шагов для `goto`, проверка dot-path |
| E2E           | Сценарий: `premise → reject → retry → approve → outline`     |

---

## 13. Прочие уточнения

* Убрать точку в конце строки у `tone.prompt` (для чистоты).
* Русскоязычные промпты (например, `premise.system_prompt`) вынести в отдельные файлы `/prompts` для повторного использования и перевода.

---

### ✅ Итог

После внесения этих изменений:

* Воркфлоу станет детерминированным и согласованным по переходам.
* Система ревью получит контроль по полям и уровням строгости.
* Треды и артефакты будут сохраняться и переиспользоваться.
* Рантайм и CLI будут иметь унифицированный контроль и API-интерфейсы.

**Целевая версия:** `v0.3` — готово к интеграции с генератором и рантаймом.
