//go:build windows

package system

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows/registry"
)

type windowsProtocolHandler struct{}

func newProtocolHandler() ProtocolHandler {
	return &windowsProtocolHandler{}
}

func (h *windowsProtocolHandler) Register(scheme string, command string) error {
	// HKCR\goconnect
	key, _, err := registry.CreateKey(registry.CLASSES_ROOT, scheme, registry.ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("failed to create registry key: %w", err)
	}
	defer key.Close()

	if err := key.SetStringValue("", fmt.Sprintf("URL:%s Protocol", scheme)); err != nil {
		return fmt.Errorf("failed to set default value: %w", err)
	}
	if err := key.SetStringValue("URL Protocol", ""); err != nil {
		return fmt.Errorf("failed to set URL Protocol value: %w", err)
	}

	// HKCR\goconnect\DefaultIcon
	iconKey, _, err := registry.CreateKey(key, "DefaultIcon", registry.ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("failed to create DefaultIcon key: %w", err)
	}
	defer iconKey.Close()

	// Use the executable itself as the icon
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	if err := iconKey.SetStringValue("", fmt.Sprintf("%s,1", exePath)); err != nil {
		return fmt.Errorf("failed to set DefaultIcon value: %w", err)
	}

	// HKCR\goconnect\shell\open\command
	cmdKey, _, err := registry.CreateKey(key, `shell\open\command`, registry.ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("failed to create command key: %w", err)
	}
	defer cmdKey.Close()

	// Command: "path\to\exe" "%1"
	// If command arg is provided, use it, otherwise use current executable
	cmdStr := command
	if cmdStr == "" {
		cmdStr = fmt.Sprintf("\"%s\" \"%%1\"", exePath)
	}

	if err := cmdKey.SetStringValue("", cmdStr); err != nil {
		return fmt.Errorf("failed to set command value: %w", err)
	}

	return nil
}

func (h *windowsProtocolHandler) Unregister(scheme string) error {
	return registry.DeleteKey(registry.CLASSES_ROOT, scheme)
}
