//go:build !windows && !linux && !darwin

package system

import (
	"fmt"
)

type otherProtocolHandler struct{}

func newProtocolHandler() ProtocolHandler {
	return &otherProtocolHandler{}
}

func (h *otherProtocolHandler) Register(scheme string, command string) error {
	return fmt.Errorf("protocol registration not supported on this OS")
}

func (h *otherProtocolHandler) Unregister(scheme string) error {
	return nil
}
