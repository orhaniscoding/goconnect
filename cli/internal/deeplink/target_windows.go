//go:build windows

package deeplink

// getDefaultGRPCTarget returns the default gRPC target for Windows
func getDefaultGRPCTarget() string {
	return "127.0.0.1:34101"
}
