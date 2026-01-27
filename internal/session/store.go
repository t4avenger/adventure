package session

import "context"

type Store[T any] interface {
	Get(ctx context.Context, id string) (T, bool, error)
	Put(ctx context.Context, id string, v T) error
	NewID() string
}
