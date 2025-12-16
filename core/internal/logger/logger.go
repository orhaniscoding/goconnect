package logger

import (
	"log/slog"
	"os"
	"sync"
)

var (
	once   sync.Once
	logger *slog.Logger
)

// Config represents logging configuration
type Config struct {
	Environment string
	Level       string
}

// Init initializes the global logger
func Init(cfg Config) {
	once.Do(func() {
		var handler slog.Handler
		opts := &slog.HandlerOptions{
			Level: parseLevel(cfg.Level),
		}

		if cfg.Environment == "production" {
			handler = slog.NewJSONHandler(os.Stdout, opts)
		} else {
			handler = slog.NewTextHandler(os.Stdout, opts)
		}

		logger = slog.New(handler)
		slog.SetDefault(logger)
	})
}

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Get returns the global logger instance
func Get() *slog.Logger {
	if logger == nil {
		// Fallback if Init wasn't called
		return slog.Default()
	}
	return logger
}

// Helper functions for quick access
func Info(msg string, args ...any) {
	Get().Info(msg, args...)
}

func Error(msg string, args ...any) {
	Get().Error(msg, args...)
}

func Warn(msg string, args ...any) {
	Get().Warn(msg, args...)
}

func Debug(msg string, args ...any) {
	Get().Debug(msg, args...)
}
