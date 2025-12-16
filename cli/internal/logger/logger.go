package logger

import (
	"log/slog"
	"os"
	"path/filepath"
)

// Setup configures the default slog logger.
// If logPath is provided, it logs to that file (JSON).
// Otherwise it logs to Stdout (Text).
func Setup(logPath string, debug bool) error {
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	if debug {
		opts.Level = slog.LevelDebug
	}

	if logPath != "" {
		// Ensure directory exists
		dir := filepath.Dir(logPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}

		f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		handler = slog.NewJSONHandler(f, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return nil
}

// Info logs an info-level message with optional key-value pairs.
func Info(msg string, args ...any) {
	slog.Info(msg, args...)
}

// Warn logs a warning-level message with optional key-value pairs.
func Warn(msg string, args ...any) {
	slog.Warn(msg, args...)
}

// Error logs an error-level message with optional key-value pairs.
func Error(msg string, args ...any) {
	slog.Error(msg, args...)
}

// Debug logs a debug-level message with optional key-value pairs.
func Debug(msg string, args ...any) {
	slog.Debug(msg, args...)
}
