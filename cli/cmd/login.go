package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/atotto/clipboard"
	"github.com/orhaniscoding/goconnect/cli/internal/proto"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to GoConnect using headless device flow",
	Long:  `Log in to GoConnect by authorizing this device via a web browser on another device (e.g., your phone or laptop).`,
	RunE: runLogin,
}

func init() {
	rootCmd.AddCommand(loginCmd)
}

func runLogin(cmd *cobra.Command, args []string) error {
	ctx := context.Background() // TODO: signal handling
	
	// Connect to Daemon
	client, conn, err := connectDaemon(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer conn.Close()
	
	fmt.Println("Initiating login request...")
	
	// Start Login stream
	stream, err := client.Login(ctx, &proto.LoginRequest{ClientName: "cli"})
	if err != nil {
		return fmt.Errorf("login rpc failed: %w", err)
	}
	
	for {
		update, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("stream error: %w", err)
		}
		
		switch u := update.Update.(type) {
		case *proto.LoginUpdate_Instructions:
			fmt.Printf("\nPlease visit: %s\n", u.Instructions.Url)
			fmt.Printf("And enter code: %s\n", u.Instructions.Code)
			fmt.Println("\nWaiting for approval...")
			
			// Attempt to copy code to clipboard (best effort)
			if err := clipboard.WriteAll(u.Instructions.Code); err == nil {
				fmt.Println("(Code copied to clipboard)")
			}
			
		case *proto.LoginUpdate_Error:
			return fmt.Errorf("login failed: %s", u.Error.Message)
			
		case *proto.LoginUpdate_Success:
			fmt.Println("\nLogin successful!")
			fmt.Printf("Logged in as user ID: %s (%s)\n", u.Success.UserId, u.Success.Email)
			return nil
		}
	}
}

func connectDaemon(ctx context.Context) (proto.DaemonServiceClient, *grpc.ClientConn, error) {
	// TODO: Get socket path from config/flag
	home, _ := os.UserHomeDir()
	configDir := filepath.Join(home, ".goconnect")
	socketPath := filepath.Join(configDir, "daemon.sock")

	conn, err := grpc.DialContext(ctx, "unix://"+socketPath,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(2*time.Second),
	)
	if err != nil {
		return nil, nil, err
	}

	return proto.NewDaemonServiceClient(conn), conn, nil
}
