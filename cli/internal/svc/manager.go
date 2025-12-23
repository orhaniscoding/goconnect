package svc

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/kardianos/service"
	"github.com/orhaniscoding/goconnect/cli/internal/system"
)

const (
	ServiceName = "goconnect-daemon"
	ServiceDesc = "GoConnect Core Daemon"
)

var (
	ErrNotAdmin       = errors.New("administrative privileges required")
	ErrAlreadyInstalled = errors.New("service is already installed")
)

type Manager struct {
	svc service.Service
}

func NewManager() (*Manager, error) {
	// Resolve the daemon executable path.
	// We expect the daemon to be in the same directory as the CLI
	// or in /usr/local/bin/goconnect-daemon (standard install).
	ex, err := os.Executable()
	if err != nil {
		return nil, err
	}
	
	dir := filepath.Dir(ex)
	daemonPath := filepath.Join(dir, "goconnect-daemon")

	// Fallback check if it doesn't exist in the same dir (e.g. if CLI is in path but daemon isn't)
	if _, err := os.Stat(daemonPath); os.IsNotExist(err) {
		daemonPath = "/usr/local/bin/goconnect-daemon"
	}

	svcConfig := &service.Config{
		Name:        ServiceName,
		DisplayName: ServiceDesc,
		Description: "Secure Headless Network Daemon",
		Executable:  daemonPath,
		Arguments:   []string{"-env", "production", "-log-path", "/var/log/goconnect/daemon.log"},
		Option: service.KeyValue{
			"Restart":               "on-failure",
			"RestartSec":            "5",
			"StartLimitIntervalSec": "60",
			"StartLimitBurst":       "5",
			"KeepAlive":             true,
			"RunAtLoad":             true,
		},
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		return nil, err
	}

	return &Manager{svc: s}, nil
}

type program struct{}

func (p *program) Start(s service.Service) error { return nil }
func (p *program) Stop(s service.Service) error { return nil }

func (m *Manager) Install() error {
	if !system.IsAdmin() {
		return ErrNotAdmin
	}
	
	// Check if already installed
	_, err := m.svc.Status()
	if err == nil {
		return ErrAlreadyInstalled
	}

	return m.svc.Install()
}

func (m *Manager) Uninstall() error {
	if !system.IsAdmin() {
		return ErrNotAdmin
	}
	return m.svc.Uninstall()
}

func (m *Manager) Start() error {
	if !system.IsAdmin() {
		return ErrNotAdmin
	}
	return m.svc.Start()
}

func (m *Manager) Stop() error {
	if !system.IsAdmin() {
		return ErrNotAdmin
	}
	return m.svc.Stop()
}

func (m *Manager) Status() (string, error) {
	status, err := m.svc.Status()
	if err != nil {
		if errors.Is(err, service.ErrNotInstalled) || strings.Contains(err.Error(), "not installed") {
			return "Not Installed", nil
		}
		return "Unknown", err
	}
	switch status {
	case service.StatusRunning:
		return "Running", nil
	case service.StatusStopped:
		return "Stopped", nil
	default:
		return "Installed (Inactive)", nil
	}
}


