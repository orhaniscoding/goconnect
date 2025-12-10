package system

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
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
		if err != nil {
			t.Errorf("EnsureInterface failed: %v", err)
		}
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
		if err != nil {
			t.Errorf("EnsureInterface failed: %v", err)
		}

		if len(cmds) != 2 {
			t.Errorf("Expected 2 commands, got %d", len(cmds))
		}
		if !strings.Contains(cmds[1], "ip link add dev wg0 type wireguard") {
			t.Errorf("Expected creation command, got: %s", cmds[1])
		}
	})
}

func TestLinuxConfigurator_ConfigureInterface(t *testing.T) {
	origRunCommand := runCommand
	defer func() { runCommand = origRunCommand }()

	c := newConfigurator()
	cmds := []string{}

	runCommand = func(name string, args ...string) error {
		cmdStr := fmt.Sprintf("%s %s", name, strings.Join(args, " "))
		cmds = append(cmds, cmdStr)
		return nil
	}

	addresses := []string{"10.0.0.1/24", "fc00::1/64"}
	err := c.ConfigureInterface("wg0", addresses, nil, 1420)
	if err != nil {
		t.Fatalf("ConfigureInterface failed: %v", err)
	}

	expectedCmds := []string{
		"ip link set mtu 1420 dev wg0",
		"ip link set wg0 up",
		"ip addr add 10.0.0.1/24 dev wg0",
		"ip addr add fc00::1/64 dev wg0",
	}

	if len(cmds) != len(expectedCmds) {
		t.Errorf("Expected %d commands, got %d", len(expectedCmds), len(cmds))
		for i, cmd := range cmds {
			t.Logf("Cmd %d: %s", i, cmd)
		}
	}
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

	// Mock execCommand to capture stdin
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "resolvconf" {
			if args[0] == "-a" && args[1] == "wg0" {
				return exec.Command("sh", "-c", "cat > /dev/null") 
			}
		}
		return exec.Command("false")
	}

	// Wait, we need to verify what was written to stdin.
	// We can't do that easily with a real exec.Command unless we use the TestHelperProcess trick.
	// But let's try a different approach:
	// Since we are mocking execCommand variable, we can return a structural *exec.Cmd? No, it's a struct.
	// We have to use the helper process.
	
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
	if err != nil {
		t.Fatalf("ConfigureInterface failed: %v", err)
	}

	// Since verifying stdin content with helper process is complex (needs piping back to parent or file),
	// we will trust the helper process validation if we can implement it.
	// For now, let's just assert no error occurred, meaning resolvconf was called.
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
    
    cmd, args := args[0], args[1:]
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
	cmds := []string{}

	runCommand = func(name string, args ...string) error {
		cmdStr := fmt.Sprintf("%s %s", name, strings.Join(args, " "))
		cmds = append(cmds, cmdStr)
		return nil
	}

	routes := []string{"192.168.1.0/24", "10.0.0.0/8"}
	err := c.AddRoutes("wg0", routes)
	if err != nil {
		t.Fatalf("AddRoutes failed: %v", err)
	}

	expectedCmds := []string{
		"ip route add 192.168.1.0/24 dev wg0",
		"ip route add 10.0.0.0/8 dev wg0",
	}

	if len(cmds) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(cmds))
	}
	
	for i, expected := range expectedCmds {
		if i < len(cmds) && cmds[i] != expected {
			t.Errorf("Cmd %d mismatch.\nExpected: %s\nGot:      %s", i, expected, cmds[i])
		}
	}
}
