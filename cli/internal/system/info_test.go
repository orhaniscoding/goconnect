package system

import (
	"os"
	"os/exec"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetOSVersion_ReturnsNonEmpty(t *testing.T) {
	version := GetOSVersion()
	assert.NotEmpty(t, version, "GetOSVersion should return a non-empty string")
}

func TestGetOSVersion_ContainsOSIdentifier(t *testing.T) {
	version := GetOSVersion()
	
	switch runtime.GOOS {
	case "windows":
		assert.Contains(t, version, "Windows", "Windows version should contain 'Windows'")
	case "darwin":
		assert.Contains(t, version, "macOS", "macOS version should contain 'macOS'")
	case "linux":
		// Linux returns either "Linux" or a distribution name
		assert.NotEmpty(t, version, "Linux version should not be empty")
	default:
		// For other platforms, should return runtime.GOOS
		assert.Equal(t, runtime.GOOS, version)
	}
}

func TestGetOSVersion_Linux(t *testing.T) {
	// Only valid on Linux runtime for GetOSVersion call
	// But we can test getLinuxVersion directly everywhere since it's in info.go without build constraints
	
	// Save/Restore
	origPath := osReleasePath
	defer func() { osReleasePath = origPath }()

	t.Run("PrettyName", func(t *testing.T) {
		start := "PRETTY_NAME=\"Ubuntu 22.04 LTS\"\nNAME=\"Ubuntu\"\n"
		f, err := os.CreateTemp("", "os-release")
		require.NoError(t, err)
		defer os.Remove(f.Name())
		f.WriteString(start)
		f.Close()

		osReleasePath = f.Name()
		
		version := getLinuxVersion()
		assert.Equal(t, "Ubuntu 22.04 LTS", version)
	})

	t.Run("NameAndVersion", func(t *testing.T) {
		content := "NAME=\"Fedora Linux\"\nVERSION_ID=\"38\"\n"
		f, err := os.CreateTemp("", "os-release")
		require.NoError(t, err)
		defer os.Remove(f.Name())
		f.WriteString(content)
		f.Close()

		osReleasePath = f.Name()

		version := getLinuxVersion()
		assert.Equal(t, "Fedora Linux 38", version)
	})

	t.Run("NameOnly", func(t *testing.T) {
		content := "NAME=\"Arch Linux\"\n"
		f, err := os.CreateTemp("", "os-release")
		require.NoError(t, err)
		defer os.Remove(f.Name())
		f.WriteString(content)
		f.Close()

		osReleasePath = f.Name()

		version := getLinuxVersion()
		assert.Equal(t, "Arch Linux", version)
	})

	t.Run("FileNotFound", func(t *testing.T) {
		osReleasePath = "/non/existent/path"
		version := getLinuxVersion()
		assert.Equal(t, "Linux", version)
	})

	t.Run("EmptyFile", func(t *testing.T) {
		f, err := os.CreateTemp("", "os-release")
		require.NoError(t, err)
		defer os.Remove(f.Name())
		f.Close()

		osReleasePath = f.Name()
		version := getLinuxVersion()
		assert.Equal(t, "Linux", version, "Empty file should return 'Linux'")
	})

	t.Run("NoRelevantFields", func(t *testing.T) {
		content := "ID=ubuntu\nID_LIKE=debian\n"
		f, err := os.CreateTemp("", "os-release")
		require.NoError(t, err)
		defer os.Remove(f.Name())
		f.WriteString(content)
		f.Close()

		osReleasePath = f.Name()
		version := getLinuxVersion()
		assert.Equal(t, "Linux", version, "File without NAME/PRETTY_NAME should return 'Linux'")
	})

	t.Run("VersionIDWithoutName", func(t *testing.T) {
		// This tests the case where VERSION_ID exists but NAME doesn't
		content := "VERSION_ID=\"22.04\"\n"
		f, err := os.CreateTemp("", "os-release")
		require.NoError(t, err)
		defer os.Remove(f.Name())
		f.WriteString(content)
		f.Close()

		osReleasePath = f.Name()
		version := getLinuxVersion()
		assert.Equal(t, "Linux", version, "VERSION_ID without NAME should return 'Linux'")
	})

	t.Run("PrettyNameTakesPrecedence", func(t *testing.T) {
		// PRETTY_NAME should be returned even if NAME and VERSION_ID are present
		content := "PRETTY_NAME=\"Ubuntu 22.04.3 LTS\"\nNAME=\"Ubuntu\"\nVERSION_ID=\"22.04\"\n"
		f, err := os.CreateTemp("", "os-release")
		require.NoError(t, err)
		defer os.Remove(f.Name())
		f.WriteString(content)
		f.Close()

		osReleasePath = f.Name()
		version := getLinuxVersion()
		assert.Equal(t, "Ubuntu 22.04.3 LTS", version, "PRETTY_NAME should take precedence")
	})
}

func TestGetWindowsVersion(t *testing.T) {
	origExec := execCommand
	defer func() { execCommand = origExec }()

	t.Run("Success", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			assert.Equal(t, "cmd", name, "Expected 'cmd' command")
			return exec.Command("echo", "Microsoft Windows [Version 10.0.19045.2965]")
		}

		version := getWindowsVersion()
		assert.Equal(t, "Microsoft Windows [Version 10.0.19045.2965]", version)
	})

	t.Run("CommandArgumentsCorrect", func(t *testing.T) {
		var capturedArgs []string
		execCommand = func(name string, args ...string) *exec.Cmd {
			capturedArgs = args
			return exec.Command("echo", "Windows Test")
		}

		getWindowsVersion()
		assert.Equal(t, []string{"/c", "ver"}, capturedArgs, "Expected correct cmd arguments")
	})
}

func TestGetWindowsVersion_Error(t *testing.T) {
	origExec := execCommand
	defer func() { execCommand = origExec }()

	execCommand = func(name string, args ...string) *exec.Cmd {
		// Return a command that fails
		return exec.Command("false")
	}

	version := getWindowsVersion()
	assert.Equal(t, "Windows", version, "Should return 'Windows' on error")
}

func TestGetMacVersion(t *testing.T) {
	origExec := execCommand
	defer func() { execCommand = origExec }()

	t.Run("Success", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			assert.Equal(t, "sw_vers", name, "Expected 'sw_vers' command")
			return exec.Command("echo", "13.4")
		}

		version := getMacVersion()
		assert.Equal(t, "macOS 13.4", version)
	})

	t.Run("CommandArgumentsCorrect", func(t *testing.T) {
		var capturedArgs []string
		execCommand = func(name string, args ...string) *exec.Cmd {
			capturedArgs = args
			return exec.Command("echo", "14.0")
		}

		getMacVersion()
		assert.Equal(t, []string{"-productVersion"}, capturedArgs, "Expected correct sw_vers arguments")
	})

	t.Run("OlderMacOSVersion", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			return exec.Command("echo", "10.15.7")
		}

		version := getMacVersion()
		assert.Equal(t, "macOS 10.15.7", version)
	})
}

func TestGetMacVersion_Error(t *testing.T) {
	origExec := execCommand
	defer func() { execCommand = origExec }()

	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("false")
	}

	version := getMacVersion()
	assert.Equal(t, "macOS", version, "Should return 'macOS' on error")
}

func TestGetMacVersion_WhitespaceHandling(t *testing.T) {
	origExec := execCommand
	defer func() { execCommand = origExec }()

	execCommand = func(name string, args ...string) *exec.Cmd {
		// Output with trailing newline/whitespace
		return exec.Command("echo", "  14.1.1  ")
	}

	version := getMacVersion()
	// The actual output from echo includes newlines which TrimSpace handles
	assert.Contains(t, version, "macOS", "Should contain macOS prefix")
	assert.Contains(t, version, "14.1.1", "Should contain version number")
}

func TestGetWindowsVersion_WhitespaceHandling(t *testing.T) {
	origExec := execCommand
	defer func() { execCommand = origExec }()

	execCommand = func(name string, args ...string) *exec.Cmd {
		// Output with trailing newline/whitespace
		return exec.Command("echo", "  Microsoft Windows [Version 10.0.22621.2506]  ")
	}

	version := getWindowsVersion()
	// TrimSpace should handle whitespace
	assert.Contains(t, version, "Microsoft Windows", "Should contain Windows info")
}
