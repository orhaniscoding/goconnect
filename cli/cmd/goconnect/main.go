package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kardianos/service"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/config"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/daemon"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/deeplink"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/system"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/tui"
)

var (
	version = "dev"
	commit  = "none"
	date    = "2025-09-22"
	builtBy = "orhaniscoding"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s <command> [arguments]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s goconnect://<action>...\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  setup      Interactive setup wizard (recommended for first time)\n")
		fmt.Fprintf(os.Stderr, "  run        Run the daemon directly\n")
		fmt.Fprintf(os.Stderr, "  install    Install the daemon as a system service & register protocol\n")
		fmt.Fprintf(os.Stderr, "  uninstall  Uninstall the system service & unregister protocol\n")
		fmt.Fprintf(os.Stderr, "  start      Start the system service\n")
		fmt.Fprintf(os.Stderr, "  stop       Stop the system service\n")
		fmt.Fprintf(os.Stderr, "  status     Show daemon and connection status\n")
		fmt.Fprintf(os.Stderr, "  networks   List all networks\n")
		fmt.Fprintf(os.Stderr, "  peers      List peers in the current network\n")
		fmt.Fprintf(os.Stderr, "  invite     Generate an invite code for the current network\n")
		fmt.Fprintf(os.Stderr, "  doctor     Diagnose configuration and connectivity issues\n")
		fmt.Fprintf(os.Stderr, "  login      Login via CLI (args: -server <url> -token <jwt>)\n")
		fmt.Fprintf(os.Stderr, "  -version   Print version and exit\n")
		flag.PrintDefaults()
	}

	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("goconnect-daemon %s (commit %s, build %s) built by %s\n", version, commit, date, builtBy)
		return
	}

	// Check for protocol handler (Deep Linking)
	if len(os.Args) > 1 && strings.HasPrefix(os.Args[1], "goconnect://") {
		handleDeepLink(os.Args[1])
		return
	}

	// Load configuration
	cfgPath := config.DefaultConfigPath()
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		// If config fails, we might still want to run TUI to help user setup
		log.Printf("Warning: Failed to load config: %v", err)
	}

	svcOptions := make(service.KeyValue)
	if service.Platform() == "windows" {
		svcOptions["StartType"] = "automatic"
	}

	// Handle commands
	if len(os.Args) > 1 {
		cmd := os.Args[1]
		protoHandler := system.NewProtocolHandler()

		switch cmd {
		case "setup":
			runSetupWizard()
			return

		case "install":
			if err := daemon.RunDaemon(cfg, version, svcOptions); err != nil {
				log.Fatalf("Install failed: %v", err)
			}
			// After service install, register protocol
			if err := protoHandler.Register("goconnect", ""); err != nil {
				log.Printf("Warning: Failed to register protocol handler: %v", err)
			} else {
				fmt.Println("Protocol handler registered.")
			}
			return

		case "uninstall":
			if err := daemon.RunDaemon(cfg, version, svcOptions); err != nil {
				log.Fatalf("Uninstall failed: %v", err)
			}
			if err := protoHandler.Unregister("goconnect"); err != nil {
				log.Printf("Warning: Failed to unregister protocol handler: %v", err)
			} else {
				fmt.Println("Protocol handler unregistered.")
			}
			return

		case "run", "start", "stop":
			// Pass control to daemon service logic
			err = daemon.RunDaemon(cfg, version, svcOptions)
			if err != nil {
				log.Fatalf("GoConnect Daemon failed: %v", err)
			}
			return

		case "status":
			runStatusCommand()
			return

		case "networks":
			runNetworksCommand()
			return

		case "peers":
			runPeersCommand()
			return

		case "invite":
			runInviteCommand()
			return

		case "doctor":
			runDoctorCommand()
			return

		case "create":
			runTUIWithState(tui.StateCreateNetwork)
			return

		case "join":
			runTUIWithState(tui.StateJoinNetwork)
			return
		}
	}

	// No arguments provided - Smart first-run detection
	client := tui.NewClient()
	
	// Check if this is first run (no config file)
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		// First time user - show friendly welcome
		runFirstTimeWelcome()
		return
	}
	
	// Config exists, check if daemon is running
	if client.CheckDaemonStatus() {
		// Daemon is running, launch TUI
		runTUI()
	} else {
		// Daemon not running - offer to start it
		fmt.Println()
		fmt.Println("  🔗 GoConnect")
		fmt.Println()
		fmt.Println("  The daemon is not running.")
		fmt.Println()
		fmt.Println("  Quick actions:")
		fmt.Println("    1. Start daemon:  goconnect run")
		fmt.Println("    2. Reconfigure:   goconnect setup")
		fmt.Println()
		fmt.Print("  Start daemon now? [Y/n]: ")
		
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		
		if answer == "" || answer == "y" || answer == "yes" {
			fmt.Println()
			fmt.Println("  Starting daemon...")
			err = daemon.RunDaemon(cfg, version, svcOptions)
			if err != nil {
				log.Fatalf("GoConnect Daemon failed: %v", err)
			}
		}
	}
}

func runTUI() {
	runTUIWithState(tui.StateDashboard)
}

func runTUIWithState(initialState tui.SessionState) {
	model := tui.NewModelWithState(initialState)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

// runStatusCommand shows the daemon and connection status without TUI
func runStatusCommand() {
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

// runNetworksCommand lists all networks without TUI
func runNetworksCommand() {
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

// runPeersCommand lists peers in the current network
func runPeersCommand() {
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

// runInviteCommand generates an invite code for the current network
func runInviteCommand() {
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

// runDoctorCommand diagnoses configuration and connectivity issues
func runDoctorCommand() {
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

// runFirstTimeWelcome shows a friendly welcome for first-time users
func runFirstTimeWelcome() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println("  ╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("  ║                                                           ║")
	fmt.Println("  ║   🔗 Welcome to GoConnect!                                ║")
	fmt.Println("  ║                                                           ║")
	fmt.Println("  ║   Virtual LAN made simple.                                ║")
	fmt.Println("  ║                                                           ║")
	fmt.Println("  ╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("  What would you like to do?")
	fmt.Println()
	fmt.Println("    1. 🌐 Create a network    - Start your own private LAN")
	fmt.Println("    2. 🔗 Join a network      - Connect with an invite link")
	fmt.Println("    3. ⚙️  Advanced setup      - Configure server, auth, etc.")
	fmt.Println()
	fmt.Print("  Enter choice [1]: ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)
	if choice == "" {
		choice = "1"
	}

	switch choice {
	case "1":
		// Quick setup for creating a network
		runQuickSetup("create")
	case "2":
		// Quick setup for joining a network
		runQuickSetup("join")
	case "3":
		// Full setup wizard
		runSetupWizard()
	default:
		fmt.Println("  Invalid choice. Running quick setup...")
		runQuickSetup("create")
	}
}

// runQuickSetup does minimal configuration and launches the appropriate action
func runQuickSetup(action string) {
	fmt.Println()
	fmt.Println("  ⚡ Quick Setup")
	fmt.Println("  ─────────────────────────────────────────────────────────────")
	fmt.Println()
	fmt.Println("  Setting up with default configuration...")
	fmt.Println()

	// Create default config
	cfgPath := config.DefaultConfigPath()
	cfg := &config.Config{}
	cfg.Server.URL = "http://localhost:8081" // Default server
	cfg.WireGuard.InterfaceName = "goconnect0"
	cfg.Daemon.LocalPort = 12345
	cfg.Daemon.HealthCheckInterval = 30 * 1000000000 // 30 seconds

	// Set default identity path
	home, _ := os.UserHomeDir()
	cfg.IdentityPath = home + "/.goconnect/identity.json"

	if err := config.SaveConfig(cfg, cfgPath); err != nil {
		fmt.Printf("  ❌ Failed to save configuration: %v\n", err)
		fmt.Println()
		fmt.Println("  Please run 'goconnect setup' for manual configuration.")
		return
	}

	fmt.Println("  ✅ Configuration created!")
	fmt.Println()
	fmt.Println("  📝 Note: Using default server (localhost:8081)")
	fmt.Println("     Run 'goconnect setup' to change server settings.")
	fmt.Println()
	fmt.Println("  ⏳ Starting GoConnect...")
	fmt.Println()

	// Launch TUI with the appropriate state
	if action == "create" {
		runTUIWithState(tui.StateCreateNetwork)
	} else {
		runTUIWithState(tui.StateJoinNetwork)
	}
}

func handleDeepLink(uri string) {
	log.Printf("Processing deep link: %s", uri)
	
	// Parse using the deeplink package
	dl, err := deeplink.Parse(uri)
	if err != nil {
		log.Fatalf("Invalid deep link: %v", err)
	}

	log.Printf("Action: %s, Target: %s", dl.Action, dl.Target)

	switch dl.Action {
	case deeplink.ActionLogin:
		handleLoginDeepLink(dl)
	case deeplink.ActionJoin:
		handleJoinDeepLink(dl)
	case deeplink.ActionNetwork:
		handleNetworkDeepLink(dl)
	case deeplink.ActionConnect:
		handleConnectDeepLink(dl)
	default:
		log.Fatalf("Unknown action: %s", dl.Action)
	}
}

func handleJoinDeepLink(dl *deeplink.DeepLink) {
	fmt.Printf("🔗 Joining network with invite code: %s\n", dl.Target)
	
	handler := deeplink.NewHandler()
	result, err := handler.Handle(dl)
	if err != nil {
		log.Fatalf("Failed to process join link: %v", err)
	}

	if result.Success {
		fmt.Printf("✅ %s\n", result.Message)
		if networkName, ok := result.Data["network_name"].(string); ok {
			fmt.Printf("   Network: %s\n", networkName)
		}
		if role, ok := result.Data["role"].(string); ok {
			fmt.Printf("   Your role: %s\n", role)
		}
	} else {
		fmt.Printf("❌ %s\n", result.Message)
		os.Exit(1)
	}
}

func handleNetworkDeepLink(dl *deeplink.DeepLink) {
	fmt.Printf("🔗 Opening network: %s\n", dl.Target)
	
	handler := deeplink.NewHandler()
	result, err := handler.Handle(dl)
	if err != nil {
		log.Fatalf("Failed to process network link: %v", err)
	}

	if result.Success {
		fmt.Printf("✅ %s\n", result.Message)
		if networkName, ok := result.Data["network_name"].(string); ok {
			fmt.Printf("   Network: %s\n", networkName)
		}
		if connected, ok := result.Data["connected"].(bool); ok && connected {
			fmt.Printf("   Status: Connected\n")
		}
	} else {
		fmt.Printf("❌ %s\n", result.Message)
		os.Exit(1)
	}
}

func handleConnectDeepLink(dl *deeplink.DeepLink) {
	fmt.Printf("🔗 Connecting to peer: %s\n", dl.Target)
	
	handler := deeplink.NewHandler()
	result, err := handler.Handle(dl)
	if err != nil {
		log.Fatalf("Failed to process connect link: %v", err)
	}

	if result.Success {
		fmt.Printf("✅ %s\n", result.Message)
	} else {
		fmt.Printf("❌ %s\n", result.Message)
		os.Exit(1)
	}
}

func handleLoginDeepLink(dl *deeplink.DeepLink) {
	token := dl.Params["token"]
	server := dl.Params["server"]

	if token == "" || server == "" {
		log.Fatal("Login link missing token or server params")
	}

	// Load config to get Keyring
	cfgPath := config.DefaultConfigPath()
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 1. Save Server URL
	// Note: This is a bit hacky, ideally we'd update the YAML file.
	// Since config.LoadConfig reads from file but doesn't expose a Save method yet,
	// we might need to implement Save or just log it for now.
	// For Phase 1, let's assume we update the config file.
	// But config package doesn't have SaveConfig yet. Let's add it?
	// Or just print for now as Proof of Concept.

	log.Printf("Deep Link Login:\nServer: %s\nToken: [REDACTED]", server)

	// 2. Save Token to Keyring
	if cfg.Keyring != nil {
		if err := cfg.Keyring.StoreAuthToken(token); err != nil {
			log.Fatalf("Failed to store token: %v", err)
		}
		fmt.Println("Token stored successfully.")
	} else {
		log.Fatal("Keyring not available.")
	}

	// 3. Restart Service (if needed) or Notify Daemon
	if err := notifyDaemonConnect(); err != nil {
		log.Printf("Warning: Could not notify daemon to connect: %v", err)
		log.Println("You may need to restart the goconnect-daemon service manually.")
	} else {
		fmt.Println("Daemon notified to connect.")
	}
}

func notifyDaemonConnect() error {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Post("http://127.0.0.1:12345/connect", "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("daemon returned status %d", resp.StatusCode)
	}
	return nil
}

// runSetupWizard runs an interactive setup wizard for first-time configuration
func runSetupWizard() {
	reader := bufio.NewReader(os.Stdin)

	// Print welcome banner
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║           🚀 GoConnect Daemon Setup Wizard                   ║")
	fmt.Println("║                                                              ║")
	fmt.Println("║  This wizard will help you configure the GoConnect daemon   ║")
	fmt.Println("║  to connect to your GoConnect server.                        ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Step 1: Server URL
	fmt.Println("┌─────────────────────────────────────────────────────────────┐")
	fmt.Println("│ Step 1: Server Connection                                   │")
	fmt.Println("└─────────────────────────────────────────────────────────────┘")
	fmt.Println()
	fmt.Println("  ❓ What is Server URL?")
	fmt.Println("     The address of your GoConnect server.")
	fmt.Println()
	fmt.Println("  📝 Examples:")
	fmt.Println("     • http://localhost:8081 (local development)")
	fmt.Println("     • https://vpn.company.com (production)")
	fmt.Println()
	fmt.Print("  🌐 Server URL [http://localhost:8081]: ")

	serverURL, _ := reader.ReadString('\n')
	serverURL = strings.TrimSpace(serverURL)
	if serverURL == "" {
		serverURL = "http://localhost:8081"
	}

	// Validate URL
	if !strings.HasPrefix(serverURL, "http://") && !strings.HasPrefix(serverURL, "https://") {
		serverURL = "http://" + serverURL
	}

	fmt.Printf("  ✅ Server URL: %s\n", serverURL)
	fmt.Println()

	// Test connection
	fmt.Println("  ⏳ Testing connection to server...")
	if err := testServerConnection(serverURL); err != nil {
		fmt.Printf("  ⚠️  Warning: Could not connect to server: %v\n", err)
		fmt.Print("  Continue anyway? [y/N]: ")
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Println("  ❌ Setup cancelled.")
			return
		}
	} else {
		fmt.Println("  ✅ Server is reachable!")
	}
	fmt.Println()

	// Step 2: Authentication Token
	fmt.Println("┌─────────────────────────────────────────────────────────────┐")
	fmt.Println("│ Step 2: Authentication                                      │")
	fmt.Println("└─────────────────────────────────────────────────────────────┘")
	fmt.Println()
	fmt.Println("  ❓ What is Auth Token?")
	fmt.Println("     A JWT token from the GoConnect server to authenticate")
	fmt.Println("     this device. You can get it from the server web UI.")
	fmt.Println()
	fmt.Println("  📝 How to get a token:")
	fmt.Println("     1. Login to GoConnect web interface")
	fmt.Println("     2. Go to Settings → Devices → Add Device")
	fmt.Println("     3. Copy the token or use the QR code")
	fmt.Println()
	fmt.Print("  🔑 Auth Token (or 'skip' to configure later): ")

	authToken, _ := reader.ReadString('\n')
	authToken = strings.TrimSpace(authToken)

	if authToken != "" && authToken != "skip" {
		fmt.Println("  ✅ Auth token received")
	} else {
		fmt.Println("  ⏭️  Skipping authentication (configure later with 'login' command)")
		authToken = ""
	}
	fmt.Println()

	// Step 3: WireGuard Interface
	fmt.Println("┌─────────────────────────────────────────────────────────────┐")
	fmt.Println("│ Step 3: WireGuard Interface                                 │")
	fmt.Println("└─────────────────────────────────────────────────────────────┘")
	fmt.Println()
	fmt.Println("  ❓ What is Interface Name?")
	fmt.Println("     The name of the WireGuard network interface on your system.")
	fmt.Println()
	fmt.Println("  📝 Examples:")
	fmt.Println("     • goconnect0 (recommended)")
	fmt.Println("     • wg0 (traditional WireGuard naming)")
	fmt.Println()
	fmt.Print("  🔧 Interface Name [goconnect0]: ")

	interfaceName, _ := reader.ReadString('\n')
	interfaceName = strings.TrimSpace(interfaceName)
	if interfaceName == "" {
		interfaceName = "goconnect0"
	}

	fmt.Printf("  ✅ Interface: %s\n", interfaceName)
	fmt.Println()

	// Step 4: Save Configuration
	fmt.Println("┌─────────────────────────────────────────────────────────────┐")
	fmt.Println("│ Step 4: Save Configuration                                  │")
	fmt.Println("└─────────────────────────────────────────────────────────────┘")
	fmt.Println()

	cfgPath := config.DefaultConfigPath()
	cfg := &config.Config{}
	cfg.Server.URL = serverURL
	cfg.WireGuard.InterfaceName = interfaceName
	cfg.Daemon.LocalPort = 12345
	cfg.Daemon.HealthCheckInterval = 30 * 1000000000 // 30 seconds in nanoseconds
	cfg.IdentityPath = "./identity.json"

	fmt.Printf("  📁 Saving configuration to: %s\n", cfgPath)

	if err := config.SaveConfig(cfg, cfgPath); err != nil {
		fmt.Printf("  ❌ Failed to save configuration: %v\n", err)
		return
	}
	fmt.Println("  ✅ Configuration saved!")

	// Save auth token if provided
	if authToken != "" {
		cfg, err := config.LoadConfig(cfgPath)
		if err == nil && cfg.Keyring != nil {
			if err := cfg.Keyring.StoreAuthToken(authToken); err != nil {
				fmt.Printf("  ⚠️  Warning: Could not store auth token: %v\n", err)
			} else {
				fmt.Println("  ✅ Auth token stored securely!")
			}
		}
	}
	fmt.Println()

	// Step 5: Install as Service
	fmt.Println("┌─────────────────────────────────────────────────────────────┐")
	fmt.Println("│ Step 5: Install as System Service (Optional)                │")
	fmt.Println("└─────────────────────────────────────────────────────────────┘")
	fmt.Println()
	fmt.Println("  ❓ What does this do?")
	fmt.Println("     Installs GoConnect as a background service that starts")
	fmt.Println("     automatically when your computer boots.")
	fmt.Println()
	fmt.Print("  🔧 Install as system service? [Y/n]: ")

	installService, _ := reader.ReadString('\n')
	installService = strings.TrimSpace(strings.ToLower(installService))

	if installService == "" || installService == "y" || installService == "yes" {
		fmt.Println("  ⏳ Installing system service...")
		fmt.Println("  ⚠️  Note: This may require administrator/root privileges.")
		fmt.Println()
		fmt.Println("  To install the service manually, run:")
		fmt.Println("    goconnect-daemon install")
		fmt.Println()
		fmt.Println("  To start the service after installation:")
		fmt.Println("    goconnect-daemon start")
	} else {
		fmt.Println("  ⏭️  Skipping service installation")
		fmt.Println()
		fmt.Println("  To run the daemon manually:")
		fmt.Println("    goconnect-daemon run")
	}
	fmt.Println()

	// Summary
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    🎉 Setup Complete!                        ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("  📋 Configuration Summary:")
	fmt.Println("  ─────────────────────────────────────────────────────────────")
	fmt.Printf("  • Server URL:     %s\n", serverURL)
	fmt.Printf("  • Interface:      %s\n", interfaceName)
	fmt.Printf("  • Config Path:    %s\n", cfgPath)
	if authToken != "" {
		fmt.Println("  • Auth Token:     ✅ Configured")
	} else {
		fmt.Println("  • Auth Token:     ⚠️  Not configured (use 'login' command)")
	}
	fmt.Println()
	fmt.Println("  🚀 Next Steps:")
	fmt.Println("  ─────────────────────────────────────────────────────────────")
	fmt.Println("  1. Install the service:  goconnect-daemon install")
	fmt.Println("  2. Start the service:    goconnect-daemon start")
	fmt.Println("  3. Check status:         goconnect-daemon run (foreground)")
	fmt.Println()
	fmt.Println("  💡 Tip: You can re-run 'goconnect-daemon setup' anytime to")
	fmt.Println("         reconfigure the daemon.")
	fmt.Println()
}

// testServerConnection tests if the server is reachable
func testServerConnection(serverURL string) error {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(serverURL + "/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}
	return nil
}
