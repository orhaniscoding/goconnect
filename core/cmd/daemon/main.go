package main

import (
	"flag"
	"log"
	"os"

	"path/filepath"

	"github.com/kardianos/service"
	"github.com/orhaniscoding/goconnect/server/internal/auth"
	"github.com/orhaniscoding/goconnect/server/internal/daemon"
	"github.com/orhaniscoding/goconnect/server/internal/logger"
	"github.com/orhaniscoding/goconnect/server/internal/rpc"
	appService "github.com/orhaniscoding/goconnect/server/internal/service"
)
var (
	env       = flag.String("env", "development", "environment (development|production)")
	logPath    = flag.String("log-path", "daemon.log", "path to log file")
	configDir  = flag.String("config-dir", "", "path to config directory")
	socketPath = flag.String("socket-path", "", "path to unix domain socket")
	backendURL = flag.String("backend-url", "https://api.goconnect.dev", "URL of the GoConnect backend")
)

func main() {
	flag.Parse()

	if *configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("failed to get user home directory: %v", err)
		}
		*configDir = filepath.Join(home, ".config", "goconnect")
	}

	// Initialize Logger
	logger.Setup(logger.Config{
		Environment: *env,
		LogPath:     *logPath,
		MaxSize:     100, // 100MB
		MaxBackups:  5,
		MaxAge:      30, // 30 days
		Compress:    true,
	})

	if *socketPath == "" {
		*socketPath = filepath.Join(*configDir, "daemon.sock")
	}

	// Initialize TokenManager
	tokenMgr, err := auth.NewTokenManager(*configDir)
	if err != nil {
		logger.Error("failed to initialize token manager", "error", err)
		log.Fatalf("failed to initialize token manager: %v", err)
	}

	keyMgr := auth.NewKeyManager(*configDir)
	d := daemon.New(keyMgr, tokenMgr, "0.1.0-mvp", *backendURL)

	// Initialize RPC Listener (Unix Socket)
	lis, err := rpc.ListenUnix(*socketPath)
	if err != nil {
		logger.Error("failed to create rpc listener", "error", err, "path", *socketPath)
		// We don't fatal here yet, but maybe we should if IPC is critical
	} else {
		d.Listener = lis
	}

	prg := &appService.Program{
		Daemon: d,
	}

	svcConfig := &service.Config{
		Name:        "goconnect",
		DisplayName: "GoConnect Daemon",
		Description: "Secure virtual LAN daemon for GoConnect.",
	}

	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatalf("failed to create service: %v", err)
	}

	// service.Run is blocking
	if err := s.Run(); err != nil {
		logger.Error("service failed", "error", err)
		os.Exit(1)
	}
}

