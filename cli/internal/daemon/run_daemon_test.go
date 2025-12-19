package daemon

import (
	"testing"

	"github.com/kardianos/service"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/config"
	"github.com/stretchr/testify/assert"
)

type mockService struct {
	service.Service
	installed   bool
	uninstalled bool
	started     bool
	stopped     bool
	runCalled   bool
}

func (m *mockService) Install() error                                   { m.installed = true; return nil }
func (m *mockService) Uninstall() error                                 { m.uninstalled = true; return nil }
func (m *mockService) Start() error                                     { m.started = true; return nil }
func (m *mockService) Stop() error                                      { m.stopped = true; return nil }
func (m *mockService) Run() error                                       { m.runCalled = true; return nil }
func (m *mockService) Logger(errs chan<- error) (service.Logger, error) { return &mockSvcLogger{}, nil }

type mockSvcLogger struct{}

func (l *mockSvcLogger) Info(v ...interface{}) error                 { return nil }
func (l *mockSvcLogger) Infof(format string, v ...interface{}) error  { return nil }
func (l *mockSvcLogger) Warning(v ...interface{}) error              { return nil }
func (l *mockSvcLogger) Warningf(format string, v ...interface{}) error { return nil }
func (l *mockSvcLogger) Error(v ...interface{}) error                { return nil }
func (l *mockSvcLogger) Errorf(format string, v ...interface{}) error { return nil }

func TestRunDaemonInternal_Commands(t *testing.T) {
	cfg := &config.Config{}
	version := "1.0.0"

	tests := []struct {
		name     string
		args     []string
		verify   func(t *testing.T, m *mockService)
		expected bool // expect error
	}{
		{
			name: "install",
			args: []string{"goconnect", "install"},
			verify: func(t *testing.T, m *mockService) {
				assert.True(t, m.installed)
			},
		},
		{
			name: "uninstall",
			args: []string{"goconnect", "uninstall"},
			verify: func(t *testing.T, m *mockService) {
				assert.True(t, m.uninstalled)
			},
		},
		{
			name: "start",
			args: []string{"goconnect", "start"},
			verify: func(t *testing.T, m *mockService) {
				assert.True(t, m.started)
			},
		},
		{
			name: "stop",
			args: []string{"goconnect", "stop"},
			verify: func(t *testing.T, m *mockService) {
				assert.True(t, m.stopped)
			},
		},
		{
			name: "run (default)",
			args: []string{"goconnect"},
			verify: func(t *testing.T, m *mockService) {
				assert.True(t, m.runCalled)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockService{}
			factory := func(i service.Interface, c *service.Config) (service.Service, error) {
				return m, nil
			}

			err := runDaemonInternal(cfg, version, nil, factory, tt.args)
			if tt.expected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			tt.verify(t, m)
		})
	}
}

func TestRunDaemonInternal_FactoryError(t *testing.T) {
	cfg := &config.Config{}
	factory := func(i service.Interface, c *service.Config) (service.Service, error) {
		return nil, assert.AnError
	}

	err := runDaemonInternal(cfg, "1.0.0", nil, factory, []string{"goconnect"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create service")
}
