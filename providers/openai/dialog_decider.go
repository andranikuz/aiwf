package openai

import (
	"encoding/json"
	"fmt"

	"github.com/andranikuz/aiwf/runtime/go/aiwf"
)

// DefaultDialogDecider завершает диалог после первого проход (уже реализован в runtime)
// Оставляем как ссылку на runtime реализацию

// ImmediateRetryDecider всегда запрашивает повторное выполнение с обратной связью
type ImmediateRetryDecider struct {
	MaxAttempts int
	Message     string
}

// NewImmediateRetryDecider создает новый ImmediateRetryDecider
func NewImmediateRetryDecider(maxAttempts int, message string) *ImmediateRetryDecider {
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	if message == "" {
		message = "Please review and retry this step"
	}
	return &ImmediateRetryDecider{
		MaxAttempts: maxAttempts,
		Message:     message,
	}
}

// Decide реализует интерфейс DialogDecider
func (d *ImmediateRetryDecider) Decide(ctx aiwf.DialogContext) aiwf.DialogDecision {
	// Если превышены попытки, завершаем
	if ctx.Attempt > d.MaxAttempts {
		return aiwf.DialogDecision{
			Action: aiwf.DialogActionComplete,
		}
	}

	// Иначе запрашиваем повторное выполнение
	return aiwf.DialogDecision{
		Action:   aiwf.DialogActionRetry,
		Feedback: d.Message,
	}
}

// QualityCheckDecider проверяет качество вывода перед завершением
type QualityCheckDecider struct {
	// Функция для проверки качества
	CheckFn func(output any) (bool, string)
	// Максимальное количество попыток
	MaxAttempts int
}

// NewQualityCheckDecider создает новый QualityCheckDecider
func NewQualityCheckDecider(checkFn func(output any) (bool, string)) *QualityCheckDecider {
	return &QualityCheckDecider{
		CheckFn:     checkFn,
		MaxAttempts: 3,
	}
}

// Decide реализует интерфейс DialogDecider
func (d *QualityCheckDecider) Decide(ctx aiwf.DialogContext) aiwf.DialogDecision {
	if d.CheckFn == nil {
		return aiwf.DialogDecision{Action: aiwf.DialogActionComplete}
	}

	// Проверяем качество вывода
	passed, feedback := d.CheckFn(ctx.Output)
	if passed {
		return aiwf.DialogDecision{
			Action: aiwf.DialogActionComplete,
		}
	}

	// Если не прошла проверка и не превышены попытки, повторяем
	if ctx.Attempt <= d.MaxAttempts {
		return aiwf.DialogDecision{
			Action:   aiwf.DialogActionRetry,
			Feedback: feedback,
		}
	}

	// Превышены попытки, завершаем с текущим выводом
	return aiwf.DialogDecision{
		Action: aiwf.DialogActionComplete,
	}
}

// LengthCheckDecider проверяет, что вывод имеет минимальную длину
type LengthCheckDecider struct {
	MinLength   int
	MaxAttempts int
}

// NewLengthCheckDecider создает новый LengthCheckDecider
func NewLengthCheckDecider(minLength int) *LengthCheckDecider {
	if minLength <= 0 {
		minLength = 50
	}
	return &LengthCheckDecider{
		MinLength:   minLength,
		MaxAttempts: 3,
	}
}

// Decide реализует интерфейс DialogDecider
func (d *LengthCheckDecider) Decide(ctx aiwf.DialogContext) aiwf.DialogDecision {
	outputStr := fmt.Sprintf("%v", ctx.Output)

	if len(outputStr) >= d.MinLength {
		return aiwf.DialogDecision{
			Action: aiwf.DialogActionComplete,
		}
	}

	if ctx.Attempt <= d.MaxAttempts {
		return aiwf.DialogDecision{
			Action:   aiwf.DialogActionRetry,
			Feedback: fmt.Sprintf("Output is too short (current: %d chars, minimum: %d chars). Please provide a more detailed response.", len(outputStr), d.MinLength),
		}
	}

	return aiwf.DialogDecision{
		Action: aiwf.DialogActionComplete,
	}
}

// JSONValidationDecider проверяет, что вывод - валидный JSON и соответствует структуре
type JSONValidationDecider struct {
	Schema      map[string]any // Простая схема для проверки
	MaxAttempts int
}

// NewJSONValidationDecider создает новый JSONValidationDecider
func NewJSONValidationDecider(schema map[string]any) *JSONValidationDecider {
	return &JSONValidationDecider{
		Schema:      schema,
		MaxAttempts: 3,
	}
}

// Decide реализует интерфейс DialogDecider
func (d *JSONValidationDecider) Decide(ctx aiwf.DialogContext) aiwf.DialogDecision {
	// Пытаемся распарсить как JSON
	var jsonData map[string]any
	outputStr := fmt.Sprintf("%v", ctx.Output)

	if err := json.Unmarshal([]byte(outputStr), &jsonData); err != nil {
		if ctx.Attempt <= d.MaxAttempts {
			return aiwf.DialogDecision{
				Action:   aiwf.DialogActionRetry,
				Feedback: fmt.Sprintf("Output must be valid JSON. Error: %v", err),
			}
		}
		return aiwf.DialogDecision{
			Action: aiwf.DialogActionComplete,
		}
	}

	// Если схема определена, проверяем наличие требуемых полей
	if len(d.Schema) > 0 {
		for field := range d.Schema {
			if _, ok := jsonData[field]; !ok {
				if ctx.Attempt <= d.MaxAttempts {
					return aiwf.DialogDecision{
						Action:   aiwf.DialogActionRetry,
						Feedback: fmt.Sprintf("JSON must include required field: %s", field),
					}
				}
			}
		}
	}

	return aiwf.DialogDecision{
		Action: aiwf.DialogActionComplete,
	}
}

// ChainedDecider применяет несколько проверяющих функций последовательно
type ChainedDecider struct {
	Deciders []aiwf.DialogDecider
}

// NewChainedDecider создает новый ChainedDecider
func NewChainedDecider(deciders ...aiwf.DialogDecider) *ChainedDecider {
	return &ChainedDecider{
		Deciders: deciders,
	}
}

// Decide реализует интерфейс DialogDecider
func (d *ChainedDecider) Decide(ctx aiwf.DialogContext) aiwf.DialogDecision {
	for _, decider := range d.Deciders {
		decision := decider.Decide(ctx)
		// Если какой-то decider запрашивает действие, отличное от Complete, возвращаем его
		if decision.Action != aiwf.DialogActionComplete {
			return decision
		}
	}
	// Все deciders завершены, возвращаем Complete
	return aiwf.DialogDecision{
		Action: aiwf.DialogActionComplete,
	}
}
