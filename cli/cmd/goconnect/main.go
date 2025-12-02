package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kardianos/service"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/config"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/daemon"
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
	u, err := url.Parse(uri)
	if err != nil {
		log.Fatalf("Invalid URL: %v", err)
	}

	switch u.Host {
	case "login":
		handleLoginDeepLink(u)
	default:
		log.Fatalf("Unknown action: %s", u.Host)
	}
}

func handleLoginDeepLink(u *url.URL) {
	q := u.Query()
	token := q.Get("token")
	server := q.Get("server")

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
