//go:build linux

package system

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProtocolHandler_ReturnsHandler(t *testing.T) {
	h := newProtocolHandler()
	assert.NotNil(t, h, "newProtocolHandler should return a non-nil handler")
	
	_, ok := h.(*linuxProtocolHandler)
	assert.True(t, ok, "Should return a linuxProtocolHandler")
}

func TestLinuxProtocolHandler_Register(t *testing.T) {
	// Save original runCommand and restore after test
	origRunCommand := runCommand
	defer func() { runCommand = origRunCommand }()

	// Create a temp directory to act as home
	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	// Mock runCommand to succeed
	runCommand = func(_ string, _ ...string) error {
		return nil
	}

	h := newProtocolHandler()
	err := h.Register("goconnect", "")
	require.NoError(t, err, "Register should succeed")

	// Verify desktop file was created
	desktopFile := filepath.Join(tmpHome, ".local", "share", "applications", "goconnect.desktop")
	content, err := os.ReadFile(desktopFile)
	require.NoError(t, err, "Desktop file should exist")

	contentStr := string(content)
	assert.Contains(t, contentStr, "[Desktop Entry]", "Should have desktop entry header")
	assert.Contains(t, contentStr, "Type=Application", "Should have application type")
	assert.Contains(t, contentStr, "Name=GoConnect", "Should have name")
	assert.Contains(t, contentStr, "MimeType=x-scheme-handler/goconnect;", "Should have mime type")
}

func TestLinuxProtocolHandler_Register_CustomCommand(t *testing.T) {
	origRunCommand := runCommand
	defer func() { runCommand = origRunCommand }()

	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	runCommand = func(name string, args ...string) error {
		return nil
	}

	h := newProtocolHandler()
	customCmd := "/usr/bin/myapp --handle %u"
	err := h.Register("myscheme", customCmd)
	require.NoError(t, err)

	desktopFile := filepath.Join(tmpHome, ".local", "share", "applications", "myscheme.desktop")
	content, err := os.ReadFile(desktopFile)
	require.NoError(t, err)

	assert.Contains(t, string(content), "Exec="+customCmd, "Should use custom command")
	assert.Contains(t, string(content), "x-scheme-handler/myscheme", "Should have correct scheme")
}

func TestLinuxProtocolHandler_Register_XdgMimeFailure(t *testing.T) {
	origRunCommand := runCommand
	defer func() { runCommand = origRunCommand }()

	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	// Make xdg-mime command fail
	runCommand = func(name string, args ...string) error {
		if name == "xdg-mime" {
			return fmt.Errorf("xdg-mime not found")
		}
		return nil
	}

	h := newProtocolHandler()
	err := h.Register("goconnect", "")

	assert.Error(t, err, "Should return error when xdg-mime fails")
	assert.Contains(t, err.Error(), "failed to update mime database", "Error should mention mime database")
}

func TestLinuxProtocolHandler_Register_CreatesDirStructure(t *testing.T) {
	origRunCommand := runCommand
	defer func() { runCommand = origRunCommand }()

	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	runCommand = func(name string, args ...string) error {
		return nil
	}

	// Ensure the applications directory doesn't exist
	applicationsDir := filepath.Join(tmpHome, ".local", "share", "applications")
	_, err := os.Stat(applicationsDir)
	assert.True(t, os.IsNotExist(err), "Applications dir should not exist before Register")

	h := newProtocolHandler()
	err = h.Register("testscheme", "")
	require.NoError(t, err)

	// Directory should now exist
	info, err := os.Stat(applicationsDir)
	require.NoError(t, err, "Applications dir should exist after Register")
	assert.True(t, info.IsDir(), "Should be a directory")
}

func TestLinuxProtocolHandler_Unregister(t *testing.T) {
	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	// Create the directory structure and desktop file
	applicationsDir := filepath.Join(tmpHome, ".local", "share", "applications")
	err := os.MkdirAll(applicationsDir, 0755)
	require.NoError(t, err)

	desktopFile := filepath.Join(applicationsDir, "goconnect.desktop")
	err = os.WriteFile(desktopFile, []byte("[Desktop Entry]\nType=Application"), 0644)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(desktopFile)
	require.NoError(t, err, "Desktop file should exist before unregister")

	h := newProtocolHandler()
	err = h.Unregister("goconnect")
	require.NoError(t, err, "Unregister should succeed")

	// Verify file was removed
	_, err = os.Stat(desktopFile)
	assert.True(t, os.IsNotExist(err), "Desktop file should be removed after unregister")
}

func TestLinuxProtocolHandler_Unregister_FileNotExist(t *testing.T) {
	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	h := newProtocolHandler()
	err := h.Unregister("nonexistent")

	// Should return an error since file doesn't exist
	assert.Error(t, err, "Unregister should fail when file doesn't exist")
}

func TestLinuxProtocolHandler_Register_VerifiesXdgMimeCommand(t *testing.T) {
	origRunCommand := runCommand
	defer func() { runCommand = origRunCommand }()

	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	var capturedCmd string
	var capturedArgs []string
	runCommand = func(name string, args ...string) error {
		capturedCmd = name
		capturedArgs = args
		return nil
	}

	h := newProtocolHandler()
	err := h.Register("testscheme", "")
	require.NoError(t, err)

	assert.Equal(t, "xdg-mime", capturedCmd, "Should call xdg-mime")
	assert.Contains(t, capturedArgs, "default", "Should use 'default' subcommand")
	assert.Contains(t, capturedArgs, "testscheme.desktop", "Should reference desktop file")
	
	// Find the x-scheme-handler argument
	found := false
	for _, arg := range capturedArgs {
		if strings.Contains(arg, "x-scheme-handler/testscheme") {
			found = true
			break
		}
	}
	assert.True(t, found, "Should include scheme handler mime type")
}

func TestLinuxProtocolHandler_Register_MultipleSchemes(t *testing.T) {
	origRunCommand := runCommand
	defer func() { runCommand = origRunCommand }()

	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	runCommand = func(name string, args ...string) error {
		return nil
	}

	h := newProtocolHandler()
	
	// Register multiple schemes
	schemes := []string{"scheme1", "scheme2", "scheme3"}
	for _, scheme := range schemes {
		err := h.Register(scheme, "")
		require.NoError(t, err, "Register should succeed for "+scheme)
	}

	// Verify all desktop files exist
	applicationsDir := filepath.Join(tmpHome, ".local", "share", "applications")
	for _, scheme := range schemes {
		desktopFile := filepath.Join(applicationsDir, scheme+".desktop")
		_, err := os.Stat(desktopFile)
		assert.NoError(t, err, "Desktop file should exist for "+scheme)
	}
}
