package system

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Variables for mocking in tests
var (
	osReleasePath = "/etc/os-release"
	execCommand   = exec.Command
)

// GetOSVersion returns a human-readable OS version string
func GetOSVersion() string {
	switch runtime.GOOS {
	case "windows":
		return getWindowsVersion()
	case "linux":
		return getLinuxVersion()
	case "darwin":
		return getMacVersion()
	default:
		return runtime.GOOS
	}
}

func getWindowsVersion() string {
	cmd := execCommand("cmd", "/c", "ver")
	out, err := cmd.Output()
	if err != nil {
		return "Windows"
	}
	return strings.TrimSpace(string(out))
}

func getLinuxVersion() string {
	// Try /etc/os-release
	f, err := os.Open(osReleasePath)
	if err != nil {
		return "Linux"
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var name, version string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			return strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
		}
		if strings.HasPrefix(line, "NAME=") {
			name = strings.Trim(strings.TrimPrefix(line, "NAME="), "\"")
		}
		if strings.HasPrefix(line, "VERSION_ID=") {
			version = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), "\"")
		}
	}
	if name != "" {
		if version != "" {
			return fmt.Sprintf("%s %s", name, version)
		}
		return name
	}
	return "Linux"
}

func getMacVersion() string {
	cmd := execCommand("sw_vers", "-productVersion")
	out, err := cmd.Output()
	if err != nil {
		return "macOS"
	}
	return "macOS " + strings.TrimSpace(string(out))
}
