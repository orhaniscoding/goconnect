package system

// ProtocolHandler handles OS-specific custom protocol registration (goconnect://)
type ProtocolHandler interface {
	// Register registers the custom protocol handler with the OS
	Register(scheme string, command string) error
	// Unregister removes the custom protocol handler
	Unregister(scheme string) error
}

// NewProtocolHandler creates a new ProtocolHandler for the current OS
func NewProtocolHandler() ProtocolHandler {
	return newProtocolHandler()
}
