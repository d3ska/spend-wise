package auth

import (
	"sync"
	"time"
)

// StateStore provides in-memory storage for OAuth state parameters with automatic
// expiry. Each state token can only be consumed once (single-use).
type StateStore struct {
	mu     sync.Mutex
	states map[string]time.Time
	ttl    time.Duration
}

// NewStateStore creates a StateStore with the given TTL for state tokens.
func NewStateStore(ttl time.Duration) *StateStore {
	return &StateStore{
		states: make(map[string]time.Time),
		ttl:    ttl,
	}
}

// Store records a state token. It will expire after the configured TTL.
func (s *StateStore) Store(state string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Opportunistic cleanup of expired entries.
	now := time.Now()
	for k, exp := range s.states {
		if now.After(exp) {
			delete(s.states, k)
		}
	}

	s.states[state] = now.Add(s.ttl)
}

// Validate checks whether the state token exists and has not expired.
// If valid, the token is consumed (deleted) to prevent replay.
// Returns true if the state was valid, false otherwise.
func (s *StateStore) Validate(state string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	exp, ok := s.states[state]
	if !ok {
		return false
	}

	// Always delete the token (single-use).
	delete(s.states, state)

	return time.Now().Before(exp)
}
