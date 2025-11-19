package store

import "errors"

var (
	ErrKeyNotFound = errors.New("key not found")
)

// Store defines the interface for a key-value store.
type Store interface {
	// Set stores a value for a given key.
	Set(key string, value string) error

	// Get retrieves a value for a given key.
	Get(key string) (string, error)

	// Delete removes a key from the store.
	Delete(key string) error
}
