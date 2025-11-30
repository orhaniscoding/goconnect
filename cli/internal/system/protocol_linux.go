//go:build linux

package system

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type linuxProtocolHandler struct{}

func newProtocolHandler() ProtocolHandler {
	return &linuxProtocolHandler{}
}

func (h *linuxProtocolHandler) Register(scheme string, command string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	desktopPath := filepath.Join(home, ".local", "share", "applications")
	if err := os.MkdirAll(desktopPath, 0755); err != nil {
		return err
	}

	appName := fmt.Sprintf("%s.desktop", scheme)
	desktopFile := filepath.Join(desktopPath, appName)

	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	cmdStr := command
	if cmdStr == "" {
		cmdStr = fmt.Sprintf("%s %%u", exePath)
	}

	content := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=GoConnect
Exec=%s
StartupNotify=false
MimeType=x-scheme-handler/%s;
`, cmdStr, scheme)

	if err := os.WriteFile(desktopFile, []byte(content), 0644); err != nil {
		return err
	}

	// Update mime database
	if err := exec.Command("xdg-mime", "default", appName, fmt.Sprintf("x-scheme-handler/%s", scheme)).Run(); err != nil {
		return fmt.Errorf("failed to update mime database: %w", err)
	}

	return nil
}

func (h *linuxProtocolHandler) Unregister(scheme string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	desktopFile := filepath.Join(home, ".local", "share", "applications", fmt.Sprintf("%s.desktop", scheme))
	return os.Remove(desktopFile)
}
