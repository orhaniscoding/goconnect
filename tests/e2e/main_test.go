package e2e

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

var (
	apiURL string
)

func TestMain(m *testing.M) {
	// 1. Setup
	apiURL = os.Getenv("API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}

	fmt.Printf("Endpoint: %s\n", apiURL)
	if err := waitForServer(apiURL); err != nil {
		fmt.Fprintf(os.Stderr, "Server failed to become ready: %v\n", err)
		os.Exit(1)
	}

	// 2. Run Tests
	code := m.Run()

	// 3. Teardown (optional, containers handled by compose)
	os.Exit(code)
}

func waitForServer(url string) error {
	deadline := time.Now().Add(30 * time.Second)
	client := http.Client{Timeout: 1 * time.Second}

	for time.Now().Before(deadline) {
		resp, err := client.Get(url + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("timeout waiting for health check")
}
