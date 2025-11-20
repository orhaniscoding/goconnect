package engine

import (
	"context"
	"log"
	"runtime"
	"time"

	"github.com/orhaniscoding/goconnect/client-daemon/internal/api"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/identity"
)

// Engine is the main daemon engine
type Engine struct {
	idMgr     *identity.Manager
	apiClient *api.Client
	stopChan  chan struct{}
	version   string
}

// New creates a new engine
func New(idMgr *identity.Manager, apiClient *api.Client, version string) *Engine {
	return &Engine{
		idMgr:     idMgr,
		apiClient: apiClient,
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
