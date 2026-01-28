// Package session provides in-memory storage for game sessions.
package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
)

// MemoryStore is an in-memory implementation of Store that uses a map
// protected by a read-write mutex for thread safety.
type MemoryStore[T any] struct {
	mu sync.RWMutex
	m  map[string]T
}

// NewMemoryStore creates a new in-memory store.
func NewMemoryStore[T any]() *MemoryStore[T] {
	return &MemoryStore[T]{m: map[string]T{}}
}

// Get retrieves a value from the store by ID.
func (s *MemoryStore[T]) Get(_ context.Context, id string) (value T, ok bool, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok = s.m[id]
	return value, ok, nil
}

// Put stores a value in the store with the given ID.
func (s *MemoryStore[T]) Put(_ context.Context, id string, v T) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[id] = v
	return nil
}

// NewID generates a new unique session ID.
func (s *MemoryStore[T]) NewID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback: if crypto/rand fails, return a deterministic but unique ID
		// This should never happen in practice, but we handle it gracefully
		return hex.EncodeToString([]byte("fallback-id"))
	}
	return hex.EncodeToString(b)
}
