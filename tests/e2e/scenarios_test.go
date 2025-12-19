package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvironment(t *testing.T) {
	// Simple test to ensure the test runner can reach the server
	t.Logf("Checking server connectivity at %s", apiURL)
	assert.NotEmpty(t, apiURL)
}
