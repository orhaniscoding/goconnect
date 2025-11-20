package system

// Configurator handles OS-specific network configuration
type Configurator interface {
	// EnsureInterface ensures the interface exists (creates it if necessary on supported platforms)
	EnsureInterface(name string) error
	// ConfigureInterface configures the network interface with the given addresses and MTU
	ConfigureInterface(name string, addresses []string, dns []string, mtu int) error
	// AddRoutes adds routes to the interface
	AddRoutes(name string, routes []string) error
}

// NewConfigurator returns a platform-specific configurator
func NewConfigurator() Configurator {
	return newConfigurator()
}
