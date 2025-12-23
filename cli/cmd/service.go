package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/orhaniscoding/goconnect/cli/internal/svc"
	"github.com/spf13/cobra"
)


var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Manage the GoConnect daemon (system service)",
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the daemon as a system service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServiceAction("install")
	},
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall the daemon service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServiceAction("uninstall")
	},
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the daemon service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServiceAction("start")
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the daemon service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServiceAction("stop")
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check service status",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := svc.NewManager()
		if err != nil {
			return fmt.Errorf("error initializing manager: %w", err)
		}
		status, err := mgr.Status()
		if err != nil {
			return fmt.Errorf("error getting status: %w", err)
		}
		fmt.Printf("Service Status: %s\n", status)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serviceCmd)
	serviceCmd.AddCommand(installCmd, uninstallCmd, startCmd, stopCmd, statusCmd)
}

func runServiceAction(action string) error {
	mgr, err := svc.NewManager()
	if err != nil {
		return fmt.Errorf("error initializing manager: %w", err)
	}

	var opErr error
	switch action {
	case "install":
		opErr = mgr.Install()
	case "uninstall":
		opErr = mgr.Uninstall()
	case "start":
		opErr = mgr.Start()
	case "stop":
		opErr = mgr.Stop()
	}

	if opErr != nil {
		if errors.Is(opErr, svc.ErrNotAdmin) {
			return fmt.Errorf("failed to %s service: administrative privileges required (try using sudo)", action)
		}
		if errors.Is(opErr, svc.ErrAlreadyInstalled) {
			fmt.Println("Service is already installed.")
			return nil
		}
		return fmt.Errorf("failed to %s service: %w", action, opErr)
	}
	
	actionPast := action + "ed"
	if strings.HasSuffix(action, "p") {
		actionPast = action + "ped" // stop -> stopped
	} else if strings.HasSuffix(action, "l") {
		actionPast = action + "led" // install -> installed
	}
	
	fmt.Printf("Service %s successfully.\n", actionPast)
	return nil
}



