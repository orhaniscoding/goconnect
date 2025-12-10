package system

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinuxConfigurator_EnsureInterface(t *testing.T) {
	// Save original and defer restore
	origRunCommand := runCommand
	defer func() { runCommand = origRunCommand }()

	c := newConfigurator()

	t.Run("AlreadyExists", func(t *testing.T) {
		runCommand = func(name string, args ...string) error {
			// Mock 'ip link show <name>' success
			if name == "ip" && args[0] == "link" && args[1] == "show" {
				return nil
			}
			return fmt.Errorf("unexpected command: %s %v", name, args)
		}

		err := c.EnsureInterface("wg0")
		assert.NoError(t, err, "EnsureInterface should succeed when interface exists")
	})

	t.Run("CreateNew", func(t *testing.T) {
		cmds := []string{}
		runCommand = func(name string, args ...string) error {
			cmdStr := fmt.Sprintf("%s %s", name, strings.Join(args, " "))
			cmds = append(cmds, cmdStr)

			if name == "ip" && args[0] == "link" && args[1] == "show" {
				return fmt.Errorf("device not found")
			}
			if name == "ip" && args[0] == "link" && args[1] == "add" {
				return nil
			}
			return fmt.Errorf("unexpected command: %s %v", name, args)
		}

		err := c.EnsureInterface("wg0")
		assert.NoError(t, err, "EnsureInterface should succeed")

		assert.Len(t, cmds, 2, "Expected 2 commands")
		assert.Contains(t, cmds[1], "ip link add dev wg0 type wireguard", "Expected creation command")
	})

	t.Run("CreateFails", func(t *testing.T) {
		runCommand = func(name string, args ...string) error {
			if name == "ip" && args[0] == "link" && args[1] == "show" {
				return fmt.Errorf("device not found")
			}
			if name == "ip" && args[0] == "link" && args[1] == "add" {
				return fmt.Errorf("permission denied")
			}
			return nil
		}

		err := c.EnsureInterface("wg0")
		assert.Error(t, err, "EnsureInterface should fail when creation fails")
		assert.Contains(t, err.Error(), "failed to create wireguard interface", "Error should mention interface creation")
	})
}

func TestLinuxConfigurator_ConfigureInterface(t *testing.T) {
	origRunCommand := runCommand
	defer func() { runCommand = origRunCommand }()

	c := newConfigurator()

	t.Run("Success", func(t *testing.T) {
		cmds := []string{}
		runCommand = func(name string, args ...string) error {
			cmdStr := fmt.Sprintf("%s %s", name, strings.Join(args, " "))
			cmds = append(cmds, cmdStr)
			return nil
		}

		addresses := []string{"10.0.0.1/24", "fc00::1/64"}
		err := c.ConfigureInterface("wg0", addresses, nil, 1420)
		require.NoError(t, err, "ConfigureInterface should succeed")

		expectedCmds := []string{
			"ip link set mtu 1420 dev wg0",
			"ip link set wg0 up",
			"ip addr add 10.0.0.1/24 dev wg0",
			"ip addr add fc00::1/64 dev wg0",
		}

		assert.Len(t, cmds, len(expectedCmds), "Expected correct number of commands")
		for i, expected := range expectedCmds {
			if i < len(cmds) {
				assert.Equal(t, expected, cmds[i], "Command mismatch at index %d", i)
			}
		}
	})

	t.Run("MTUSetFails", func(t *testing.T) {
		runCommand = func(name string, args ...string) error {
			if name == "ip" && args[0] == "link" && args[1] == "set" && args[2] == "mtu" {
				return fmt.Errorf("invalid MTU")
			}
			return nil
		}

		err := c.ConfigureInterface("wg0", []string{"10.0.0.1/24"}, nil, 1420)
		assert.Error(t, err, "Should fail when MTU set fails")
		assert.Contains(t, err.Error(), "failed to set MTU", "Error should mention MTU")
	})

	t.Run("InterfaceUpFails", func(t *testing.T) {
		runCommand = func(name string, args ...string) error {
			if name == "ip" && args[0] == "link" && args[1] == "set" && len(args) > 2 && args[2] == "wg0" {
				return fmt.Errorf("cannot bring up interface")
			}
			return nil
		}

		err := c.ConfigureInterface("wg0", []string{"10.0.0.1/24"}, nil, 1420)
		assert.Error(t, err, "Should fail when interface up fails")
		assert.Contains(t, err.Error(), "failed to set interface up", "Error should mention interface up")
	})

	t.Run("AddressFails", func(t *testing.T) {
		callCount := 0
		runCommand = func(name string, args ...string) error {
			if name == "ip" && args[0] == "addr" && args[1] == "add" {
				return fmt.Errorf("some other error")
			}
			return nil
		}
		_ = callCount

		err := c.ConfigureInterface("wg0", []string{"10.0.0.1/24"}, nil, 1420)
		assert.Error(t, err, "Should fail when address add fails")
		assert.Contains(t, err.Error(), "failed to add address", "Error should mention address")
	})

	t.Run("AddressAlreadyExists", func(t *testing.T) {
		runCommand = func(name string, args ...string) error {
			if name == "ip" && args[0] == "addr" && args[1] == "add" {
				return fmt.Errorf("File exists: some error")
			}
			return nil
		}

		err := c.ConfigureInterface("wg0", []string{"10.0.0.1/24"}, nil, 1420)
		assert.NoError(t, err, "Should ignore 'File exists' error")
	})
}

func TestLinuxConfigurator_ConfigureInterface_DNS(t *testing.T) {
	origRunCommand := runCommand
	defer func() { runCommand = origRunCommand }()
	origExecLookPath := execLookPath
	defer func() { execLookPath = origExecLookPath }()
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	c := newConfigurator()
	
	// Mock runCommand
	runCommand = func(name string, args ...string) error {
		return nil
	}

	// Mock LookPath to simulate resolvconf installed
	execLookPath = func(file string) (string, error) {
		if file == "resolvconf" {
			return "/usr/bin/resolvconf", nil
		}
		return "", fmt.Errorf("not found")
	}

	// Use TestHelperProcess pattern
	execCommand = func(name string, args ...string) *exec.Cmd {
        cs := []string{"-test.run=TestHelperProcess", "--", name}
        cs = append(cs, args...)
        cmd := exec.Command(os.Args[0], cs...)
        cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
        return cmd
    }

	addresses := []string{"10.0.0.1/24"}
	dns := []string{"8.8.8.8", "1.1.1.1"}
	
	err := c.ConfigureInterface("wg0", addresses, dns, 1420)
	assert.NoError(t, err, "ConfigureInterface should succeed with DNS")
}

func TestLinuxConfigurator_ConfigureInterface_DNSNotAvailable(t *testing.T) {
	origRunCommand := runCommand
	defer func() { runCommand = origRunCommand }()
	origExecLookPath := execLookPath
	defer func() { execLookPath = origExecLookPath }()

	c := newConfigurator()
	
	runCommand = func(name string, args ...string) error {
		return nil
	}

	// Mock LookPath to simulate resolvconf NOT installed
	execLookPath = func(file string) (string, error) {
		return "", fmt.Errorf("not found")
	}

	addresses := []string{"10.0.0.1/24"}
	dns := []string{"8.8.8.8"}
	
	// Should succeed but skip DNS configuration
	err := c.ConfigureInterface("wg0", addresses, dns, 1420)
	assert.NoError(t, err, "ConfigureInterface should succeed even without resolvconf")
}

// TestHelperProcess isn't a real test. It's used to mock exec.Command
func TestHelperProcess(t *testing.T) {
    if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
        return
    }
    
    args := os.Args
    for len(args) > 0 {
        if args[0] == "--" {
            args = args[1:]
            break
        }
        args = args[1:]
    }
    
    if len(args) == 0 {
        fmt.Fprintf(os.Stderr, "No command\n")
        os.Exit(2)
    }
    
    cmd := args[0]
    switch cmd {
    case "resolvconf":
		// Read stdin
		input, _ := io.ReadAll(os.Stdin)
		content := string(input)
		if !strings.Contains(content, "nameserver 8.8.8.8") {
			fmt.Fprintf(os.Stderr, "Expected nameserver 8.8.8.8 in stdin\n")
			os.Exit(1)
		}
		if !strings.Contains(content, "nameserver 1.1.1.1") {
			fmt.Fprintf(os.Stderr, "Expected nameserver 1.1.1.1 in stdin\n")
			os.Exit(1)
		}
        os.Exit(0)
	default:
		os.Exit(0)
    }
}

func TestLinuxConfigurator_AddRoutes(t *testing.T) {
	origRunCommand := runCommand
	defer func() { runCommand = origRunCommand }()

	c := newConfigurator()

	t.Run("Success", func(t *testing.T) {
		cmds := []string{}
		runCommand = func(name string, args ...string) error {
			cmdStr := fmt.Sprintf("%s %s", name, strings.Join(args, " "))
			cmds = append(cmds, cmdStr)
			return nil
		}

		routes := []string{"192.168.1.0/24", "10.0.0.0/8"}
		err := c.AddRoutes("wg0", routes)
		require.NoError(t, err, "AddRoutes should succeed")

		expectedCmds := []string{
			"ip route add 192.168.1.0/24 dev wg0",
			"ip route add 10.0.0.0/8 dev wg0",
		}

		assert.Len(t, cmds, 2, "Expected 2 commands")
		for i, expected := range expectedCmds {
			if i < len(cmds) {
				assert.Equal(t, expected, cmds[i], "Command mismatch at index %d", i)
			}
		}
	})

	t.Run("RouteAlreadyExists", func(t *testing.T) {
		runCommand = func(name string, args ...string) error {
			return fmt.Errorf("File exists: route already present")
		}

		routes := []string{"192.168.1.0/24"}
		err := c.AddRoutes("wg0", routes)
		assert.NoError(t, err, "Should ignore 'File exists' error")
	})

	t.Run("RouteFails", func(t *testing.T) {
		runCommand = func(name string, args ...string) error {
			return fmt.Errorf("network unreachable")
		}

		routes := []string{"192.168.1.0/24"}
		err := c.AddRoutes("wg0", routes)
		assert.Error(t, err, "Should fail on non-'File exists' error")
		assert.Contains(t, err.Error(), "failed to add route", "Error should mention route")
	})

	t.Run("EmptyRoutes", func(t *testing.T) {
		runCommand = func(name string, args ...string) error {
			t.Error("Should not be called for empty routes")
			return nil
		}

		err := c.AddRoutes("wg0", []string{})
		assert.NoError(t, err, "Empty routes should succeed")
	})

	t.Run("MultipleRoutesPartialFailure", func(t *testing.T) {
		callCount := 0
		runCommand = func(name string, args ...string) error {
			callCount++
			if callCount == 2 {
				return fmt.Errorf("network unreachable")
			}
			return nil
		}

		routes := []string{"192.168.1.0/24", "10.0.0.0/8", "172.16.0.0/12"}
		err := c.AddRoutes("wg0", routes)
		assert.Error(t, err, "Should fail on second route")
		assert.Equal(t, 2, callCount, "Should stop at first error")
	})
}

func TestNewConfigurator_ReturnsLinuxConfigurator(t *testing.T) {
	c := newConfigurator()
	assert.NotNil(t, c, "newConfigurator should return a non-nil configurator")
	
	_, ok := c.(*linuxConfigurator)
	assert.True(t, ok, "Should return a linuxConfigurator")
}

func TestLinuxConfigurator_ConfigureInterface_EmptyAddresses(t *testing.T) {
	origRunCommand := runCommand
	defer func() { runCommand = origRunCommand }()

	c := newConfigurator()

	cmdCount := 0
	runCommand = func(name string, args ...string) error {
		cmdCount++
		return nil
	}

	err := c.ConfigureInterface("wg0", []string{}, nil, 1420)
	assert.NoError(t, err, "Should succeed with empty addresses")
	// Should only call MTU and UP commands
	assert.Equal(t, 2, cmdCount, "Should only run MTU and UP commands")
}

func TestLinuxConfigurator_ConfigureInterface_LargeMTU(t *testing.T) {
	origRunCommand := runCommand
	defer func() { runCommand = origRunCommand }()

	c := newConfigurator()

	var capturedMTU string
	runCommand = func(name string, args ...string) error {
		if name == "ip" && args[0] == "link" && args[1] == "set" && args[2] == "mtu" {
			capturedMTU = args[3]
		}
		return nil
	}

	err := c.ConfigureInterface("wg0", []string{}, nil, 9000)
	assert.NoError(t, err)
	assert.Equal(t, "9000", capturedMTU, "Should use correct MTU value")
}
