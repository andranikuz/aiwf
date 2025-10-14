package retry

import "time"

// Strategy задаёт параметры экспоненциального бэкоффа.
type Strategy struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

// Next рассчитывает, стоит ли повторять попытку и возвращает задержку.
func (s Strategy) Next(attempt int) (bool, time.Duration) {
	if attempt >= s.MaxAttempts {
		return false, 0
	}
	delay := s.BaseDelay << attempt
	if s.MaxDelay > 0 && delay > s.MaxDelay {
		delay = s.MaxDelay
	}
	return true, delay
}

// DefaultStrategy возвращает стратегию по умолчанию для провайдеров.
func DefaultStrategy() Strategy {
	return Strategy{MaxAttempts: 3, BaseDelay: 200 * time.Millisecond, MaxDelay: 2 * time.Second}
}
