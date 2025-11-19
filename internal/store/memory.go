package store

import (
	"sync"
)

// InMemoryStore is a thread-safe in-memory key-value store.
type InMemoryStore struct {
	mu   sync.RWMutex
	data map[string]string
}

// NewInMemoryStore creates a new instance of InMemoryStore.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		data: make(map[string]string),
	}
}

// Set stores a value for a given key.
func (s *InMemoryStore) Set(key string, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	return nil
}

// Get retrieves a value for a given key.
func (s *InMemoryStore) Get(key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, ok := s.data[key]
	if !ok {
		return "", ErrKeyNotFound
	}
	return val, nil
}

// Delete removes a key from the store.
func (s *InMemoryStore) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	return nil
}
