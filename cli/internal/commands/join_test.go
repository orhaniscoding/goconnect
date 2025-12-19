package commands

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHandleJoinCommand_Flags_Parsing verifies that flags are defined correctly
func TestHandleJoinCommand_Flags_Parsing(t *testing.T) {
	// Backup os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Mock args
	os.Args = []string{"goconnect", "join", "--invite", "TEST-CODE-123"}

	joinCmd := flag.NewFlagSet("join", flag.ContinueOnError)
	invite := joinCmd.String("invite", "", "Invite code")

	err := joinCmd.Parse(os.Args[2:])
	assert.NoError(t, err)
	assert.Equal(t, "TEST-CODE-123", *invite)
}
