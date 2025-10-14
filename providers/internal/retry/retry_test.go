package retry

import "testing"

func TestStrategyNext(t *testing.T) {
	strat := Strategy{MaxAttempts: 2, BaseDelay: 100, MaxDelay: 150}
	if ok, delay := strat.Next(0); !ok || delay != 100 {
		t.Fatalf("expected first attempt 100ms, got ok=%v delay=%v", ok, delay)
	}
	if ok, delay := strat.Next(1); !ok || delay != 150 {
		t.Fatalf("expected capped delay 150ms, got ok=%v delay=%v", ok, delay)
	}
	if ok, _ := strat.Next(2); ok {
		t.Fatalf("expected stop after max attempts")
	}
}
