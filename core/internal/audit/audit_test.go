package audit

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStdoutAuditor(t *testing.T) {
	auditor := NewStdoutAuditor()
	assert.NotNil(t, auditor)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	ctx := context.Background()
	auditor.Event(ctx, "tenant-1", "user.login", "user-1", "session-1", nil)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, `"action":"user.login"`)
	assert.Contains(t, output, `"tenant_id":"tenant-1"`)
	assert.Contains(t, output, `"actor":"[redacted]"`)
	assert.Contains(t, output, `"object":"[redacted]"`)
}

func TestNewStdoutAuditorWithHashing(t *testing.T) {
	t.Run("With secret", func(t *testing.T) {
		secret := []byte("test-secret-key-1234567890")
		auditor := NewStdoutAuditorWithHashing(secret)
		assert.NotNil(t, auditor)

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		ctx := context.Background()
		auditor.Event(ctx, "tenant-1", "user.login", "user-1", "session-1", map[string]any{"ip": "192.168.1.1"})

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		assert.Contains(t, output, `"action":"user.login"`)
		// Actor and object should be hashed, not [redacted]
		assert.NotContains(t, output, `"actor":"[redacted]"`)
		assert.NotContains(t, output, `"object":"[redacted]"`)
	})

	t.Run("With empty secret", func(t *testing.T) {
		auditor := NewStdoutAuditorWithHashing([]byte{})
		assert.NotNil(t, auditor)
	})
}

func TestNewStdoutAuditorWithHashSecrets(t *testing.T) {
	t.Run("With secrets", func(t *testing.T) {
		secrets := [][]byte{
			[]byte("primary-secret"),
			[]byte("secondary-secret"),
		}
		auditor := NewStdoutAuditorWithHashSecrets(secrets...)
		assert.NotNil(t, auditor)

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		ctx := context.Background()
		auditor.Event(ctx, "tenant-1", "user.logout", "user-2", "session-2", nil)

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		assert.Contains(t, output, `"action":"user.logout"`)
	})

	t.Run("With no secrets", func(t *testing.T) {
		auditor := NewStdoutAuditorWithHashSecrets()
		assert.NotNil(t, auditor)
	})

	t.Run("With empty first secret", func(t *testing.T) {
		auditor := NewStdoutAuditorWithHashSecrets([]byte{})
		assert.NotNil(t, auditor)
	})
}

func TestWrapWithMetrics(t *testing.T) {
	baseAuditor := NewStdoutAuditor()

	callCount := 0
	lastAction := ""
	inc := func(action string) {
		callCount++
		lastAction = action
	}

	metricsAuditor := WrapWithMetrics(baseAuditor, inc)
	assert.NotNil(t, metricsAuditor)

	// Capture stdout to suppress output
	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	ctx := context.Background()
	metricsAuditor.Event(ctx, "tenant-1", "device.register", "user-1", "device-1", nil)
	metricsAuditor.Event(ctx, "tenant-1", "device.delete", "user-1", "device-2", nil)

	w.Close()
	os.Stdout = old

	assert.Equal(t, 2, callCount)
	assert.Equal(t, "device.delete", lastAction)
}

func TestWrapWithMetrics_NilInc(t *testing.T) {
	baseAuditor := NewStdoutAuditor()
	metricsAuditor := WrapWithMetrics(baseAuditor, nil)
	assert.NotNil(t, metricsAuditor)

	// Should not panic with nil inc function
	// Capture stdout to suppress output
	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	ctx := context.Background()
	metricsAuditor.Event(ctx, "tenant-1", "test.action", "actor", "object", nil)

	w.Close()
	os.Stdout = old
}
