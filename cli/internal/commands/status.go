package commands

import (
	"fmt"

	"github.com/orhaniscoding/goconnect/client-daemon/internal/tui"
)

// RunStatusCommand shows the daemon and connection status without TUI
func RunStatusCommand() {
	fmt.Println()
	fmt.Println("  🔗 GoConnect Status")
	fmt.Println("  ══════════════════════")

	// Try to connect via gRPC
	grpcClient, err := tui.NewGRPCClient()
	if err != nil {
		fmt.Println()
		fmt.Println("  Daemon Status:  ❌ Not Running")
		fmt.Println()
		fmt.Println("  To start the daemon:")
		fmt.Println("    goconnect run      (foreground)")
		fmt.Println("    goconnect start    (background service)")
		fmt.Println()
		return
	}
	defer grpcClient.Close()

	// Check daemon status
	if !grpcClient.CheckDaemonStatus() {
		fmt.Println()
		fmt.Println("  Daemon Status:  ❌ Not Responding")
		fmt.Println()
		return
	}

	fmt.Println()
	fmt.Println("  Daemon Status:  ✅ Running")

	// Get version info
	versionInfo, err := grpcClient.GetVersion()
	if err == nil && versionInfo != nil {
		fmt.Printf("  Daemon Version: %s\n", versionInfo.Version)
	}

	// Get detailed status
	status, err := grpcClient.GetStatus()
	if err != nil {
		fmt.Printf("\n  ⚠️  Could not get detailed status: %v\n", err)
		return
	}

	// Connection status
	fmt.Println()
	if status.Connected {
		fmt.Println("  Connection:     ✅ Connected")
		if status.NetworkName != "" {
			fmt.Printf("  Network:        %s\n", status.NetworkName)
		}
		if status.IP != "" {
			fmt.Printf("  Virtual IP:     %s\n", status.IP)
		}
		fmt.Printf("  Active Peers:   %d\n", status.OnlineMembers)
	} else {
		fmt.Println("  Connection:     ⚪ Disconnected")
	}

	// List networks
	if len(status.Networks) > 0 {
		fmt.Println()
		fmt.Println("  Networks:")
		for _, n := range status.Networks {
			icon := "⚪"
			if n.ID == status.NetworkName || (status.Connected && n.Name == status.NetworkName) {
				icon = "🟢"
			}
			roleStr := ""
			if n.Role != "" {
				roleStr = fmt.Sprintf(" (%s)", n.Role)
			}
			fmt.Printf("    %s %s%s\n", icon, n.Name, roleStr)
		}
	}

	fmt.Println()
}
