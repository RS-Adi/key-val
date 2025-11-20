package main

import (
	"fmt"
	"log/slog"
	"os"

	"distributed-kv/internal/config"
	"distributed-kv/internal/server"
	"distributed-kv/internal/storage"
	"distributed-kv/internal/store"
	"distributed-kv/pkg/logger"
)

func main() {
	logger.Setup()
	cfg := config.Load()

	slog.Info("Starting Distributed Key-Value Store", "port", cfg.Port)

	// Initialize WAL
	walFile := fmt.Sprintf("wal-%d.log", cfg.Port)
	wal, err := storage.NewWAL(walFile)
	if err != nil {
		slog.Error("Failed to create WAL", "error", err)
		os.Exit(1)
	}
	defer wal.Close()

	// Initialize store
	kvStore := store.NewInMemoryStore(wal)
	if err := kvStore.Recover(); err != nil {
		slog.Error("Failed to recover from WAL", "error", err)
		os.Exit(1)
	}

	// Start HTTP server
	server := server.NewHTTPServer(cfg.Port, kvStore)
	if err := server.Start(); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}
