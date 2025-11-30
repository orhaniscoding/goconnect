package system

import (
	"fmt"
	"os/exec"
)

type darwinConfigurator struct{}

func newConfigurator() Configurator {
	return &darwinConfigurator{}
}

func (c *darwinConfigurator) ConfigureInterface(name string, addresses []string, dns []string, mtu int) error {
	// macOS implementation using ifconfig
	// ifconfig <name> <ip> <dest_ip>
	// macOS usually requires point-to-point addressing for tun devices or just alias

	for _, addr := range addresses {
		// ifconfig utunX inet 10.0.0.2/24 10.0.0.2
		if err := runCommand("ifconfig", name, "inet", addr, addr); err != nil {
			return fmt.Errorf("failed to add address %s: %w", addr, err)
		}
	}

	// Set MTU
	if err := runCommand("ifconfig", name, "mtu", fmt.Sprintf("%d", mtu)); err != nil {
		return fmt.Errorf("failed to set MTU: %w", err)
	}

	return nil
}

func (c *darwinConfigurator) EnsureInterface(name string) error {
	// On macOS, the interface is typically created by the wireguard-go userspace implementation.
	// We just check if it exists.
	if err := runCommand("ifconfig", name); err != nil {
		return fmt.Errorf("interface %s does not exist (it should be created by the tunnel): %w", name, err)
	}
	return nil
}

func (c *darwinConfigurator) AddRoutes(name string, routes []string) error {
	for _, route := range routes {
		// route -n add -net <cidr> -interface <name>
		// Example: route -n add -net 10.0.0.0/24 -interface utun1
		if err := runCommand("route", "-n", "add", "-net", route, "-interface", name); err != nil {
			return fmt.Errorf("failed to add route %s: %w", route, err)
		}
	}
	return nil
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, string(out))
	}
	return nil
}
