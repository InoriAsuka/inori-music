package playerstate

import (
	"context"
	"sync"
)

// MemoryRepository is an in-memory player state repository for tests and development.
type MemoryRepository struct {
	mu     sync.RWMutex
	states map[string]PlayerState
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{states: make(map[string]PlayerState)}
}

func (r *MemoryRepository) Get(_ context.Context, userID string) (PlayerState, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	state, ok := r.states[userID]
	if !ok {
		return PlayerState{}, ErrNotFound
	}
	return cloneState(state), nil
}

func (r *MemoryRepository) Put(_ context.Context, userID string, state PlayerState) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.states[userID] = cloneState(state)
	return nil
}

// cloneState returns a copy of state with its queue slice defensively copied,
// so stored state cannot be mutated through a caller-held slice reference.
func cloneState(state PlayerState) PlayerState {
	queue := make([]string, len(state.Queue))
	copy(queue, state.Queue)
	state.Queue = queue
	return state
}
