//go:build darwin

package system

import (
	"fmt"
)

type darwinProtocolHandler struct{}

func newProtocolHandler() ProtocolHandler {
	return &darwinProtocolHandler{}
}

func (h *darwinProtocolHandler) Register(scheme string, command string) error {
	// TODO: On macOS, URL handlers are typically defined in the Info.plist of an .app bundle.
	// For a standalone binary/service, registration is non-trivial without creating a shim .app.
	// We will skip this for now or implement a shim later.
	return fmt.Errorf("protocol registration on macOS requires an Application Bundle")
}

func (h *darwinProtocolHandler) Unregister(scheme string) error {
	return nil
}
