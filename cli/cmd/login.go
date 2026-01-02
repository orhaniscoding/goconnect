package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/atotto/clipboard"
	"github.com/orhaniscoding/goconnect/cli/internal/config"
	"github.com/orhaniscoding/goconnect/cli/internal/proto"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to GoConnect using headless device flow",
	Long:  `Log in to GoConnect by authorizing this device via a web browser on another device (e.g., your phone or laptop).`,
	RunE:  runLogin,
}

func init() {
	rootCmd.AddCommand(loginCmd)
}

func runLogin(cmd *cobra.Command, args []string) error {
	// Create context with signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nLogin cancelled.")
		cancel()
	}()

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
			// Check if context was cancelled (user interrupt)
			if ctx.Err() != nil {
				return ctx.Err()
			}
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
	// Load config to get socket path
	cfg, err := config.LoadConfig(config.DefaultConfigPath())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	socketPath := cfg.Daemon.SocketPath

	// Build dial target based on platform
	var dialTarget string
	if runtime.GOOS == "windows" {
		// Windows uses named pipes
		dialTarget = socketPath
	} else {
		// Unix-like systems use Unix domain sockets
		dialTarget = "unix://" + socketPath
	}

	// Use a reasonable timeout for daemon connection
	const dialTimeout = 5 * time.Second
	dialCtx, dialCancel := context.WithTimeout(ctx, dialTimeout)
	defer dialCancel()

	conn, err := grpc.DialContext(dialCtx, dialTarget,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("daemon not reachable at %s: %w (is the daemon running?)", socketPath, err)
	}

	return proto.NewDaemonServiceClient(conn), conn, nil
}
