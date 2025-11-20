package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type WAL struct {
	file *os.File
	mu   sync.Mutex
}

type Entry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func NewWAL(filename string) (*WAL, error) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	return &WAL{file: file}, nil
}

func (w *WAL) Write(key, value string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	entry := Entry{Key: key, Value: value}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	if _, err := w.file.Write(append(data, '\n')); err != nil {
		return err
	}

	return w.file.Sync()
}

func (w *WAL) Close() error {
	return w.file.Close()
}

// ReadAll reads all entries from the WAL.
func (w *WAL) ReadAll() ([]Entry, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Seek to start
	if _, err := w.file.Seek(0, 0); err != nil {
		return nil, err
	}

	var entries []Entry
	decoder := json.NewDecoder(w.file)
	for decoder.More() {
		var entry Entry
		if err := decoder.Decode(&entry); err != nil {
			return nil, fmt.Errorf("corrupt WAL: %v", err)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}
