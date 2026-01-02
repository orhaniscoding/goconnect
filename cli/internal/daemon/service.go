package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/kardianos/service"
	"github.com/orhaniscoding/goconnect/cli/internal/api"
	"github.com/orhaniscoding/goconnect/cli/internal/chat"
	"github.com/orhaniscoding/goconnect/cli/internal/config"
	"github.com/orhaniscoding/goconnect/cli/internal/engine"
	"github.com/orhaniscoding/goconnect/cli/internal/identity"
	"github.com/orhaniscoding/goconnect/cli/internal/logger"
	"github.com/orhaniscoding/goconnect/cli/internal/system"
	"github.com/orhaniscoding/goconnect/cli/internal/transfer"
	"github.com/orhaniscoding/goconnect/cli/internal/wireguard"
)

// DaemonService implements the kardianos/service.Service interface.
type DaemonService struct {
	config          *config.Config
	engine          DaemonEngine
	logf            service.Logger
	cancel          context.CancelFunc
	localHTTPServer *http.Server
	grpcServer      *GRPCServer
	idManager       *identity.Manager
	apiClient       DaemonAPIClient
	daemonVersion   string
	buildDate       string
	commit          string

	// SSE fields
	sseClients map[chan string]bool
	sseMu      sync.RWMutex

	// Shutdown synchronization
	wg sync.WaitGroup
}

var (
	autoConnectRetryInterval = 5 * time.Second
	shutdownTimeout          = 5 * time.Second
)

// NewDaemonService creates a new DaemonService.
func NewDaemonService(cfg *config.Config, daemonVersion string) *DaemonService {
	return &DaemonService{
		config:        cfg,
		daemonVersion: daemonVersion,
		sseClients:    make(map[chan string]bool),
	}
}

// NewDaemonServiceWithBuildInfo creates a new DaemonService with build info.
func NewDaemonServiceWithBuildInfo(cfg *config.Config, version, buildDate, commit string) *DaemonService {
	return &DaemonService{
		config:        cfg,
		daemonVersion: version,
		buildDate:     buildDate,
		commit:        commit,
		sseClients:    make(map[chan string]bool),
	}
}

// Start is called when the service is starting.
func (s *DaemonService) Start(srv service.Service) error {
	var ctx context.Context
	ctx, s.cancel = context.WithCancel(context.Background())

	// Initialize logger (already set in RunDaemon for interactive/service mode)
	if s.logf == nil { // Fallback, should ideally not happen if RunDaemon sets it
		s.logf = &fallbackServiceLogger{}
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
	concreteClient := api.NewClient(s.config)
	s.apiClient = concreteClient

	// Initialize Engine
	s.engine, err = engine.NewEngine(s.config, s.idManager, wgClient, concreteClient, s.logf)
	if err != nil {
		s.logf.Error("Failed to initialize engine: ", err.Error())
		return err
	}

	// Setup Chat Callback for SSE
	s.engine.SetOnChatMessage(func(msg chat.Message) {
		// Broadcast to SSE clients
		payload := map[string]interface{}{
			"type": "chat_message",
			"data": msg,
		}
		jsonBytes, err := json.Marshal(payload)
		if err == nil {
			s.broadcastSSE(string(jsonBytes))
		}
	})

	// Setup Transfer Callbacks for SSE
	s.engine.SetTransferCallbacks(func(session transfer.Session) {
		payload := map[string]interface{}{
			"type": "file_progress",
			"data": session,
		}
		jsonBytes, err := json.Marshal(payload)
		if err == nil {
			s.broadcastSSE(string(jsonBytes))
		}
	}, func(req transfer.Request, senderID string) {
		payload := map[string]interface{}{
			"type": "file_request",
			"data": map[string]interface{}{
				"request":   req,
				"sender_id": senderID,
			},
		}
		jsonBytes, err := json.Marshal(payload)
		if err == nil {
			s.broadcastSSE(string(jsonBytes))
		}
	})

	s.engine.Start()

	// Setup Localhost Bridge Handlers
	mux := http.NewServeMux() // Use a new ServeMux for the local HTTP server
	s.setupLocalhostBridgeHandlers(mux)
	s.localHTTPServer = &http.Server{
		Addr:              fmt.Sprintf("127.0.0.1:%d", s.config.Daemon.LocalPort),
		Handler:           mux, // Assign the mux to the server
		ReadTimeout:       0,   // Disable read timeout for SSE
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      0, // Disable write timeout for SSE
		IdleTimeout:       0, // Disable idle timeout for SSE
	}

	go func() {
		s.logf.Info("Daemon bridge at http://", s.localHTTPServer.Addr)
		if err := s.localHTTPServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logf.Error("Localhost bridge failed: ", err.Error())
		}
	}()

	// Start gRPC server
	s.grpcServer = NewGRPCServer(s, s.daemonVersion, s.buildDate, s.commit)
	if err := s.grpcServer.Start(ctx); err != nil {
		s.logf.Error("Failed to start gRPC server: ", err.Error())
		// Continue without gRPC - HTTP bridge still works
	}

	// Auto-connect logic
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.autoConnectLoop(ctx)
	}()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.run(ctx)
	}()

	return nil
}

// Stop is called when the service is stopping.
func (s *DaemonService) Stop(srk service.Service) error {
	s.logf.Info("GoConnect Daemon stopping...")
	if s.cancel != nil {
		s.cancel()
	}
	if s.grpcServer != nil {
		s.grpcServer.Stop()
	}
	if s.engine != nil {
		s.engine.Stop()
	}
	if s.localHTTPServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_ = s.localHTTPServer.Shutdown(ctx)
		cancel()
	}

	// Wait for goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.logf.Info("Daemon goroutines exited cleanly.")
	case <-time.After(shutdownTimeout):
		s.logf.Warning("Daemon shutdown timed out. Force exiting.")
	}

	s.logf.Info("GoConnect Daemon stopped.")
	return nil
}

// run contains the main logic of the daemon.
func (s *DaemonService) run(ctx context.Context) {
	s.logf.Info("Daemon main loop started.")

	interval := s.config.Daemon.HealthCheckInterval
	if interval < time.Second {
		interval = 5 * time.Second // Safety default
	}

	// Example: Periodically check server connection or trigger updates
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	lastTick := time.Now()

	for {
		select {
		case <-ctx.Done():
			s.logf.Info("Daemon run loop context cancelled.")
			return
		case now := <-ticker.C:
			// Detect time jump (system sleep/resume)
			// If time since last tick is significantly larger than interval (e.g. 3x),
			// we likely slept.
			timeSinceLast := now.Sub(lastTick)
			if timeSinceLast > interval*3 {
				s.logf.Warningf("System resume detected? Time jump: %v. Triggering immediate reconnect/sync.", timeSinceLast)

				// Force connection check/reconnect
				if s.engine != nil {
					// If we were connected, ensure we still are or reconnect
					s.logf.Info("Refreshing connection state after resume...")
					s.engine.Connect() // Or a more specific "Reconnect()" if available, but Connect() usually handles idempotency
				}
			}
			lastTick = now

			// Perform health check or sync with server
			s.logf.Info("Performing daemon health check/sync...")
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
	retryInterval := autoConnectRetryInterval
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

// broadcastSSE sends a message to all connected SSE clients
func (s *DaemonService) broadcastSSE(message string) {
	s.sseMu.RLock()
	defer s.sseMu.RUnlock()
	for clientChan := range s.sseClients {
		// Non-blocking send
		select {
		case clientChan <- message:
		default:
			// Client might be slow or disconnected, skip
		}
	}
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
			} else if strings.HasPrefix(origin, "tauri://") {
				// Tauri desktop app uses tauri:// protocol
				allowed = true
			} else if s.config.Server.URL != "" {
				// Allow requests from the configured server domain
				// e.g., if server is "https://vpn.example.com", allow that origin
				if strings.HasPrefix(origin, s.config.Server.URL) {
					allowed = true
				}
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
			Name  string `json:"name"`
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
		displayName := req.Name
		if displayName == "" {
			displayName = hostname
		}

		regReq := api.RegisterDeviceRequest{
			Name:      displayName,
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

	mux.HandleFunc("/connect", cors(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.engine.Connect()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "connected"})
	}))

	mux.HandleFunc("/p2p/connect", cors(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			PeerID string `json:"peer_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.PeerID == "" {
			http.Error(w, "Peer ID required", http.StatusBadRequest)
			return
		}

		if err := s.engine.ManualConnect(req.PeerID); err != nil {
			s.logf.Errorf("Manual P2P connect failed: %v", err)
			http.Error(w, fmt.Sprintf("Failed to connect: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":  "initiated",
			"peer_id": req.PeerID,
		})
	}))

	mux.HandleFunc("/config", cors(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(s.config)
			return
		}

		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			P2PEnabled *bool   `json:"p2p_enabled"`
			StunServer *string `json:"stun_server"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.P2PEnabled != nil {
			s.config.P2P.Enabled = *req.P2PEnabled
			s.logf.Infof("Updated P2P enabled status to: %v", *req.P2PEnabled)
		}
		if req.StunServer != nil {
			s.config.P2P.StunServer = *req.StunServer
			s.logf.Infof("Updated STUN server to: %s", *req.StunServer)
		}

		// Save config
		if err := s.config.Save(config.DefaultConfigPath()); err != nil {
			s.logf.Errorf("Failed to save config: %v", err)
			http.Error(w, "Failed to save configuration", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "updated",
			"config": s.config,
		})
	}))

	mux.HandleFunc("/chat/send", cors(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			PeerID  string `json:"peer_id"`
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.PeerID == "" || req.Content == "" {
			http.Error(w, "Peer ID and Content required", http.StatusBadRequest)
			return
		}

		if err := s.engine.SendChatMessage(req.PeerID, req.Content); err != nil {
			s.logf.Errorf("Failed to send chat message: %v", err)
			http.Error(w, fmt.Sprintf("Failed to send message: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "sent"})
	}))

	mux.HandleFunc("/file/send", cors(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			PeerID   string `json:"peer_id"`
			FilePath string `json:"file_path"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		session, err := s.engine.SendFileRequest(req.PeerID, req.FilePath)
		if err != nil {
			s.logf.Errorf("Failed to send file request: %v", err)
			http.Error(w, fmt.Sprintf("Failed to send file request: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(session)
	}))

	mux.HandleFunc("/file/accept", cors(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Request  transfer.Request `json:"request"`
			PeerID   string           `json:"peer_id"`
			SavePath string           `json:"save_path"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := s.engine.AcceptFile(req.Request.ID, req.SavePath); err != nil {
			s.logf.Errorf("Failed to accept file: %v", err)
			http.Error(w, fmt.Sprintf("Failed to accept file: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "accepted"})
	}))

	mux.HandleFunc("/networks", cors(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		networks, err := s.apiClient.GetNetworks(r.Context())
		if err != nil {
			s.logf.Errorf("Failed to get networks: %v", err)
			http.Error(w, fmt.Sprintf("Failed to get networks: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(networks)
	}))

	mux.HandleFunc("/networks/create", cors(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Name == "" {
			http.Error(w, "Network name required", http.StatusBadRequest)
			return
		}

		network, err := s.apiClient.CreateNetwork(r.Context(), req.Name, "")
		if err != nil {
			s.logf.Errorf("Failed to create network: %v", err)
			http.Error(w, fmt.Sprintf("Failed to create network: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(network)
	}))

	mux.HandleFunc("/networks/join", cors(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			InviteCode string `json:"invite_code"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.InviteCode == "" {
			http.Error(w, "Invite code required", http.StatusBadRequest)
			return
		}

		network, err := s.apiClient.JoinNetwork(r.Context(), req.InviteCode)
		if err != nil {
			s.logf.Errorf("Failed to join network: %v", err)
			http.Error(w, fmt.Sprintf("Failed to join network: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(network)
	}))

	// SSE Endpoint
	mux.HandleFunc("/events", cors(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		clientChan := make(chan string)
		s.sseMu.Lock()
		s.sseClients[clientChan] = true
		s.sseMu.Unlock()

		s.logf.Info("New SSE client connected")

		defer func() {
			s.sseMu.Lock()
			delete(s.sseClients, clientChan)
			s.sseMu.Unlock()
			close(clientChan)
			s.logf.Info("SSE client disconnected")
		}()

		// Keep connection open
		for {
			select {
			case msg := <-clientChan:
				fmt.Fprintf(w, "data: %s\n\n", msg)
				w.(http.Flusher).Flush()
			case <-r.Context().Done():
				return
			}
		}
	}))
}

// Service program structure for kardianos/service
type program struct {
	daemon *DaemonService
}

// Start is the service Start method for kardianos/service
func (p *program) Start(s service.Service) error {
	// Get the logger from the service instance and set it to the daemonService
	svcLogger, err := s.Logger(nil)
	if err != nil {
		logger.Warn("Failed to get logger from service", "error", err)
		// Fallback to stderr logger if service logger is unavailable
		p.daemon.logf = &fallbackServiceLogger{}
	} else {
		p.daemon.logf = svcLogger
	}
	return p.daemon.Start(s)
}

// Stop is the service Stop method for kardianos/service
func (p *program) Stop(s service.Service) error {
	// Get the logger from the service instance and set it to the daemonService
	svcLogger, err := s.Logger(nil)
	if err != nil {
		logger.Warn("Failed to get logger from service", "error", err)
		// Fallback to stderr logger if service logger is unavailable
		p.daemon.logf = &fallbackServiceLogger{}
	} else {
		p.daemon.logf = svcLogger
	}
	return p.daemon.Stop(s)
}

// fallbackServiceLogger provides a simple implementation of service.Logger using stdlib log.
type fallbackServiceLogger struct{}

func (l *fallbackServiceLogger) Info(v ...interface{}) error {
	logger.Info(fmt.Sprint(v...))
	return nil
}

func (l *fallbackServiceLogger) Infof(format string, v ...interface{}) error {
	logger.Info(fmt.Sprintf(format, v...))
	return nil
}

func (l *fallbackServiceLogger) Warning(v ...interface{}) error {
	logger.Warn(fmt.Sprint(v...))
	return nil
}

func (l *fallbackServiceLogger) Warningf(format string, v ...interface{}) error {
	logger.Warn(fmt.Sprintf(format, v...))
	return nil
}

func (l *fallbackServiceLogger) Error(v ...interface{}) error {
	logger.Error(fmt.Sprint(v...))
	return nil
}

func (l *fallbackServiceLogger) Errorf(format string, v ...interface{}) error {
	logger.Error(fmt.Sprintf(format, v...))
	return nil
}

// RunDaemon sets up and runs the daemon as a service or directly.
func RunDaemon(cfg *config.Config, daemonVersion string, options map[string]interface{}) error {
	return runDaemonInternal(cfg, daemonVersion, options, service.New, os.Args)
}

// ServiceFactory is a function that creates a new service.
type ServiceFactory func(i service.Interface, c *service.Config) (service.Service, error)

// runDaemonInternal allows testing the daemon lifecycle by injecting dependencies.
func runDaemonInternal(cfg *config.Config, daemonVersion string, options map[string]interface{}, factory ServiceFactory, args []string) error {
	svcConfig := &service.Config{
		Name:        "goconnect",
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

	svc, err := factory(prg, svcConfig)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	// Use the service logger if available, otherwise fallback to stderr
	if service.Interactive() {
		// In interactive mode, log to stderr with the fallback logger
		daemonService.logf = &fallbackServiceLogger{}
	} else {
		// In service mode, get the logger from svc.Logger()
		logger, err := svc.Logger(nil)
		if err != nil {
			slog.Warn("Failed to get logger from service in non-interactive mode", "error", err)
			daemonService.logf = &fallbackServiceLogger{}
		} else {
			daemonService.logf = logger
		}
	}

	if len(args) > 1 {
		cmd := args[1]
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
