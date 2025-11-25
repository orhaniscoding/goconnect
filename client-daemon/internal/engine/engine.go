package engine

import (
	"context"
	"log"
	"os"
	"strings"
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
	hostsMgr  *system.HostsManager
	stopChan  chan struct{}
	version   string
	paused    bool
}

// New creates a new engine
func New(idMgr *identity.Manager, apiClient *api.Client, version string) *Engine {
	sysConf := system.NewConfigurator()
	hostsMgr := system.NewHostsManager()

	// Get interface name from env or default to wg0
	ifaceName := os.Getenv("GOCONNECT_INTERFACE")
	if ifaceName == "" {
		ifaceName = "wg0"
	}

	// Ensure interface exists (Linux only mostly)
	if err := sysConf.EnsureInterface(ifaceName); err != nil {
		log.Printf("Warning: Failed to ensure interface %s: %v", ifaceName, err)
	}

	// Initialize WireGuard manager
	wgMgr, err := wireguard.NewManager(ifaceName)
	if err != nil {
		log.Printf("Warning: Failed to initialize WireGuard manager: %v", err)
	}

	return &Engine{
		idMgr:     idMgr,
		apiClient: apiClient,
		wgMgr:     wgMgr,
		sysConf:   sysConf,
		hostsMgr:  hostsMgr,
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

// Connect enables the VPN connection
func (e *Engine) Connect() {
	e.paused = false
	go e.syncConfig()
}

// Disconnect disables the VPN connection
func (e *Engine) Disconnect() {
	e.paused = true
	// Clear WireGuard config (remove all peers)
	if e.wgMgr != nil {
		// Create empty config to clear peers
		emptyConfig := &api.DeviceConfig{
			Peers: []api.PeerConfig{},
			Interface: api.InterfaceConfig{
				// Keep current listen port or default
				ListenPort: 51820,
			},
		}
		id := e.idMgr.Get()
		if err := e.wgMgr.ApplyConfig(emptyConfig, id.PrivateKey); err != nil {
			log.Printf("Failed to clear WireGuard config: %v", err)
		}
	}
}

func (e *Engine) configLoop() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	// Initial sync
	if !e.paused {
		e.syncConfig()
	}

	for {
		select {
		case <-ticker.C:
			if !e.paused {
				e.syncConfig()
			}
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

			// Update hosts file (MagicDNS Lite)
			hostEntries := make([]system.HostEntry, 0)
			for _, peer := range config.Peers {
				if len(peer.AllowedIPs) > 0 {
					// Use the first allowed IP (usually the device IP)
					// Strip CIDR mask if present
					ip := peer.AllowedIPs[0]
					if idx := strings.Index(ip, "/"); idx != -1 {
						ip = ip[:idx]
					}

					if peer.Name != "" {
						hostEntries = append(hostEntries, system.HostEntry{
							IP:       ip,
							Hostname: peer.Name, // Use friendly name
						})
					}
					if peer.Hostname != "" && peer.Hostname != peer.Name {
						hostEntries = append(hostEntries, system.HostEntry{
							IP:       ip,
							Hostname: peer.Hostname, // Use actual hostname
						})
					}
				}
			}

			if len(hostEntries) > 0 {
				if err := e.hostsMgr.UpdateHosts(hostEntries); err != nil {
					log.Printf("Failed to update hosts file: %v", err)
				} else {
					log.Printf("Updated hosts file with %d entries", len(hostEntries))
				}
			}
		}
	}
}

// GetStatus returns the current engine status
func (e *Engine) GetStatus() map[string]interface{} {
	status := map[string]interface{}{
		"running": true,
		"version": e.version,
		"paused":  e.paused,
	}

	if e.wgMgr != nil {
		wgStatus, err := e.wgMgr.GetStatus()
		if err == nil {
			status["wg"] = wgStatus
		} else {
			status["wg"] = map[string]interface{}{"active": false, "error": err.Error()}
		}
	} else {
		status["wg"] = map[string]interface{}{"active": false}
	}

	return status
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
		OSVersion: system.GetOSVersion(),
	}

	if err := e.apiClient.SendHeartbeat(ctx, id.DeviceID, id.Token, req); err != nil {
		log.Printf("Heartbeat failed: %v", err)
	}
}
