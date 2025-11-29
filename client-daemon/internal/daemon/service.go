package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/kardianos/service"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/api"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/config"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/engine"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/identity"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/system"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/wireguard"
)

// DaemonService implements the kardianos/service.Service interface.
type DaemonService struct {
	config *config.Config
	engine *engine.Engine
	logf   service.Logger
	cancel context.CancelFunc
	localHTTPServer *http.Server
	idManager *identity.Manager
	apiClient *api.Client
	daemonVersion string // Added daemon version
}

// NewDaemonService creates a new DaemonService.
func NewDaemonService(cfg *config.Config, daemonVersion string) *DaemonService {
	return &DaemonService{
		config: cfg,
		daemonVersion: daemonVersion,
	}
}

// Start is called when the service is starting.
func (s *DaemonService) Start(srv service.Service) error {
	var ctx context.Context
	ctx, s.cancel = context.WithCancel(context.Background())

	// Initialize logger (already set in RunDaemon for interactive/service mode)
	if s.logf == nil { // Fallback, should ideally not happen if RunDaemon sets it
		stdLogger := log.New(os.Stderr, "[goconnect-daemon] ", log.LstdFlags)
		s.logf = &fallbackServiceLogger{stdLogger}
	}

	s.logf.Info("GoConnect Daemon starting...")

	// Initialize Identity Manager
	s.idManager = identity.NewManager(s.config.IdentityPath)
	deviceIdentity, err := s.idManager.LoadOrCreateIdentity()
	if err != nil {
		s.logf.Error("Failed to load or create device identity: ", err.Error())
		return err
	}
	s.config.DevicePrivateKey = deviceIdentity.PrivateKey // Ensure private key is set in config
	s.config.DevicePublicKey = deviceIdentity.PublicKey   // Ensure public key is set in config

	// Initialize WireGuard client
	wgClient, err := wireguard.NewClient(s.config.WireGuard.InterfaceName)
	if err != nil {
		s.logf.Error("Failed to initialize WireGuard client: ", err.Error())
		return err
	}

	// Initialize API Client
	s.apiClient = api.NewClient(s.config) // Pass full config

	// Initialize Engine
	s.engine = engine.NewEngine(s.config, s.idManager, wgClient, s.apiClient, s.logf)
	s.engine.Start()


	// Setup Localhost Bridge Handlers
	mux := http.NewServeMux() // Use a new ServeMux for the local HTTP server
	s.setupLocalhostBridgeHandlers(mux)
	s.localHTTPServer = &http.Server{
		Addr:              fmt.Sprintf("127.0.0.1:%d", s.config.Daemon.LocalPort),
		Handler:           mux, // Assign the mux to the server
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	go func() {
		s.logf.Info("Daemon bridge at http://", s.localHTTPServer.Addr)
		if err := s.localHTTPServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logf.Error("Localhost bridge failed: ", err.Error())
		}
	}()

	// Auto-connect logic
	go s.autoConnectLoop(ctx)


	go s.run(ctx)
	return nil
}

// Stop is called when the service is stopping.
func (s *DaemonService) Stop(srk service.Service) error {
	s.logf.Info("GoConnect Daemon stopping...")
	if s.cancel != nil {
		s.cancel()
	}
	if s.engine != nil {
		s.engine.Stop()
	}
	if s.localHTTPServer != nil {
		s.localHTTPServer.Shutdown(context.Background())
	}
	s.logf.Info("GoConnect Daemon stopped.")
	return nil
}

// run contains the main logic of the daemon.
func (s *DaemonService) run(ctx context.Context) {
	s.logf.Info("Daemon main loop started.")
	
	// Example: Periodically check server connection or trigger updates
	ticker := time.NewTicker(s.config.Daemon.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logf.Info("Daemon run loop context cancelled.")
			return
		case <-ticker.C:
			// Perform health check or sync with server
			s.logf.Info("Performing daemon health check/sync...") // Changed to Info
			// Here you would add logic to:
			// 1. Authenticate with the server if needed
			// 2. Fetch network configuration
			// 3. Apply WireGuard config via s.engine.ApplyWireGuardConfig()
			// 4. Send heartbeat to server
			// 5. Handle potential re-registration or token refresh
		}
	}
}

// autoConnectLoop attempts to connect to the VPN if an identity and token exist.
func (s *DaemonService) autoConnectLoop(ctx context.Context) {
	s.logf.Info("Auto-connect loop started.")
	
	// Retry mechanism for network availability
	retryInterval := 5 * time.Second
	maxRetries := 12 // Try for 1 minute (12 * 5s)
	
	for i := 0; i < maxRetries; i++ {
		select {
		case <-ctx.Done():
			s.logf.Info("Auto-connect loop context cancelled.")
			return
		case <-time.After(retryInterval):
			id := s.idManager.Get()
			if id.DeviceID != "" {
				// Check if auth token exists in keyring
				_, err := s.config.Keyring.RetrieveAuthToken()
				if err == nil {
					s.logf.Info("Identity and auth token found. Attempting to connect VPN.")
					s.engine.Connect()
					return // Successfully initiated connection
				} else {
					s.logf.Warningf("Auth token not found in keyring: %v. Retrying...", err)
				}
			} else {
				s.logf.Warning("Device not registered. Waiting for registration.")
			}
		}
	}
	s.logf.Error("Max retries reached. Auto-connect failed to establish initial connection.")
}


// setupLocalhostBridgeHandlers configures the HTTP handlers for the local bridge.
func (s *DaemonService) setupLocalhostBridgeHandlers(mux *http.ServeMux) {
	// CORS middleware
	cors := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			
			// Security: Only allow requests from trusted origins
			// 1. No origin (e.g. direct API calls, mobile apps) -> Allow (or restrict if strictly browser-only)
			// 2. Localhost (development)
			// 3. Production domain (app.goconnect.example - needs to be configurable)
			
			allowed := false
			if origin == "" {
				allowed = true // Allow non-browser clients for now
			} else if strings.HasPrefix(origin, "http://localhost") || strings.HasPrefix(origin, "http://127.0.0.1") {
				allowed = true
			} else if strings.Contains(origin, "goconnect") { // Loose check for now, tighten later
				allowed = true
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			} else {
				// Log invalid origin attempt
				s.logf.Warningf("Blocked CORS request from origin: %s", origin)
			}

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next(w, r)
		}
	}

	mux.HandleFunc("/status", cors(func(w http.ResponseWriter, _ *http.Request) {
		id := s.idManager.Get()
		w.Header().Set("Content-Type", "application/json")

		resp := s.engine.GetStatus()

		resp["device"] = map[string]interface{}{
			"registered": id.DeviceID != "",
			"public_key": id.PublicKey,
			"device_id":  id.DeviceID,
		}

		_ = json.NewEncoder(w).Encode(resp)
	}))

	mux.HandleFunc("/register", cors(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Token string `json:"token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Token == "" {
			http.Error(w, "Token required", http.StatusBadRequest)
			return
		}

		hostname, _ := os.Hostname()

		regReq := api.RegisterDeviceRequest{
			Name:      hostname,
			Platform:  runtime.GOOS,
			PubKey:    s.idManager.Get().PublicKey,
			HostName:  hostname,
			OSVersion: system.GetOSVersion(),
			DaemonVer: s.daemonVersion, // Use daemonVersion from service struct
		}

		resp, err := s.apiClient.Register(r.Context(), req.Token, regReq)
		if err != nil {
			s.logf.Error("Registration failed: ", err.Error())
			http.Error(w, fmt.Sprintf("Registration failed: %v", err.Error()), http.StatusInternalServerError)
			return
		}

		if err := s.idManager.Update(resp.ID); err != nil { // Token is no longer stored in identity.Manager
			s.logf.Error("Failed to save identity device ID: ", err.Error())
			http.Error(w, "Failed to save identity device ID", http.StatusInternalServerError)
			return
		}

		// Store the auth token securely in the keyring
		if s.config.Keyring != nil {
			if err := s.config.Keyring.StoreAuthToken(req.Token); err != nil {
				s.logf.Error("Failed to store auth token in keyring: ", err.Error())
				http.Error(w, "Failed to store auth token", http.StatusInternalServerError)
				return
			}
		} else {
			s.logf.Error("Keyring is not initialized, cannot store auth token.")
			http.Error(w, "Keyring not available to store auth token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":    "success",
			"device_id": resp.ID,
		})
	}))

	mux.HandleFunc("/connect", cors(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.engine.Connect()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "connected"})
	}))

	mux.HandleFunc("/disconnect", cors(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.engine.Disconnect()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "disconnected"})
	}))
}

// Service program structure for kardianos/service
type program struct {
	daemon *DaemonService
}

// Start is the service Start method for kardianos/service
func (p *program) Start(s service.Service) error {
	// Get the logger from the service instance and set it to the daemonService
	logger, err := s.Logger(nil)
	if err != nil {
		log.Printf("Failed to get logger from service: %v", err)
		// Fallback to stderr logger if service logger is unavailable
		stdLogger := log.New(os.Stderr, "[goconnect-daemon] ", log.LstdFlags)
		p.daemon.logf = &fallbackServiceLogger{stdLogger}
	} else {
		p.daemon.logf = logger
	}
	return p.daemon.Start(s)
}

// Stop is the service Stop method for kardianos/service
func (p *program) Stop(s service.Service) error {
	// Get the logger from the service instance and set it to the daemonService
	logger, err := s.Logger(nil)
	if err != nil {
		log.Printf("Failed to get logger from service: %v", err)
		// Fallback to stderr logger if service logger is unavailable
		stdLogger := log.New(os.Stderr, "[goconnect-daemon] ", log.LstdFlags)
		p.daemon.logf = &fallbackServiceLogger{stdLogger}
	} else {
		p.daemon.logf = logger
	}
	return p.daemon.Stop(s)
}

// fallbackServiceLogger provides a simple implementation of service.Logger using stdlib log.
type fallbackServiceLogger struct {
	*log.Logger
}

func (l *fallbackServiceLogger) Info(v ...interface{}) error {
	l.Logger.Println("[INFO]", fmt.Sprintln(v...))
	return nil
}

func (l *fallbackServiceLogger) Infof(format string, v ...interface{}) error {
	l.Logger.Printf("[INFO] "+format, v...)
	return nil
}

func (l *fallbackServiceLogger) Warning(v ...interface{}) error {
	l.Logger.Println("[WARNING]", fmt.Sprintln(v...))
	return nil
}

func (l *fallbackServiceLogger) Warningf(format string, v ...interface{}) error {
	l.Logger.Printf("[WARNING] "+format, v...)
	return nil
}

func (l *fallbackServiceLogger) Error(v ...interface{}) error {
	l.Logger.Println("[ERROR]", fmt.Sprintln(v...))
	return nil
}

func (l *fallbackServiceLogger) Errorf(format string, v ...interface{}) error {
	l.Logger.Printf("[ERROR] "+format, v...)
	return nil
}


// RunDaemon sets up and runs the daemon as a service or directly.
func RunDaemon(cfg *config.Config, daemonVersion string, options map[string]interface{}) error {
	svcConfig := &service.Config{
		Name:        "goconnect-daemon",
		DisplayName: "GoConnect Daemon",
		Description: "Manages WireGuard VPN connections for GoConnect.",
	}

	// Override default config with platform-specific options
	if options != nil {
		for k, v := range options {
			svcConfig.Option[k] = v // Corrected: Access Option field directly
		}
	}

	daemonService := NewDaemonService(cfg, daemonVersion)
	prg := &program{daemon: daemonService}
	
	svc, err := service.New(prg, svcConfig)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	// Use the service logger if available, otherwise fallback to stderr
	if service.Interactive() {
		// In interactive mode, log to stderr with the fallback logger
		stdLogger := log.New(os.Stderr, "[goconnect-daemon] ", log.LstdFlags)
		daemonService.logf = &fallbackServiceLogger{stdLogger}
	} else {
		// In service mode, get the logger from svc.Logger()
		logger, err := svc.Logger(nil)
		if err != nil {
			log.Printf("Failed to get logger from service in non-interactive mode: %v", err)
			stdLogger := log.New(os.Stderr, "[goconnect-daemon] ", log.LstdFlags)
			daemonService.logf = &fallbackServiceLogger{stdLogger}
		} else {
			daemonService.logf = logger
		}
	}


	if len(os.Args) > 1 {
		cmd := os.Args[1]
		switch cmd {
		case "install":
			err = svc.Install()
			if err != nil {
				return fmt.Errorf("failed to install service: %w", err)
			}
			fmt.Println("Service installed successfully. Use 'start' to run.")
			return nil
		case "uninstall":
			err = svc.Uninstall()
			if err != nil {
				return fmt.Errorf("failed to uninstall service: %w", err)
			}
			fmt.Println("Service uninstalled.")
			return nil
		case "start":
			err = svc.Start()
			if err != nil {
				return fmt.Errorf("failed to start service: %w", err)
			}
			fmt.Println("Service started.")
			return nil
		case "stop":
			err = svc.Stop()
			if err != nil {
				return fmt.Errorf("failed to stop service: %w", err)
			}
			fmt.Println("Service stopped.")
			return nil
		}
	}

	// Run the service
	err = svc.Run()
	if err != nil {
		return fmt.Errorf("service failed to run: %w", err)
	}
	return nil
}