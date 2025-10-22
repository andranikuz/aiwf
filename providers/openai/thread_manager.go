package openai

import (
	"context"
	"fmt"
	"sync"

	"github.com/andranikuz/aiwf/runtime/go/aiwf"
)

// InMemoryThreadManager - простая реализация ThreadManager в памяти
// Для production используйте OpenAI Threads API или другое решение
type InMemoryThreadManager struct {
	threads map[string]*ThreadData
	mu      sync.RWMutex
}

// ThreadData хранит информацию о треде
type ThreadData struct {
	ID       string
	Messages []string
	Metadata map[string]any
}

// NewInMemoryThreadManager создаёт новый менеджер тредов в памяти
func NewInMemoryThreadManager() *InMemoryThreadManager {
	return &InMemoryThreadManager{
		threads: make(map[string]*ThreadData),
	}
}

// Start создаёт новый тред
func (m *InMemoryThreadManager) Start(ctx context.Context, assistant string, binding aiwf.ThreadBinding) (*aiwf.ThreadState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	threadID := fmt.Sprintf("thread_%d", len(m.threads))
	thread := &ThreadData{
		ID:       threadID,
		Messages: []string{},
		Metadata: make(map[string]any),
	}

	m.threads[threadID] = thread

	return &aiwf.ThreadState{
		ID:       threadID,
		Metadata: thread.Metadata,
	}, nil
}

// Continue добавляет сообщение в тред
func (m *InMemoryThreadManager) Continue(ctx context.Context, state *aiwf.ThreadState, feedback string) error {
	if state == nil {
		return fmt.Errorf("thread state is nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	thread, ok := m.threads[state.ID]
	if !ok {
		return fmt.Errorf("thread %s not found", state.ID)
	}

	thread.Messages = append(thread.Messages, feedback)
	return nil
}

// Close закрывает тред
func (m *InMemoryThreadManager) Close(ctx context.Context, state *aiwf.ThreadState) error {
	if state == nil {
		return fmt.Errorf("thread state is nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if thread exists before closing
	if _, ok := m.threads[state.ID]; !ok {
		return fmt.Errorf("thread %s not found", state.ID)
	}

	delete(m.threads, state.ID)
	return nil
}

// GetThread возвращает информацию о треде (для отладки)
func (m *InMemoryThreadManager) GetThread(threadID string) (*ThreadData, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	thread, ok := m.threads[threadID]
	if !ok {
		return nil, fmt.Errorf("thread %s not found", threadID)
	}

	return thread, nil
}
