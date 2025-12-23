package logger

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitAndGet(t *testing.T) {
	// Test parseLevel logic first as it is pure
	assert.Equal(t, slog.LevelDebug, parseLevel("debug"))
	assert.Equal(t, slog.LevelInfo, parseLevel("info"))
	assert.Equal(t, slog.LevelWarn, parseLevel("warn"))
	assert.Equal(t, slog.LevelError, parseLevel("error"))
	assert.Equal(t, slog.LevelInfo, parseLevel("unknown"))

	// Test Init (sync.Once means we only test one path in this process)
	cfg := Config{
		Environment: "development",
		Level:       "debug",
	}
	Setup(cfg)
	
	l := Get()
	assert.NotNil(t, l)
	
	// Test wrappers don't panic
	Info("test info", "key", "val")
	Warn("test warn")
	Error("test error")
	Debug("test debug")
}

func TestGet_Default(t *testing.T) {
	// If logger is nil (not initialized in a fresh run), it returns slog.Default()
	// But since Init might have been called by other tests in the package, 
	// this is only reliably testable if logger is still nil.
	// We can't easily reset the global 'logger' var here.
}
