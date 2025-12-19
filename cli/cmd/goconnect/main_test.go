package main

import (
	"bufio"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/orhaniscoding/goconnect/client-daemon/internal/config"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/tui"
	"github.com/stretchr/testify/assert"
)

func TestServerConnection(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client := server.Client()
		err := testServerConnection(client, server.URL)
		assert.NoError(t, err)
	})

	t.Run("ServerError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := server.Client()
		err := testServerConnection(client, server.URL)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "server returned status 500")
	})

	t.Run("NetworkError", func(t *testing.T) {
		client := &http.Client{}
		// Invalid/unreachable URL
		err := testServerConnection(client, "http://invalid.local")
		assert.Error(t, err)
	})
}

func TestRunQuickSetup(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var savedCfg *config.Config
		mockSaver := func(cfg *config.Config, path string) error {
			savedCfg = cfg
			return nil
		}
		var ranState tui.SessionState
		mockTuiRunner := func(state tui.SessionState) {
			ranState = state
		}

		runQuickSetup("create", mockSaver, mockTuiRunner)

		assert.NotNil(t, savedCfg)
		assert.Equal(t, "http://localhost:8081", savedCfg.Server.URL)
		assert.Equal(t, tui.StateCreateNetwork, ranState)
	})

	t.Run("SaveError", func(t *testing.T) {
		mockSaver := func(cfg *config.Config, path string) error {
			return fmt.Errorf("disk full")
		}
		mockTuiRunner := func(state tui.SessionState) {}

		// Should not panic, just return
		runQuickSetup("join", mockSaver, mockTuiRunner)
	})
}

func TestRunFirstTimeWelcome(t *testing.T) {
	t.Run("CreateNetwork", func(t *testing.T) {
		reader := bufio.NewReader(strings.NewReader("1\n"))
		mockSaver := func(cfg *config.Config, path string) error { return nil }
		var ranState tui.SessionState
		mockTuiRunner := func(state tui.SessionState) {
			ranState = state
		}

		runFirstTimeWelcome(reader, mockSaver, mockTuiRunner, &http.Client{})
		assert.Equal(t, tui.StateCreateNetwork, ranState)
	})

	t.Run("JoinNetwork", func(t *testing.T) {
		reader := bufio.NewReader(strings.NewReader("2\n"))
		mockSaver := func(cfg *config.Config, path string) error { return nil }
		var ranState tui.SessionState
		mockTuiRunner := func(state tui.SessionState) {
			ranState = state
		}

		runFirstTimeWelcome(reader, mockSaver, mockTuiRunner, &http.Client{})
		assert.Equal(t, tui.StateJoinNetwork, ranState)
	})

	t.Run("SetupWizardInput", func(t *testing.T) {
		// mock server for wizard health check
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		// Choice 3 (Setup Wizard) -> then some wizard inputs
		// Wizard inputs: serverURL, authToken, interfaceName, installService
		input := server.URL + "\nmy-token\nmy-interface\ny\n"
		// Choice 3 first
		choiceInput := "3\n" + input
		reader := bufio.NewReader(strings.NewReader(choiceInput))
		mockSaver := func(cfg *config.Config, path string) error { return nil }
		mockTuiRunner := func(state tui.SessionState) {}

		runFirstTimeWelcome(reader, mockSaver, mockTuiRunner, http.DefaultClient)
	})
}

func TestRunSetupWizard(t *testing.T) {
	t.Run("FullFlow", func(t *testing.T) {
		// mock server for health check
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		// Inputs: serverURL, authToken, interfaceName, installService
		input := server.URL + "\nmy-token\nmy-interface\ny\n"
		reader := bufio.NewReader(strings.NewReader(input))
		
		var savedCfg *config.Config
		mockSaver := func(cfg *config.Config, path string) error {
			savedCfg = cfg
			return nil
		}

		runSetupWizard(reader, http.DefaultClient, mockSaver)

		assert.NotNil(t, savedCfg)
		if savedCfg != nil {
			assert.Equal(t, server.URL, savedCfg.Server.URL)
			assert.Equal(t, "my-interface", savedCfg.WireGuard.InterfaceName)
		}
	})
}
