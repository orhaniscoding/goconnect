package system

import (
	"fmt"
	"os/exec"
	"strings"
)

type linuxConfigurator struct{}

func newConfigurator() Configurator {
	return &linuxConfigurator{}
}

// Variables for mocking
var (
	execLookPath = exec.LookPath
)

func (c *linuxConfigurator) EnsureInterface(name string) error {
	// Check if interface exists
	if err := runCommand("ip", "link", "show", name); err == nil {
		return nil // Already exists
	}

	// Create interface
	if err := runCommand("ip", "link", "add", "dev", name, "type", "wireguard"); err != nil {
		return fmt.Errorf("failed to create wireguard interface: %w", err)
	}

	return nil
}

func (c *linuxConfigurator) ConfigureInterface(name string, addresses []string, dns []string, mtu int) error {
	// 1. Set MTU
	if err := runCommand("ip", "link", "set", "mtu", fmt.Sprintf("%d", mtu), "dev", name); err != nil {
		return fmt.Errorf("failed to set MTU: %w", err)
	}

	// 2. Set Up
	if err := runCommand("ip", "link", "set", name, "up"); err != nil {
		return fmt.Errorf("failed to set interface up: %w", err)
	}

	// 3. Flush existing IPs (optional, but good for sync)
	// runCommand("ip", "addr", "flush", "dev", name)

	// 4. Add Addresses
	for _, addr := range addresses {
		if err := runCommand("ip", "addr", "add", addr, "dev", name); err != nil {
			// Ignore "file exists" error if IP is already assigned
			if !strings.Contains(err.Error(), "File exists") {
				return fmt.Errorf("failed to add address %s: %w", addr, err)
			}
		}
	}

	// 5. Configure DNS
	if len(dns) > 0 {
		// Try using resolvconf if available
		if _, err := execLookPath("resolvconf"); err == nil {
			cmd := execCommand("resolvconf", "-a", name)
			stdin, err := cmd.StdinPipe()
			if err == nil {
				go func() {
					defer stdin.Close()
					for _, d := range dns {
						fmt.Fprintf(stdin, "nameserver %s\n", d)
					}
				}()
				if err := cmd.Run(); err != nil {
					fmt.Printf("Warning: resolvconf failed: %v\n", err)
				}
			}
		} else {
			// Fallback: maybe just log that we can't configure DNS automatically without resolvconf
			fmt.Println("Warning: resolvconf not found, DNS configuration skipped")
		}
	}

	return nil
}

func (c *linuxConfigurator) AddRoutes(name string, routes []string) error {
	for _, route := range routes {
		// ip route add <cidr> dev <name>
		if err := runCommand("ip", "route", "add", route, "dev", name); err != nil {
			// Ignore "file exists"
			if !strings.Contains(err.Error(), "File exists") {
				return fmt.Errorf("failed to add route %s: %w", route, err)
			}
		}
	}
	return nil
}

var runCommand = func(name string, args ...string) error {
	cmd := execCommand(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, string(out))
	}
	return nil
}
