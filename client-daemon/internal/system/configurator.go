package system

// Configurator handles OS-specific network configuration
type Configurator interface {
	// ConfigureInterface configures the network interface with the given addresses and MTU
	ConfigureInterface(name string, addresses []string, dns []string, mtu int) error
}

// NewConfigurator returns a platform-specific configurator
func NewConfigurator() Configurator {
	return newConfigurator()
}
