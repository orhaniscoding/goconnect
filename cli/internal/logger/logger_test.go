package logger

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetup(t *testing.T) {
	t.Run("Setup Stdout Logger", func(t *testing.T) {
		err := Setup("", false)
		assert.NoError(t, err)
		
		// Verify default logger is set (not nil)
		assert.NotNil(t, slog.Default())
	})

	t.Run("Setup File Logger", func(t *testing.T) {
		tmpDir := t.TempDir()
		logPath := filepath.Join(tmpDir, "logs", "test.log")

		err := Setup(logPath, true)
		assert.NoError(t, err)

		// Verify file was created
		assert.FileExists(t, logPath)

		// Verify directory was created
		assert.DirExists(t, filepath.Dir(logPath))

		// Close the log file to allow cleanup on Windows
		Close()
	})

	t.Run("Setup Error On Invalid Path", func(t *testing.T) {
		// Try to create a log in a location where we lack permissions or path is a file
		tmpFile := filepath.Join(t.TempDir(), "file.txt")
		os.WriteFile(tmpFile, []byte("test"), 0644)
		
		err := Setup(filepath.Join(tmpFile, "app.log"), false)
		assert.Error(t, err)
	})

	t.Run("Setup Error On OpenFile", func(t *testing.T) {
		tmpDir := t.TempDir()
		// logPath points to a directory, so OpenFile should fail
		err := Setup(tmpDir, false)
		assert.Error(t, err)
	})
}

func TestLogWrappers(t *testing.T) {
	// These mainly call slog, but we run them to ensure no panics and 100% coverage
	Setup("", true)
	
	Info("test info", "key", "val")
	Warn("test warn")
	Error("test error")
	Debug("test debug")
}
