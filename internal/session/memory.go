package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
)

type MemoryStore[T any] struct {
	mu sync.RWMutex
	m  map[string]T
}

func NewMemoryStore[T any]() *MemoryStore[T] {
	return &MemoryStore[T]{m: map[string]T{}}
}

func (s *MemoryStore[T]) Get(_ context.Context, id string) (T, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.m[id]
	return v, ok, nil
}

func (s *MemoryStore[T]) Put(_ context.Context, id string, v T) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[id] = v
	return nil
}

func (s *MemoryStore[T]) NewID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
