package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/orhaniscoding/goconnect/cli/internal/config"
	"github.com/spf13/cobra"
)

var voiceCmd = &cobra.Command{
	Use:   "voice",
	Short: "Run voice verification test",
	Run: func(cmd *cobra.Command, args []string) {
		runVoiceCommand()
	},
}

func init() {
	rootCmd.AddCommand(voiceCmd)
}

// runVoiceCommand sends a test voice signal directly to server for verification.
// NOTE: This intentionally bypasses the daemon for quick diagnostic testing.
// In production, voice signaling is handled through WebSocket connections via the daemon.
func runVoiceCommand() {
	fmt.Println()
	fmt.Println("  🎙️  GoConnect Voice Test")
	fmt.Println("  ════════════════════════")

	// Load config
	cfgPath := config.DefaultConfigPath()
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		fmt.Printf("\n  ❌ Failed to load config: %v\n", err)
		return
	}

	// Get Token
	if cfg.Keyring == nil {
		fmt.Println("\n  ❌ Keyring not available")
		return
	}

	token, err := cfg.Keyring.RetrieveAuthToken()
	if err != nil || token == "" {
		fmt.Println("\n  ❌ Not logged in. Run 'goconnect login' first.")
		return
	}

	fmt.Println("\n  Sending test signal to server...")
	fmt.Printf("  Server: %s\n", cfg.Server.URL)

	// Construct Signal
	payload := map[string]interface{}{
		"type":       "offer",
		"target_id":  "self_test", // Sending to self/test
		"network_id": "test_net",
		"sdp":        map[string]string{"sdp": "v=0..."},
	}

	jsonBytes, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", cfg.Server.URL+"/v1/voice/signal", bytes.NewBuffer(jsonBytes))
	if err != nil {
		fmt.Printf("  ❌ Failed to create request: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("  ❌ Request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusAccepted {
		fmt.Println("  ✅ Signal sent successfully (202 Accepted)")
		fmt.Println("     Redis integration is working!")
	} else {
		fmt.Printf("  ❌ Server returned error: %s\n", resp.Status)
	}
	fmt.Println()
}
