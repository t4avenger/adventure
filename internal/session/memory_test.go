package session

import (
	"context"
	"testing"
)

func TestMemoryStore_GetPut(t *testing.T) {
	store := NewMemoryStore[string]()

	ctx := context.Background()
	id := "test-id"
	value := "test-value"

	// Test Put
	err := store.Put(ctx, id, value)
	if err != nil {
		t.Fatalf("Unexpected error on Put: %v", err)
	}

	// Test Get existing
	got, ok, err := store.Get(ctx, id)
	if err != nil {
		t.Fatalf("Unexpected error on Get: %v", err)
	}
	if !ok {
		t.Error("Expected value to exist")
	}
	if got != value {
		t.Errorf("Expected value '%s', got '%s'", value, got)
	}

	// Test Get non-existing
	_, ok, err = store.Get(ctx, "non-existent")
	if err != nil {
		t.Fatalf("Unexpected error on Get: %v", err)
	}
	if ok {
		t.Error("Expected value to not exist")
	}
}

func TestMemoryStore_Overwrite(t *testing.T) {
	store := NewMemoryStore[int]()

	ctx := context.Background()
	id := "test-id"

	// Put initial value
	err := store.Put(ctx, id, 10)
	if err != nil {
		t.Fatalf("Unexpected error on Put: %v", err)
	}

	// Overwrite with new value
	err = store.Put(ctx, id, 20)
	if err != nil {
		t.Fatalf("Unexpected error on overwrite Put: %v", err)
	}

	// Verify overwrite
	got, ok, err := store.Get(ctx, id)
	if err != nil {
		t.Fatalf("Unexpected error on Get: %v", err)
	}
	if !ok {
		t.Error("Expected value to exist")
	}
	if got != 20 {
		t.Errorf("Expected value 20, got %d", got)
	}
}

func TestMemoryStore_NewID(t *testing.T) {
	store := NewMemoryStore[string]()

	// Generate multiple IDs and ensure they're unique
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := store.NewID()
		if ids[id] {
			t.Errorf("Duplicate ID generated: %s", id)
		}
		ids[id] = true

		// IDs should be hex strings (32 chars for 16 bytes)
		if len(id) != 32 {
			t.Errorf("Expected ID length 32, got %d", len(id))
		}
	}
}

func TestMemoryStore_Concurrent(t *testing.T) {
	store := NewMemoryStore[int]()

	ctx := context.Background()

	// Test concurrent writes
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			err := store.Put(ctx, "key", id)
			if err != nil {
				t.Errorf("Error in concurrent Put: %v", err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify we can still read
	_, ok, err := store.Get(ctx, "key")
	if err != nil {
		t.Fatalf("Unexpected error on Get: %v", err)
	}
	if !ok {
		t.Error("Expected value to exist after concurrent writes")
	}
}
