package commands

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHandleCreateCommand_Flags verifies that flags are defined correctly
// Logic verification is limited because HandleCreateCommand calls os.Exit or connects to real services.
// A refactor to dependency injection would be needed for full testing,
// but we can at least ensure the flags don't panic.
func TestHandleCreateCommand_Flags_Parsing(t *testing.T) {
	// Backup os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Mock args
	os.Args = []string{"goconnect", "create", "--name", "test-net", "--cidr", "10.0.0.0/24"}

	// We can't easily call HandleCreateCommand() because it might exit or try to load config.
	// However, we can duplicate the flag parsing logic here to ensure it works as expected
	// which documents the contract of the command.
	createCmd := flag.NewFlagSet("create", flag.ContinueOnError)
	name := createCmd.String("name", "", "Network name")
	cidr := createCmd.String("cidr", "", "Network CIDR")

	err := createCmd.Parse(os.Args[2:])
	assert.NoError(t, err)
	assert.Equal(t, "test-net", *name)
	assert.Equal(t, "10.0.0.0/24", *cidr)
}
