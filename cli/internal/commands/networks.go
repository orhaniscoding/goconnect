package commands

import (
	"fmt"
	"strings"

	"github.com/orhaniscoding/goconnect/cli/internal/tui"
)

// RunNetworksCommand lists all networks without TUI
func RunNetworksCommand() {
	fmt.Println()
	fmt.Println("  🔗 GoConnect Networks")
	fmt.Println("  ═══════════════════════")

	grpcClient, err := tui.NewGRPCClient()
	if err != nil {
		fmt.Println()
		fmt.Println("  ❌ Daemon is not running")
		fmt.Println("     Start with: goconnect run")
		fmt.Println()
		return
	}
	defer grpcClient.Close()

	networks, err := grpcClient.GetNetworks()
	if err != nil {
		fmt.Printf("\n  ❌ Failed to get networks: %v\n\n", err)
		return
	}

	if len(networks) == 0 {
		fmt.Println()
		fmt.Println("  No networks found.")
		fmt.Println()
		fmt.Println("  Create a network:  goconnect create")
		fmt.Println("  Join a network:    goconnect join")
		fmt.Println()
		return
	}

	// Get current status to highlight active network
	status, _ := grpcClient.GetStatus()

	fmt.Println()
	fmt.Printf("  %-3s %-30s %-10s %s\n", "", "NAME", "ROLE", "ID")
	fmt.Println("  " + strings.Repeat("─", 60))

	for _, n := range networks {
		icon := "⚪"
		if status != nil && (n.ID == status.NetworkName || n.Name == status.NetworkName) && status.Connected {
			icon = "🟢"
		}
		role := n.Role
		if role == "" {
			role = "member"
		}
		fmt.Printf("  %s  %-30s %-10s %s\n", icon, n.Name, role, n.ID)
	}

	fmt.Println()
	fmt.Printf("  Total: %d network(s)\n", len(networks))
	fmt.Println()
}
