package main

import (
	"log/slog"
	"os"

	"distributed-kv/internal/config"
	"distributed-kv/internal/server"
	"distributed-kv/internal/store"
	"distributed-kv/pkg/logger"
)

func main() {
	logger.Setup()
	cfg := config.Load()

	slog.Info("Starting Distributed Key-Value Store", "port", cfg.Port)

	// Initialize store
	kvStore := store.NewInMemoryStore()

	// Start HTTP server
	server := server.NewHTTPServer(cfg.Port, kvStore)
	if err := server.Start(); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}
