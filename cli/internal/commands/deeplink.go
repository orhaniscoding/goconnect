package commands

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/orhaniscoding/goconnect/client-daemon/internal/config"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/deeplink"
)

// HandleDeepLink processes a deep link URI
func HandleDeepLink(uri string) error {
	slog.Info("Processing deep link", "uri", uri)

	// Parse using the deeplink package
	dl, err := deeplink.Parse(uri)
	if err != nil {
		return fmt.Errorf("invalid deep link: %w", err)
	}

	slog.Info("Deep link action", "action", dl.Action, "target", dl.Target)

	switch dl.Action {
	case deeplink.ActionLogin:
		return handleLoginDeepLink(dl)
	case deeplink.ActionJoin:
		return handleJoinDeepLink(dl)
	case deeplink.ActionNetwork:
		return handleNetworkDeepLink(dl)
	case deeplink.ActionConnect:
		return handleConnectDeepLink(dl)
	default:
		return fmt.Errorf("unknown action: %s", dl.Action)
	}
}

func handleJoinDeepLink(dl *deeplink.DeepLink) error {
	fmt.Printf("🔗 Joining network with invite code: %s\n", dl.Target)

	handler := deeplink.NewHandler()
	result, err := handler.Handle(dl)
	if err != nil {
		return fmt.Errorf("failed to process join link: %w", err)
	}

	if result.Success {
		fmt.Printf("✅ %s\n", result.Message)
		if networkName, ok := result.Data["network_name"].(string); ok {
			fmt.Printf("   Network: %s\n", networkName)
		}
		if role, ok := result.Data["role"].(string); ok {
			fmt.Printf("   Your role: %s\n", role)
		}
		return nil
	}
	return fmt.Errorf("%s", result.Message)
}

func handleNetworkDeepLink(dl *deeplink.DeepLink) error {
	fmt.Printf("🔗 Opening network: %s\n", dl.Target)

	handler := deeplink.NewHandler()
	result, err := handler.Handle(dl)
	if err != nil {
		return fmt.Errorf("failed to process network link: %w", err)
	}

	if result.Success {
		fmt.Printf("✅ %s\n", result.Message)
		if networkName, ok := result.Data["network_name"].(string); ok {
			fmt.Printf("   Network: %s\n", networkName)
		}
		if connected, ok := result.Data["connected"].(bool); ok && connected {
			fmt.Printf("   Status: Connected\n")
		}
		return nil
	}
	return fmt.Errorf("%s", result.Message)
}

func handleConnectDeepLink(dl *deeplink.DeepLink) error {
	fmt.Printf("🔗 Connecting to peer: %s\n", dl.Target)

	handler := deeplink.NewHandler()
	result, err := handler.Handle(dl)
	if err != nil {
		return fmt.Errorf("failed to process connect link: %w", err)
	}

	if result.Success {
		fmt.Printf("✅ %s\n", result.Message)
		return nil
	}
	return fmt.Errorf("%s", result.Message)
}

func handleLoginDeepLink(dl *deeplink.DeepLink) error {
	token := dl.Params["token"]
	server := dl.Params["server"]

	if token == "" || server == "" {
		return fmt.Errorf("login link missing token or server params")
	}

	// Load config to get Keyring
	cfgPath := config.DefaultConfigPath()
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	slog.Info("Deep Link Login", "server", server, "token", "[REDACTED]")

	// Save Token to Keyring
	if cfg.Keyring != nil {
		if err := cfg.Keyring.StoreAuthToken(token); err != nil {
			return fmt.Errorf("failed to store token: %w", err)
		}
		fmt.Println("Token stored successfully.")
	} else {
		return fmt.Errorf("keyring not available")
	}

	// Restart Service (if needed) or Notify Daemon
	if err := notifyDaemonConnect(); err != nil {
		slog.Warn("Could not notify daemon to connect", "error", err)
		slog.Info("You may need to restart the goconnect-daemon service manually")
	} else {
		fmt.Println("Daemon notified to connect.")
	}
	return nil
}

func notifyDaemonConnect() error {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Post("http://127.0.0.1:34100/connect", "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("daemon returned status %d", resp.StatusCode)
	}
	return nil
}
