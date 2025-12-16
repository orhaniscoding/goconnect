package commands

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/orhaniscoding/goconnect/client-daemon/internal/api"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/config"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/tui"
)

// HandleCreateCommand handles the 'create' command, supporting both TUI and flags
func HandleCreateCommand() {
	createCmd := flag.NewFlagSet("create", flag.ExitOnError)
	name := createCmd.String("name", "", "Network name")
	cidr := createCmd.String("cidr", "", "Network CIDR (e.g. 10.100.0.0/24)")

	// Parse flags after "create" subcommand
	// Note: In main.go it was os.Args[2:], but here we assume caller might handle args?
	// But main.go calls HandleCreateCommand() without args.
	// So we should expect os.Args to still be available and identical.
	// Typically os.Args[0] is program, os.Args[1] is 'create'.
	// So os.Args[2:] are the flags.
	if len(os.Args) > 2 {
		createCmd.Parse(os.Args[2:])
	}

	if *name != "" || *cidr != "" {
		// Scripting mode
		if *name == "" {
			fmt.Println("Error: --name is required when using flags")
			os.Exit(1)
		}

		// Load config
		cfgPath := config.DefaultConfigPath()
		cfg, err := config.LoadConfig(cfgPath)
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		client := api.NewClient(cfg)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		fmt.Printf("Creating network '%s'...\n", *name)
		net, err := client.CreateNetwork(ctx, *name, *cidr)
		if err != nil {
			fmt.Printf("Error creating network: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Network created successfully!\n")
		fmt.Printf("   Name: %s\n", net.Name)
		fmt.Printf("   ID:   %s\n", net.ID)
		if net.InviteCode != "" {
			fmt.Printf("   Invite: %s\n", net.InviteCode)
		}
		return
	}

	// TUI mode (default if no flags)
	RunTUIWithState(tui.StateCreateNetwork)
}
