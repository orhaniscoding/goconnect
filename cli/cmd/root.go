package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "goconnect",
	Short: "GoConnect CLI and Daemon",
	Long:  `GoConnect is a secure, headless network daemon for creating virtual LANs.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Default behavior: Show help
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
