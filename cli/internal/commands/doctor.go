package commands

import (
	"fmt"
	"os"

	"github.com/orhaniscoding/goconnect/client-daemon/internal/config"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/tui"
)

// RunDoctorCommand diagnoses configuration and connectivity issues
func RunDoctorCommand() {
	fmt.Println()
	fmt.Println("  🩺 GoConnect Doctor")
	fmt.Println("  ═══════════════════")
	fmt.Println()

	passed := 0
	failed := 0
	warnings := 0

	// Check 1: Configuration file
	fmt.Print("  Checking configuration file... ")
	cfgPath := config.DefaultConfigPath()
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		fmt.Println("⚠️  Not found")
		fmt.Printf("     Path: %s\n", cfgPath)
		fmt.Println("     Run 'goconnect setup' to create configuration")
		warnings++
	} else {
		cfg, err := config.LoadConfig(cfgPath)
		if err != nil {
			fmt.Println("❌ Invalid")
			fmt.Printf("     Error: %v\n", err)
			failed++
		} else {
			fmt.Println("✅ Valid")
			fmt.Printf("     Path: %s\n", cfgPath)
			if cfg.Server.URL != "" {
				fmt.Printf("     Server: %s\n", cfg.Server.URL)
			}
			passed++
		}
	}

	// Check 2: Daemon status
	fmt.Print("  Checking daemon status... ")
	grpcClient, err := tui.NewGRPCClient()
	if err != nil {
		fmt.Println("❌ Not running")
		fmt.Println("     Start with: goconnect run")
		failed++
	} else {
		defer grpcClient.Close()
		if grpcClient.CheckDaemonStatus() {
			fmt.Println("✅ Running")
			// Get version
			if ver, err := grpcClient.GetVersion(); err == nil {
				fmt.Printf("     Version: %s\n", ver.Version)
			}
			passed++
		} else {
			fmt.Println("❌ Not responding")
			failed++
		}
	}

	// Check 3: Server connectivity (if daemon is running)
	if grpcClient != nil {
		fmt.Print("  Checking server connectivity... ")
		status, err := grpcClient.GetStatus()
		if err != nil {
			fmt.Println("⚠️  Cannot determine")
			fmt.Printf("     Error: %v\n", err)
			warnings++
		} else if status.Connected {
			fmt.Println("✅ Connected")
			if status.NetworkName != "" {
				fmt.Printf("     Network: %s\n", status.NetworkName)
			}
			if status.IP != "" {
				fmt.Printf("     Virtual IP: %s\n", status.IP)
			}
			passed++
		} else {
			fmt.Println("⚪ Disconnected")
			fmt.Println("     Not connected to any network")
			warnings++
		}
	}

	// Check 4: Auth token
	fmt.Print("  Checking authentication... ")
	cfg, _ := config.LoadConfig(cfgPath)
	if cfg != nil && cfg.Keyring != nil {
		token, err := cfg.Keyring.RetrieveAuthToken()
		if err != nil || token == "" {
			fmt.Println("⚠️  No token stored")
			fmt.Println("     Login with: goconnect login -server <url> -token <jwt>")
			warnings++
		} else {
			fmt.Println("✅ Token present")
			// Don't show the actual token for security
			fmt.Printf("     Token length: %d chars\n", len(token))
			passed++
		}
	} else {
		// Keyring might be nil if config.LoadConfig failed or wasn't fully init?
		// But in doctor check above we re-loaded config.
		// If keyring is nil, it usually means platform keyring init failed or config struct issue.
		// We'll treat as warning if not available.
		fmt.Println("⚠️  Keyring not available")
		warnings++
	}

	// Check 5: Protocol handler
	fmt.Print("  Checking protocol handler... ")
	// We can't easily check if it's registered, so just note it
	fmt.Println("ℹ️  Info")
	fmt.Println("     goconnect:// URLs are registered during 'goconnect install'")

	// Summary
	fmt.Println()
	fmt.Println("  ─────────────────────────────")
	fmt.Printf("  Summary: %d passed, %d failed, %d warnings\n", passed, failed, warnings)
	fmt.Println()

	if failed > 0 {
		fmt.Println("  ❌ Some checks failed. Please fix the issues above.")
	} else if warnings > 0 {
		fmt.Println("  ⚠️  Some warnings found. GoConnect may work but with limitations.")
	} else {
		fmt.Println("  ✅ All checks passed! GoConnect is properly configured.")
	}
	fmt.Println()
}
