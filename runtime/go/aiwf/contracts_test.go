package aiwf

import (
	"context"
	"testing"
	"time"
)

type fakeClient struct{}

type fakeRetry struct{}

type fakeStore struct{}

type fakeWorkflow struct{}

type fakeStream struct{}

var _ ModelClient = (*fakeClient)(nil)
var _ RetryPolicy = (*fakeRetry)(nil)
var _ ArtifactStore = (*fakeStore)(nil)
var _ Workflow[int, int] = (*fakeWorkflow)(nil)

func (fakeClient) CallJSONSchema(ctx context.Context, call ModelCall) ([]byte, Tokens, error) {
	return nil, Tokens{Prompt: 1, Completion: 1, Total: 2}, nil
}

func (fakeClient) CallJSONSchemaStream(ctx context.Context, call ModelCall) (<-chan StreamChunk, Tokens, error) {
	ch := make(chan StreamChunk)
	close(ch)
	return ch, Tokens{}, nil
}

func (fakeRetry) ShouldRetry(err error, attempt int) (bool, time.Duration) {
	return false, 0
}

func (fakeStore) Put(ctx context.Context, key string, data []byte) error { return nil }

func (fakeStore) Get(ctx context.Context, key string) ([]byte, bool, error) { return nil, false, nil }

func (fakeStore) Key(workflow, step, item, inputHash string) string {
	return workflow + ":" + step + ":" + item + ":" + inputHash
}

func (fakeWorkflow) Run(ctx context.Context, input int) (int, *Trace, error) {
	return input, &Trace{}, nil
}

func (fakeWorkflow) RunStep(ctx context.Context, step string, payload any) ([]byte, *Trace, error) {
	return nil, &Trace{StepName: step, Attempts: 1}, nil
}

func TestTraceZeroValue(t *testing.T) {
	tr := Trace{}
	if tr.Attempts != 0 {
		t.Fatalf("expected zero attempts, got %d", tr.Attempts)
	}
}

func TestRetryPolicy(t *testing.T) {
	retry := fakeRetry{}
	if ok, _ := retry.ShouldRetry(nil, 0); ok {
		t.Fatal("expected no retry on fake policy")
	}
}
