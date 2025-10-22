package aiwf

import "context"

// DialogAction описывает возможные действия после ревью шага.
type DialogAction int

const (
	DialogActionContinue DialogAction = iota
	DialogActionRetry
	DialogActionGoto
	DialogActionStop
	DialogActionComplete
)

// DialogContext содержит информацию о текущем состоянии шага.
type DialogContext struct {
	Step    string
	Output  any
	Trace   *Trace
	Attempt int
}

// DialogDecision описывает результат решения ревьюера.
type DialogDecision struct {
	Action   DialogAction
	Feedback string
	Target   string
}

// DialogDecider принимает решение о продолжении диалога.
type DialogDecider interface {
	Decide(DialogContext) DialogDecision
}

// DefaultDialogDecider завершает диалог после первого прохода.
type DefaultDialogDecider struct{}

// Decide реализует интерфейс DialogDecider.
func (DefaultDialogDecider) Decide(ctx DialogContext) DialogDecision {
	return DialogDecision{Action: DialogActionComplete}
}

// NoopThreadManager не управляет тредами и используется по умолчанию.
type NoopThreadManager struct{}

// Start возвращает nil-состояние без ошибок.
func (NoopThreadManager) Start(context.Context, string, ThreadBinding) (*ThreadState, error) {
	return nil, nil
}

// Continue ничего не делает.
func (NoopThreadManager) Continue(context.Context, *ThreadState, string) error { return nil }

// Close ничего не делает.
func (NoopThreadManager) Close(context.Context, *ThreadState) error { return nil }
