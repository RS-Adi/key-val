package logger

import (
	"log/slog"
	"os"
)

// Setup initializes the global logger.
// For now, it just sets up a JSON handler writing to stdout.
func Setup() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}
