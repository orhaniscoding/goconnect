package main

import (
	"bufio"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kardianos/service"

	"github.com/mattn/go-isatty"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/commands"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/config"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/daemon"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/logger"
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
		fmt.Fprintf(os.Stderr, "  create     Create a new network (args: --name <name> [--cidr <cidr>])\n")
		fmt.Fprintf(os.Stderr, "  join       Join a network (args: --invite <code>)\n")
		fmt.Fprintf(os.Stderr, "  doctor     Diagnose configuration and connectivity issues\n")
		fmt.Fprintf(os.Stderr, "  login      Login via CLI (args: -server <url> -token <jwt>)\n")
		fmt.Fprintf(os.Stderr, "  -version   Print version and exit\n")
		flag.PrintDefaults()
	}

	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("goconnect %s (commit %s, build %s) built by %s\n", version, commit, date, builtBy)
		return
	}

	// Check for protocol handler (Deep Linking)
	if len(os.Args) > 1 && strings.HasPrefix(os.Args[1], "goconnect://") {
		if err := commands.HandleDeepLink(os.Args[1]); err != nil {
			slog.Error("Deep link failed", "error", err)
			os.Exit(1)
		}
		return
	}

	// Load configuration
	cfgPath := config.DefaultConfigPath()
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		// If config fails, we might still want to run TUI to help user setup
		logger.Warn("Failed to load config", "error", err)
	}

	svcOptions := make(service.KeyValue)
	if service.Platform() == "windows" {
		svcOptions["StartType"] = "automatic"
	}

	// Initialize Logger
	logPath := ""
	debug := false
	if cfg != nil {
		if cfg.Settings.LogLevel == "debug" {
			debug = true
		}
		// If running as service/daemon, we might want to log to file?
		// For now simple setup:
	}
	// Configure log path properly based on OS
	if logPath == "" {
		if cacheDir, err := os.UserCacheDir(); err == nil {
			logDir := filepath.Join(cacheDir, "goconnect")
			if err := os.MkdirAll(logDir, 0755); err == nil {
				logPath = filepath.Join(logDir, "daemon.log")
			}
		}
	}
	if err := logger.Setup(logPath, debug); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup logger: %v\n", err)
	}

	// Handle commands
	if len(os.Args) > 1 {
		cmd := os.Args[1]
		protoHandler := system.NewProtocolHandler()

		switch cmd {
		case "setup":
			runSetupWizard(bufio.NewReader(os.Stdin), &http.Client{Timeout: 5 * time.Second}, config.SaveConfig)
			return

		case "install":
			if err := daemon.RunDaemon(cfg, version, svcOptions); err != nil {
				slog.Error("Install failed", "error", err)
				os.Exit(1)
			}
			// After service install, register protocol
			if err := protoHandler.Register("goconnect", ""); err != nil {
				slog.Warn("Failed to register protocol handler", "error", err)
			} else {
				fmt.Println("Protocol handler registered.")
			}
			return

		case "uninstall":
			if err := daemon.RunDaemon(cfg, version, svcOptions); err != nil {
				slog.Error("Uninstall failed", "error", err)
				os.Exit(1)
			}
			if err := protoHandler.Unregister("goconnect"); err != nil {
				slog.Warn("Failed to unregister protocol handler", "error", err)
			} else {
				fmt.Println("Protocol handler unregistered.")
			}
			return

		case "run", "start", "stop":
			// Pass control to daemon service logic
			err = daemon.RunDaemon(cfg, version, svcOptions)
			if err != nil {
				slog.Error("GoConnect Daemon failed", "error", err)
				os.Exit(1)
			}
			return

		case "status":
			commands.RunStatusCommand()
			return

		case "networks":
			commands.RunNetworksCommand()
			return

		case "peers":
			commands.RunPeersCommand()
			return

		case "invite":
			commands.RunInviteCommand()
			return

		case "doctor":
			commands.RunDoctorCommand()
			return

		case "create":
			commands.HandleCreateCommand()
			return

		case "join":
			commands.HandleJoinCommand()
			return

		case "voice":
			runVoiceCommand()
			return
		}
	}

	// No arguments provided - Smart first-run detection
	client := tui.NewClient()

	// Check if this is first run (no config file)
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		// If not interactive, fail gracefully (don't block CI/scripts)
		if !isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd()) {
			fmt.Fprintln(os.Stderr, "Error: Configuration file not found. Run 'goconnect setup' to initialize.")
			os.Exit(1)
		}

		// First time user (Interactive) - show friendly welcome
		runFirstTimeWelcome(bufio.NewReader(os.Stdin), config.SaveConfig, commands.RunTUIWithState, &http.Client{Timeout: 5 * time.Second})
		return
	}

	// Config exists, check if daemon is running
	if client.CheckDaemonStatus() {
		// Daemon is running, launch TUI
		commands.RunTUIWithState(tui.StateDashboard)
	} else {
		// Daemon not running - offer to start it (only if interactive)
		if !isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd()) {
			fmt.Println("Daemon is stopped.")
			os.Exit(1)
		}

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
				slog.Error("GoConnect Daemon failed", "error", err)
				os.Exit(1)
			}
		}
	}
}

// runFirstTimeWelcome shows a friendly welcome for first-time users
func runFirstTimeWelcome(reader *bufio.Reader, configSaver func(*config.Config, string) error, tuiRunner func(tui.SessionState), httpClient *http.Client) {
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
		runQuickSetup("create", configSaver, tuiRunner)
	case "2":
		// Quick setup for joining a network
		runQuickSetup("join", configSaver, tuiRunner)
	case "3":
		// Full setup wizard
		runSetupWizard(reader, httpClient, configSaver)
	default:
		fmt.Println("  Invalid choice. Running quick setup...")
		runQuickSetup("create", configSaver, tuiRunner)
	}
}

// runQuickSetup does minimal configuration and launches the appropriate action
func runQuickSetup(action string, saveConfig func(*config.Config, string) error, tuiRunner func(tui.SessionState)) {
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
	cfg.Daemon.LocalPort = 34100
	cfg.Daemon.HealthCheckInterval = 30 * 1000000000 // 30 seconds

	// Set default identity path
	home, _ := os.UserHomeDir()
	cfg.IdentityPath = home + "/.goconnect/identity.json"

	if err := saveConfig(cfg, cfgPath); err != nil {
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
		tuiRunner(tui.StateCreateNetwork)
	} else {
		tuiRunner(tui.StateJoinNetwork)
	}
}

// runSetupWizard runs an interactive setup wizard for first-time configuration
func runSetupWizard(reader *bufio.Reader, httpClient *http.Client, saveConfig func(*config.Config, string) error) {
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
	if err := testServerConnection(httpClient, serverURL); err != nil {
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
	cfg.Daemon.LocalPort = 34100
	cfg.Daemon.HealthCheckInterval = 30 * 1000000000 // 30 seconds in nanoseconds
	cfg.IdentityPath = "./identity.json"

	fmt.Printf("  📁 Saving configuration to: %s\n", cfgPath)

	if err := saveConfig(cfg, cfgPath); err != nil {
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
		fmt.Println("    goconnect install")
		fmt.Println()
		fmt.Println("  After installing, start it with:")
		fmt.Println("    goconnect start")
		fmt.Println()
		fmt.Println("  Or run it in the foreground for debugging:")
		fmt.Println("    goconnect run")
	} else {
		fmt.Println("  ⏭️  Skipping service installation")
		fmt.Println()
		fmt.Println("  To run the daemon manually:")
		fmt.Println("    goconnect run")
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
	fmt.Println("  1. Install the service:  goconnect install")
	fmt.Println("  2. Start the service:    goconnect start")
	fmt.Println("  3. Check status:         goconnect run (foreground)")
	fmt.Println()
	fmt.Println("  💡 Tip: You can re-run 'goconnect setup' anytime to")
	fmt.Println("         reconfigure the daemon.")
	fmt.Println()
}

// testServerConnection tests if the server is reachable
func testServerConnection(client *http.Client, serverURL string) error {
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

// Build trigger - 20251217074002
