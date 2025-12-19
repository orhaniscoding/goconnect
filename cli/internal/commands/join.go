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
	"github.com/orhaniscoding/goconnect/client-daemon/internal/uierrors"
)

// HandleJoinCommand handles the 'join' command, supporting both TUI and flags
func HandleJoinCommand() {
	joinCmd := flag.NewFlagSet("join", flag.ExitOnError)
	invite := joinCmd.String("invite", "", "Invite code to join a network")

	// Parse flags after "join" subcommand
	if len(os.Args) > 2 {
		joinCmd.Parse(os.Args[2:])
	}

	if *invite != "" {
		// Scripting mode
		// Load config
		cfgPath := config.DefaultConfigPath()
		cfg, err := config.LoadConfig(cfgPath)
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		client := api.NewClient(cfg)
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		fmt.Printf("Joining network with invite code '%s'...\n", *invite)
		net, err := client.JoinNetwork(ctx, *invite)
		if err != nil {
			uErr := uierrors.Map(err)
			fmt.Printf("❌ %s\n", uErr.Error())
			os.Exit(1)
		}

		fmt.Printf("✅ Joined network successfully!\n")
		fmt.Printf("   Name: %s\n", net.Name)
		fmt.Printf("   ID:   %s\n", net.ID)
		return
	}

	// TUI mode (default if no flags)
	RunTUIWithState(tui.StateJoinNetwork)
}
