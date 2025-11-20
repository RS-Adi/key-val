package store

import (
	"testing"
)

func TestInMemoryStore(t *testing.T) {
	s := NewInMemoryStore(nil)

	// Test Set and Get
	key := "foo"
	val := "bar"
	if err := s.Set(key, val); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	got, err := s.Get(key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got != val {
		t.Errorf("Get = %q, want %q", got, val)
	}

	// Test Get non-existent
	_, err = s.Get("missing")
	if err != ErrKeyNotFound {
		t.Errorf("Get missing = %v, want %v", err, ErrKeyNotFound)
	}

	// Test Delete
	if err := s.Delete(key); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = s.Get(key)
	if err != ErrKeyNotFound {
		t.Errorf("Get deleted = %v, want %v", err, ErrKeyNotFound)
	}
}
