package commands

import (
	"fmt"
	"strings"

	"github.com/orhaniscoding/goconnect/client-daemon/internal/tui"
)

// RunPeersCommand lists peers in the current network
func RunPeersCommand() {
	fmt.Println()
	fmt.Println("  👥 GoConnect Peers")
	fmt.Println("  ══════════════════")

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

	if !status.Connected {
		fmt.Println()
		fmt.Println("  ⚠️  Not connected to any network")
		fmt.Println("     Use 'goconnect status' to see available networks")
		fmt.Println()
		return
	}

	peers, err := grpcClient.GetPeers()
	if err != nil {
		fmt.Printf("\n  ❌ Failed to get peers: %v\n\n", err)
		return
	}

	if len(peers) == 0 {
		fmt.Println()
		fmt.Printf("  Network: %s\n", status.NetworkName)
		fmt.Println()
		fmt.Println("  No other peers in this network yet.")
		fmt.Println("  Share your invite code to add peers!")
		fmt.Println()
		return
	}

	fmt.Println()
	fmt.Printf("  Network: %s\n", status.NetworkName)
	fmt.Println()
	fmt.Printf("  %-3s %-25s %-15s %s\n", "", "NAME", "IP", "STATUS")
	fmt.Println("  " + strings.Repeat("─", 55))

	for _, p := range peers {
		icon := "⚪"
		statusText := "offline"
		if p.Status == "online" {
			icon = "🟢"
			statusText = "online"
		} else if p.Status == "idle" {
			icon = "🟡"
			statusText = "idle"
		}
		fmt.Printf("  %s  %-25s %-15s %s\n", icon, p.Name, p.VirtualIP, statusText)
	}

	fmt.Println()
	fmt.Printf("  Total: %d peer(s)\n", len(peers))
	fmt.Println()
}
