package system

import (
	"os"
	"os/exec"
	"testing"
)

func TestGetOSVersion_Linux(t *testing.T) {
	// Only valid on Linux runtime for GetOSVersion call
	// But we can test getLinuxVersion directly everywhere since it's in info.go without build constraints
	
	// Save/Restore
	origPath := osReleasePath
	defer func() { osReleasePath = origPath }()

	t.Run("PrettyName", func(t *testing.T) {
		start := "PRETTY_NAME=\"Ubuntu 22.04 LTS\"\nNAME=\"Ubuntu\"\n"
		f, err := os.CreateTemp("", "os-release")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(f.Name())
		f.WriteString(start)
		f.Close()

		osReleasePath = f.Name()
		
		version := getLinuxVersion()
		if version != "Ubuntu 22.04 LTS" {
			t.Errorf("Expected 'Ubuntu 22.04 LTS', got '%s'", version)
		}
	})

	t.Run("NameAndVersion", func(t *testing.T) {
		content := "NAME=\"Fedora Linux\"\nVERSION_ID=\"38\"\n"
		f, err := os.CreateTemp("", "os-release")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(f.Name())
		f.WriteString(content)
		f.Close()

		osReleasePath = f.Name()

		version := getLinuxVersion()
		if version != "Fedora Linux 38" {
			t.Errorf("Expected 'Fedora Linux 38', got '%s'", version)
		}
	})

	t.Run("NameOnly", func(t *testing.T) {
		content := "NAME=\"Arch Linux\"\n"
		f, err := os.CreateTemp("", "os-release")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(f.Name())
		f.WriteString(content)
		f.Close()

		osReleasePath = f.Name()

		version := getLinuxVersion()
		if version != "Arch Linux" {
			t.Errorf("Expected 'Arch Linux', got '%s'", version)
		}
	})

	t.Run("FileNotFound", func(t *testing.T) {
		osReleasePath = "/non/existent/path"
		version := getLinuxVersion()
		if version != "Linux" {
			t.Errorf("Expected 'Linux', got '%s'", version)
		}
	})
}

func TestGetWindowsVersion(t *testing.T) {
	origExec := execCommand
	defer func() { execCommand = origExec }()

	execCommand = func(name string, args ...string) *exec.Cmd {
		// Verify expected command
		if name != "cmd" {
			return exec.Command("echo", "Unexpected command")
		}
		return exec.Command("echo", "Microsoft Windows [Version 10.0.19045.2965]")
	}

	version := getWindowsVersion()
	if version != "Microsoft Windows [Version 10.0.19045.2965]" {
		t.Errorf("Expected Windows version, got '%s'", version)
	}
}

func TestGetWindowsVersion_Error(t *testing.T) {
	origExec := execCommand
	defer func() { execCommand = origExec }()

	execCommand = func(name string, args ...string) *exec.Cmd {
		// Return a command that fails
		return exec.Command("false")
	}

	version := getWindowsVersion()
	if version != "Windows" {
		t.Errorf("Expected 'Windows', got '%s'", version)
	}
}

func TestGetMacVersion(t *testing.T) {
	origExec := execCommand
	defer func() { execCommand = origExec }()

	execCommand = func(name string, args ...string) *exec.Cmd {
		if name != "sw_vers" {
			return exec.Command("echo", "Unexpected")
		}
		return exec.Command("echo", "13.4")
	}

	version := getMacVersion()
	if version != "macOS 13.4" {
		t.Errorf("Expected 'macOS 13.4', got '%s'", version)
	}
}

func TestGetMacVersion_Error(t *testing.T) {
	origExec := execCommand
	defer func() { execCommand = origExec }()

	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("false")
	}

	version := getMacVersion()
	if version != "macOS" {
		t.Errorf("Expected 'macOS', got '%s'", version)
	}
}
