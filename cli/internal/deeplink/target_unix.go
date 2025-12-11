//go:build !windows

package deeplink

// getDefaultGRPCTarget returns the default gRPC target for Unix-like systems
func getDefaultGRPCTarget() string {
	return "unix:///tmp/goconnect-daemon.sock"
}
