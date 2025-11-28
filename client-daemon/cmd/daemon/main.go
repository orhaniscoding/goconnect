package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/kardianos/service"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/config"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/daemon" // Import the new daemon package
)

var (
	version = "dev" // This will be set by build flags
	commit  = "none"
	date    = "2025-09-22"
	builtBy = "orhaniscoding"
)

func main() {
	// Custom usage for service commands
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s <command>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  run        Run the daemon directly (default if no command)\n")
		fmt.Fprintf(os.Stderr, "  install    Install the daemon as a system service\n")
		fmt.Fprintf(os.Stderr, "  uninstall  Uninstall the system service\n")
		fmt.Fprintf(os.Stderr, "  start      Start the system service\n")
		fmt.Fprintf(os.Stderr, "  stop       Stop the system service\n")
		fmt.Fprintf(os.Stderr, "  -version   Print version and exit\n")
		flag.PrintDefaults()
	}

	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("goconnect-daemon %s (commit %s, build %s) built by %s\n", version, commit, date, builtBy)
		return
	}

	// Load configuration
	cfgPath := config.DefaultConfigPath()
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Define platform specific service options
	svcOptions := make(service.KeyValue)
	
	// Example: Set specific options for Windows
	if service.Platform() == "windows" {
		svcOptions["StartType"] = "automatic" // Start service automatically on boot
		// You might add more Windows-specific options here if needed
	}

	err = daemon.RunDaemon(cfg, version, svcOptions) // Pass version to RunDaemon
	if err != nil {
		log.Fatalf("GoConnect Daemon failed: %v", err)
	}
}
