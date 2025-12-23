package commands

import (
	"fmt"

	"github.com/orhaniscoding/goconnect/cli/internal/tui"
)

// RunInviteCommand generates an invite code for the current network
func RunInviteCommand() {
	fmt.Println()
	fmt.Println("  🔗 GoConnect Invite")
	fmt.Println("  ═══════════════════")

	grpcClient, err := tui.NewGRPCClient()
	if err != nil {
		fmt.Println()
		fmt.Println("  ❌ Daemon is not running")
		fmt.Println("     Start with: goconnect run")
		fmt.Println()
		return
	}
	defer grpcClient.Close()

	// Check if connected to a network
	status, err := grpcClient.GetStatus()
	if err != nil {
		fmt.Printf("\n  ❌ Failed to get status: %v\n\n", err)
		return
	}

	if !status.Connected || status.NetworkName == "" {
		fmt.Println()
		fmt.Println("  ⚠️  Not connected to any network")
		fmt.Println("     Connect to a network first using the TUI")
		fmt.Println()
		return
	}

	// Find the network ID for the current network
	var networkID string
	for _, n := range status.Networks {
		if n.Name == status.NetworkName || n.ID == status.NetworkName {
			networkID = n.ID
			break
		}
	}

	if networkID == "" {
		fmt.Println()
		fmt.Println("  ⚠️  Could not find current network ID")
		fmt.Println()
		return
	}

	// Generate invite using the status invite code if available
	if status.InviteCode != "" {
		fmt.Println()
		fmt.Printf("  Network: %s\n", status.NetworkName)
		fmt.Println()
		fmt.Println("  📋 Invite Code:")
		fmt.Println()
		fmt.Printf("     %s\n", status.InviteCode)
		fmt.Println()
		fmt.Println("  📎 Invite Link:")
		fmt.Println()
		fmt.Printf("     goconnect://join/%s\n", status.InviteCode)
		fmt.Println()
		fmt.Println("  Share this code or link with others to let them join!")
		fmt.Println()
	} else {
		fmt.Println()
		fmt.Printf("  Network: %s\n", status.NetworkName)
		fmt.Println()
		fmt.Println("  ⚠️  No invite code available")
		fmt.Println("     You may need admin permissions to generate invites")
		fmt.Println()
	}
}
