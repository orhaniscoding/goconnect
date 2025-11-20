package system

import (
	"fmt"
	"os/exec"
	"strings"
)

type windowsConfigurator struct{}

func newConfigurator() Configurator {
	return &windowsConfigurator{}
}

func (c *windowsConfigurator) EnsureInterface(name string) error {
	// On Windows, creating the interface usually requires the WireGuard driver/service.
	// For now, we assume it exists or is managed by the user/installer.
	// We could check existence via Get-NetAdapter.
	return nil
}

func (c *windowsConfigurator) ConfigureInterface(name string, addresses []string, dns []string, mtu int) error {
	// Windows requires the interface to exist.
	// Using netsh for compatibility.

	// 1. Set Addresses
	// Note: Windows handles CIDR in "addr/mask" format differently in netsh, usually requires mask.
	// PowerShell New-NetIPAddress is better but slower.
	// Let's try netsh.
	// netsh interface ip set address "name" static <ip> <mask>

	// For simplicity in this prototype, we'll assume PowerShell is available (since we are on Windows 10+ usually)

	for _, addr := range addresses {
		// Parse CIDR
		parts := strings.Split(addr, "/")
		if len(parts) != 2 {
			continue
		}
		ip := parts[0]
		prefix := parts[1]

		// Remove existing (ignore error)
		// Remove-NetIPAddress -InterfaceAlias "name" -Confirm:$false

		// Add new
		// New-NetIPAddress -InterfaceAlias "name" -IPAddress <ip> -PrefixLength <prefix>

		cmd := fmt.Sprintf("New-NetIPAddress -InterfaceAlias \"%s\" -IPAddress %s -PrefixLength %s -PolicyStore ActiveStore -Confirm:$false", name, ip, prefix)
		// Ignore errors for now (e.g. if already exists)
		_ = runPowerShell(cmd)
	}

	// 2. Set DNS
	if len(dns) > 0 {
		dnsStr := strings.Join(dns, ",")
		// Set-DnsClientServerAddress -InterfaceAlias "name" -ServerAddresses <dns>
		cmd := fmt.Sprintf("Set-DnsClientServerAddress -InterfaceAlias \"%s\" -ServerAddresses %s", name, dnsStr)
		if err := runPowerShell(cmd); err != nil {
			return fmt.Errorf("failed to set DNS: %w", err)
		}
	}

	return nil
}

func (c *windowsConfigurator) AddRoutes(name string, routes []string) error {
	for _, route := range routes {
		// New-NetRoute -DestinationPrefix <cidr> -InterfaceAlias <name> -PolicyStore ActiveStore
		// Ignore errors if exists
		cmd := fmt.Sprintf("New-NetRoute -DestinationPrefix %s -InterfaceAlias \"%s\" -PolicyStore ActiveStore -Confirm:$false -ErrorAction SilentlyContinue", route, name)
		_ = runPowerShell(cmd)
	}
	return nil
}

func runPowerShell(command string) error {
	cmd := exec.Command("powershell", "-Command", command)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, string(out))
	}
	return nil
}
