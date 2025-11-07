# Quick Start: AI-Powered Config Generation

Самый быстрый способ создать AIWF конфигурацию - использовать AI генератор.

## 1 минута: Интерактивная генерация

```bash
# Запустите интерактивный режим
aiwf generate --interactive

# Следуйте инструкциям:
# 1. Опишите задачу
# 2. Просмотрите анализ
# 3. Утвердите или уточните
# 4. Получите готовый YAML
```

## 30 секунд: Быстрая генерация

```bash
aiwf generate -t "Create a spam filter for emails" -o spam-filter.yaml
```

## Примеры задач

### Простые (1 агент)

```bash
aiwf generate -t "Translate text from English to Spanish"
aiwf generate -t "Classify sentiment: positive, negative, or neutral"
aiwf generate -t "Extract keywords from text"
```

### Средние (2-3 агента)

```bash
aiwf generate -t "Content moderation: detect toxic language, spam, and PII"
aiwf generate -t "Data validation pipeline: validate then process"
aiwf generate -t "Customer support with automatic routing"
```

### Сложные (3+ агента)

```bash
aiwf generate -t "Interactive tutoring system that adapts difficulty and tracks progress"
aiwf generate -t "Multi-stage document analysis with summarization and Q&A"
```

## Интерактивный процесс

```
📝 Describe your task:
> Create a content moderation system for social media posts

📊 Analysis:
   Complexity: MEDIUM
   Agents: 3
   - toxicity_detector: Detect toxic/offensive language
   - spam_classifier: Identify spam content
   - pii_detector: Find personal information

❓ Questions:
   1. What should happen when multiple issues detected?
   Your answer: Flag all issues, prioritize most severe

✅ Continue? [c/r/q]: c

⚙️ Generating YAML...
✓ Done!

💾 Save? [s/e/q]: s
✓ Saved to generated-config.yaml
```

## После генерации

```bash
# Проверить
cat generated-config.yaml

# Валидировать
aiwf validate -f generated-config.yaml

# Сгенерировать SDK
aiwf sdk -f generated-config.yaml -o ./generated

# Или сразу запустить сервер
aiwf serve -f generated-config.yaml
```

## Уточнения и правки

Вы всегда можете:
- Ответить на уточняющие вопросы
- Добавить refinement instructions
- Запросить изменения в сгенерированном YAML
- Итеративно улучшать результат

## Требования

- API ключ: `export OPENAI_API_KEY="sk-..."`
- Go 1.24+
- AIWF CLI установлен

## Следующие шаги

- [Полная документация](../../docs/GENERATE_GUIDE.md)
- [Примеры использования](./README.md)
- [Система типов AIWF](../../generator/README.md)
