package cmd

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/orhaniscoding/goconnect/cli/internal/commands"
	"github.com/orhaniscoding/goconnect/cli/internal/config"
	"github.com/orhaniscoding/goconnect/cli/internal/daemon"
	"github.com/kardianos/service"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the daemon (foreground)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath := config.DefaultConfigPath()
		cfg, err := config.LoadConfig(cfgPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		svcOptions := make(service.KeyValue)
		if service.Platform() == "windows" {
			svcOptions["StartType"] = "automatic"
		}
		
		// Assuming version is injected or we fetch it
		return daemon.RunDaemon(cfg, "dev", svcOptions)
	},
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive setup wizard",
	Run: func(cmd *cobra.Command, args []string) {
		runSetupWizard(bufio.NewReader(os.Stdin), &http.Client{Timeout: 5 * time.Second}, config.SaveConfig)
	},
}

var networksCmd = &cobra.Command{
	Use:   "networks",
	Short: "List all networks",
	Run: func(cmd *cobra.Command, args []string) {
		commands.RunNetworksCommand()
	},
}

var peersCmd = &cobra.Command{
	Use:   "peers",
	Short: "List peers in current network",
	Run: func(cmd *cobra.Command, args []string) {
		commands.RunPeersCommand()
	},
}

var statusCmdLegacy = &cobra.Command{
	Use:   "status",
	Short: "Show daemon status",
	Run: func(cmd *cobra.Command, args []string) {
		commands.RunStatusCommand()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(networksCmd)
	rootCmd.AddCommand(peersCmd)
	rootCmd.AddCommand(statusCmdLegacy)
}
