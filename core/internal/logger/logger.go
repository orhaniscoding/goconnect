package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	once   sync.Once
	logger *slog.Logger
)

// Config represents logging configuration
type Config struct {
	Environment string
	Level       string
	LogPath     string
	MaxSize     int  // Megabytes
	MaxBackups  int
	MaxAge      int  // Days
	Compress    bool
}

// Setup initializes the global logger with the provided configuration
func Setup(cfg Config) {
	once.Do(func() {
		var writers []io.Writer
		
		// Always log to Stdout for container visibility or dev debugging
		writers = append(writers, os.Stdout)

		if cfg.LogPath != "" {
			// Ensure directory exists with restrictive permissions
			dir := filepath.Dir(cfg.LogPath)
			if err := os.MkdirAll(dir, 0700); err != nil {
				slog.Error("failed to create log directory", "path", dir, "error", err)
			} else {
				lumberjackLogger := &lumberjack.Logger{
					Filename:   cfg.LogPath,
					MaxSize:    cfg.MaxSize,
					MaxBackups: cfg.MaxBackups,
					MaxAge:     cfg.MaxAge,
					Compress:   cfg.Compress,
				}
				writers = append(writers, lumberjackLogger)
			}
		}

		multiWriter := io.MultiWriter(writers...)

		opts := &slog.HandlerOptions{
			Level: parseLevel(cfg.Level),
		}

		var handler slog.Handler
		if strings.ToLower(cfg.Environment) == "production" {
			handler = slog.NewJSONHandler(multiWriter, opts)
		} else {
			handler = slog.NewTextHandler(multiWriter, opts)
		}

		logger = slog.New(handler)
		slog.SetDefault(logger)
	})
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
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

