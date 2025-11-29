package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/kardianos/service"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/config"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/daemon"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/system"
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
		log.Fatalf("Failed to load config: %v", err)
	}

	svcOptions := make(service.KeyValue)
	if service.Platform() == "windows" {
		svcOptions["StartType"] = "automatic"
	}

	// Handle commands that involve protocol registration
	if len(os.Args) > 1 {
		cmd := os.Args[1]
		protoHandler := system.NewProtocolHandler()

		switch cmd {
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
		}
	}

	// Pass control to daemon service logic for other commands (start, stop, run)
	err = daemon.RunDaemon(cfg, version, svcOptions)
	if err != nil {
		log.Fatalf("GoConnect Daemon failed: %v", err)
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
	// In a real implementation, we might want to signal the running service via localhost bridge.
	// TODO: Call http://127.0.0.1:12345/refresh-config if running
}
