package aiwf

import (
	"context"
	"fmt"
)

// WorkflowStep определяет один шаг в workflow'е
type WorkflowStep interface {
	// GetName возвращает имя шага
	GetName() string

	// Execute выполняет шаг и возвращает результат
	Execute(ctx context.Context, input any) (any, error)

	// GetDependencies возвращает список зависимостей (имена шагов)
	GetDependencies() []string
}

// WorkflowStepResult содержит результат выполнения шага
type WorkflowStepResult struct {
	StepName  string
	Output    any
	Error     error
	Duration  int64 // nanoseconds
	Attempts  int
	Trace     *Trace
}

// WorkflowDefinition определяет структуру workflow'а
type WorkflowDefinition struct {
	Name        string
	Description string
	Steps       map[string]WorkflowStep
	// DAG edges: stepName -> [dependent stepNames]
	Dependencies map[string][]string
}

// NewWorkflowDefinition создает новое определение workflow'а
func NewWorkflowDefinition(name string) *WorkflowDefinition {
	return &WorkflowDefinition{
		Name:         name,
		Steps:        make(map[string]WorkflowStep),
		Dependencies: make(map[string][]string),
	}
}

// AddStep добавляет шаг в workflow
func (w *WorkflowDefinition) AddStep(step WorkflowStep) error {
	name := step.GetName()
	if _, exists := w.Steps[name]; exists {
		return fmt.Errorf("step %q already exists", name)
	}

	w.Steps[name] = step

	// Record dependencies
	deps := step.GetDependencies()
	for _, dep := range deps {
		if _, exists := w.Steps[dep]; !exists {
			return fmt.Errorf("dependency %q not found for step %q", dep, name)
		}
	}

	w.Dependencies[name] = deps
	return nil
}

// ValidateDAG проверяет что workflow не содержит циклов
func (w *WorkflowDefinition) ValidateDAG() error {
	// Проверяем что все зависимости существуют
	for stepName, deps := range w.Dependencies {
		for _, dep := range deps {
			if _, exists := w.Steps[dep]; !exists {
				return fmt.Errorf("step %q depends on non-existent step %q", stepName, dep)
			}
		}
	}

	// Проверяем на циклы используя DFS
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var hasCycle func(step string) bool
	hasCycle = func(step string) bool {
		visited[step] = true
		recStack[step] = true

		for _, dep := range w.Dependencies[step] {
			if !visited[dep] {
				if hasCycle(dep) {
					return true
				}
			} else if recStack[dep] {
				return true
			}
		}

		recStack[step] = false
		return false
	}

	for stepName := range w.Steps {
		if !visited[stepName] {
			if hasCycle(stepName) {
				return fmt.Errorf("workflow contains cycle involving step %q", stepName)
			}
		}
	}

	return nil
}

// GetTopologicalOrder возвращает шаги в топологическом порядке
func (w *WorkflowDefinition) GetTopologicalOrder() ([]string, error) {
	if err := w.ValidateDAG(); err != nil {
		return nil, err
	}

	// Строим обратный граф (who depends on me)
	dependents := make(map[string][]string)
	inDegree := make(map[string]int)

	for stepName := range w.Steps {
		inDegree[stepName] = 0
		dependents[stepName] = []string{}
	}

	for stepName, deps := range w.Dependencies {
		inDegree[stepName] = len(deps)
		for _, dep := range deps {
			dependents[dep] = append(dependents[dep], stepName)
		}
	}

	// Kahn's algorithm
	var queue []string
	for stepName, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, stepName)
		}
	}

	var result []string
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		for _, dependent := range dependents[current] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}

	if len(result) != len(w.Steps) {
		return nil, fmt.Errorf("workflow contains cycles")
	}

	return result, nil
}

// WorkflowExecutor выполняет workflow
type WorkflowExecutor struct {
	definition *WorkflowDefinition
	results    map[string]*WorkflowStepResult
	maxRetries int
}

// NewWorkflowExecutor создает новый executor
func NewWorkflowExecutor(def *WorkflowDefinition) (*WorkflowExecutor, error) {
	if err := def.ValidateDAG(); err != nil {
		return nil, err
	}

	return &WorkflowExecutor{
		definition: def,
		results:    make(map[string]*WorkflowStepResult),
		maxRetries: 3,
	}, nil
}

// SetMaxRetries устанавливает максимальное количество повторов
func (e *WorkflowExecutor) SetMaxRetries(max int) {
	e.maxRetries = max
}

// Execute выполняет весь workflow
func (e *WorkflowExecutor) Execute(ctx context.Context, initialInput any) (map[string]*WorkflowStepResult, error) {
	// Получаем топологический порядок
	order, err := e.definition.GetTopologicalOrder()
	if err != nil {
		return nil, fmt.Errorf("invalid workflow: %w", err)
	}

	// Выполняем шаги в порядке
	for _, stepName := range order {
		step := e.definition.Steps[stepName]

		// Подготавливаем input на основе зависимостей
		stepInput := initialInput
		deps := e.definition.Dependencies[stepName]

		if len(deps) > 0 && len(deps) == 1 {
			// Если одна зависимость, используем ее output
			if result, exists := e.results[deps[0]]; exists && result.Error == nil {
				stepInput = result.Output
			}
		} else if len(deps) > 1 {
			// Если несколько зависимостей, создаем map с результатами
			depResults := make(map[string]any)
			for _, dep := range deps {
				if result, exists := e.results[dep]; exists {
					if result.Error != nil {
						return nil, fmt.Errorf("dependency %q failed: %w", dep, result.Error)
					}
					depResults[dep] = result.Output
				}
			}
			stepInput = depResults
		}

		// Выполняем шаг с retry
		result := e.executeWithRetry(ctx, stepName, step, stepInput)
		e.results[stepName] = result

		if result.Error != nil {
			return e.results, fmt.Errorf("step %q failed: %w", stepName, result.Error)
		}
	}

	return e.results, nil
}

// executeWithRetry выполняет шаг с повторами
func (e *WorkflowExecutor) executeWithRetry(ctx context.Context, stepName string, step WorkflowStep, input any) *WorkflowStepResult {
	result := &WorkflowStepResult{
		StepName: stepName,
	}

	start := getTime()

	for attempt := 1; attempt <= e.maxRetries; attempt++ {
		result.Attempts = attempt

		output, err := step.Execute(ctx, input)
		if err == nil {
			result.Output = output
			result.Duration = getTime() - start
			return result
		}

		// Если это последняя попытка, сохраняем ошибку
		if attempt == e.maxRetries {
			result.Error = err
			result.Duration = getTime() - start
			return result
		}

		// Иначе повторяем
	}

	result.Duration = getTime() - start
	return result
}

// GetResult возвращает результат выполнения шага
func (e *WorkflowExecutor) GetResult(stepName string) (*WorkflowStepResult, bool) {
	result, exists := e.results[stepName]
	return result, exists
}

// GetAllResults возвращает все результаты
func (e *WorkflowExecutor) GetAllResults() map[string]*WorkflowStepResult {
	return e.results
}

// SimpleStep простая реализация WorkflowStep для тестирования
type SimpleStep struct {
	name         string
	dependencies []string
	executeFunc  func(ctx context.Context, input any) (any, error)
}

// NewSimpleStep создает простой шаг
func NewSimpleStep(name string, fn func(ctx context.Context, input any) (any, error)) *SimpleStep {
	return &SimpleStep{
		name:        name,
		dependencies: []string{},
		executeFunc: fn,
	}
}

// WithDependencies устанавливает зависимости
func (s *SimpleStep) WithDependencies(deps ...string) *SimpleStep {
	s.dependencies = deps
	return s
}

func (s *SimpleStep) GetName() string {
	return s.name
}

func (s *SimpleStep) Execute(ctx context.Context, input any) (any, error) {
	if s.executeFunc == nil {
		return input, nil
	}
	return s.executeFunc(ctx, input)
}

func (s *SimpleStep) GetDependencies() []string {
	return s.dependencies
}

// Helper для получения времени в наносекундах
func getTime() int64 {
	return int64(0) // Placeholder - использовать time.Now().UnixNano() в реальности
}
