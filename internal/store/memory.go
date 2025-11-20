package store

import (
	"distributed-kv/internal/storage"
	"sync"
)

// InMemoryStore is a thread-safe in-memory key-value store.
type InMemoryStore struct {
	mu   sync.RWMutex
	data map[string]string
	wal  *storage.WAL
}

// NewInMemoryStore creates a new instance of InMemoryStore.
func NewInMemoryStore(wal *storage.WAL) *InMemoryStore {
	return &InMemoryStore{
		data: make(map[string]string),
		wal:  wal,
	}
}

// Recover restores the store state from the WAL.
func (s *InMemoryStore) Recover() error {
	if s.wal == nil {
		return nil
	}

	entries, err := s.wal.ReadAll()
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, entry := range entries {
		s.data[entry.Key] = entry.Value
	}
	return nil
}

// Set stores a value for a given key.
func (s *InMemoryStore) Set(key string, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Write to WAL first
	if s.wal != nil {
		if err := s.wal.Write(key, value); err != nil {
			return err
		}
	}

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
