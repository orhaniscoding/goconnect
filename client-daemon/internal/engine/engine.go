package engine

import (
	"context"
	"log"
	"runtime"
	"time"

	"github.com/orhaniscoding/goconnect/client-daemon/internal/api"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/identity"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/system"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/wireguard"
)

// Engine is the main daemon engine
type Engine struct {
	idMgr     *identity.Manager
	apiClient *api.Client
	wgMgr     *wireguard.Manager
	sysConf   system.Configurator
	stopChan  chan struct{}
	version   string
}

// New creates a new engine
func New(idMgr *identity.Manager, apiClient *api.Client, version string) *Engine {
	sysConf := system.NewConfigurator()

	// Ensure interface exists (Linux only mostly)
	if err := sysConf.EnsureInterface("wg0"); err != nil {
		log.Printf("Warning: Failed to ensure interface wg0: %v", err)
	}

	// Initialize WireGuard manager
	// TODO: Make interface name configurable
	wgMgr, err := wireguard.NewManager("wg0")
	if err != nil {
		log.Printf("Warning: Failed to initialize WireGuard manager: %v", err)
	}

	return &Engine{
		idMgr:     idMgr,
		apiClient: apiClient,
		wgMgr:     wgMgr,
		sysConf:   sysConf,
		stopChan:  make(chan struct{}),
		version:   version,
	}
}

// Start starts the engine
func (e *Engine) Start() {
	log.Println("Starting daemon engine...")
	go e.heartbeatLoop()
	go e.configLoop()
}

// Stop stops the engine
func (e *Engine) Stop() {
	close(e.stopChan)
}

func (e *Engine) configLoop() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	// Initial sync
	e.syncConfig()

	for {
		select {
		case <-ticker.C:
			e.syncConfig()
		case <-e.stopChan:
			return
		}
	}
}

func (e *Engine) syncConfig() {
	id := e.idMgr.Get()
	if id.DeviceID == "" || id.Token == "" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config, err := e.apiClient.GetConfig(ctx, id.DeviceID, id.Token)
	if err != nil {
		log.Printf("Config sync failed: %v", err)
		return
	}

	// TODO: Apply config to WireGuard interface
	log.Printf("Received config: %d peers, %v addresses", len(config.Peers), config.Interface.Addresses)

	if e.wgMgr != nil {
		if err := e.wgMgr.ApplyConfig(config, id.PrivateKey); err != nil {
			log.Printf("Failed to apply WireGuard config: %v", err)
		} else {
			log.Println("WireGuard configuration applied successfully")

			// Apply system network configuration (IPs, MTU)
			if err := e.sysConf.ConfigureInterface("wg0", config.Interface.Addresses, config.Interface.DNS, config.Interface.MTU); err != nil {
				log.Printf("Failed to configure network interface: %v", err)
			} else {
				log.Println("Network interface configured successfully")
			}

			// Apply routes
			routes := make([]string, 0)
			for _, peer := range config.Peers {
				routes = append(routes, peer.AllowedIPs...)
			}
			if len(routes) > 0 {
				if err := e.sysConf.AddRoutes("wg0", routes); err != nil {
					log.Printf("Failed to add routes: %v", err)
				} else {
					log.Printf("Added %d routes successfully", len(routes))
				}
			}
		}
	}
}

func (e *Engine) heartbeatLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Initial heartbeat
	e.sendHeartbeat()

	for {
		select {
		case <-ticker.C:
			e.sendHeartbeat()
		case <-e.stopChan:
			return
		}
	}
}

func (e *Engine) sendHeartbeat() {
	id := e.idMgr.Get()
	if id.DeviceID == "" || id.Token == "" {
		// Not registered yet, skip
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := api.HeartbeatRequest{
		DaemonVer: e.version,
		OSVersion: runtime.GOOS, // TODO: Get real OS version
	}

	if err := e.apiClient.SendHeartbeat(ctx, id.DeviceID, id.Token, req); err != nil {
		log.Printf("Heartbeat failed: %v", err)
	} else {
		// log.Println("Heartbeat sent successfully") // Verbose
	}
}
