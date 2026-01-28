package session

import "context"

// Store defines the interface for session storage backends.
type Store[T any] interface {
	Get(ctx context.Context, id string) (T, bool, error)
	Put(ctx context.Context, id string, v T) error
	NewID() string
}
