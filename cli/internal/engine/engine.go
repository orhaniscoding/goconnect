package engine

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/kardianos/service"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/api"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/config"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/identity"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/system"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/wireguard"
)

// Engine is the main daemon engine
type Engine struct {
	config    *config.Config
	idMgr     *identity.Manager
	apiClient *api.Client
	wgClient  *wireguard.Client
	sysConf   system.Configurator
	hostsMgr  *system.HostsManager
	stopChan  chan struct{}
	logf      service.Logger
	paused    bool
	daemonVersion string
}

// NewEngine creates a new engine
func NewEngine(cfg *config.Config, idMgr *identity.Manager, wgClient *wireguard.Client, apiClient *api.Client, logger service.Logger) *Engine {
	if logger == nil {
		stdLogger := log.New(os.Stderr, "[goconnect-engine] ", log.LstdFlags)
		logger = &fallbackServiceLogger{stdLogger}
	}
	
	sysConf := system.NewConfigurator()
	hostsMgr := system.NewHostsManager()

	return &Engine{
		config:    cfg,
		idMgr:     idMgr,
		apiClient: apiClient,
		wgClient:  wgClient,
		sysConf:   sysConf,
		hostsMgr:  hostsMgr,
		stopChan:  make(chan struct{}),
		logf:      logger,
		daemonVersion: "dev", // TODO: Get actual version from build flags
	}
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

// Start starts the engine
func (e *Engine) Start() {
	e.logf.Info("Starting daemon engine...")
	go e.heartbeatLoop()
	go e.configLoop()
}

// Stop stops the engine
func (e *Engine) Stop() {
	e.logf.Info("Stopping daemon engine...")
	close(e.stopChan)
	if e.wgClient != nil {
		if err := e.wgClient.Down(); err != nil {
			e.logf.Error("Failed to bring down WireGuard interface: ", err.Error())
		}
	}
	e.logf.Info("Daemon engine stopped.")
}

// Connect enables the VPN connection
func (e *Engine) Connect() {
	e.logf.Info("Connecting VPN...")
	e.paused = false
	go e.syncConfig()
}

// Disconnect disables the VPN connection
func (e *Engine) Disconnect() {
	e.logf.Info("Disconnecting VPN...")
	e.paused = true
	if e.wgClient != nil {
		if err := e.wgClient.Down(); err != nil {
			e.logf.Error("Failed to clear WireGuard config: ", err.Error())
		} else {
			e.logf.Info("WireGuard interface cleared successfully.")
		}
	}
}

func (e *Engine) configLoop() {
	ticker := time.NewTicker(e.config.Daemon.HealthCheckInterval) // Use config value
	defer ticker.Stop()

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
			e.logf.Info("Config sync loop stopped.")
			return
		}
	}
}

func (e *Engine) syncConfig() {
	id := e.idMgr.Get()
	if id.DeviceID == "" { // Token is now in keyring, not in identity struct
		e.logf.Info("Device not registered, skipping config sync.")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config, err := e.apiClient.GetConfig(ctx, id.DeviceID) // No authToken needed here
	if err != nil {
		e.logf.Error("Config sync failed: ", err.Error())
		return
	}

	e.logf.Infof("Received config: %d peers, %d addresses", len(config.Peers), len(config.Interface.Addresses))

	if e.wgClient != nil {
		if err := e.wgClient.ApplyConfig(config, id.PrivateKey); err != nil {
			e.logf.Error("Failed to apply WireGuard config: ", err.Error())
		} else {
			e.logf.Info("WireGuard configuration applied successfully")

			if err := e.sysConf.ConfigureInterface(e.config.WireGuard.InterfaceName, config.Interface.Addresses, config.Interface.DNS, config.Interface.MTU); err != nil {
				e.logf.Error("Failed to configure network interface: ", err.Error())
			} else {
				e.logf.Info("Network interface configured successfully")
			}

			routes := make([]string, 0)
			for _, peer := range config.Peers {
				routes = append(routes, peer.AllowedIPs...)
			}
			if len(routes) > 0 {
				if err := e.sysConf.AddRoutes(e.config.WireGuard.InterfaceName, routes); err != nil {
					e.logf.Error("Failed to add routes: ", err.Error())
				} else {
					e.logf.Infof("Added %d routes successfully", len(routes))
				}
			}

			hostEntries := make([]system.HostEntry, 0)
			for _, peer := range config.Peers {
				if len(peer.AllowedIPs) > 0 {
					ip := peer.AllowedIPs[0]
					if idx := strings.Index(ip, "/"); idx != -1 {
						ip = ip[:idx]
					}

					if peer.Name != "" {
						hostEntries = append(hostEntries, system.HostEntry{
							IP:       ip,
							Hostname: peer.Name,
						})
					}
					if peer.Hostname != "" && peer.Hostname != peer.Name {
						hostEntries = append(hostEntries, system.HostEntry{
							IP:       ip,
							Hostname: peer.Hostname,
						})
					}
				}
			}

			if len(hostEntries) > 0 {
				if err := e.hostsMgr.UpdateHosts(hostEntries); err != nil {
					e.logf.Error("Failed to update hosts file: ", err.Error())
				} else {
					e.logf.Infof("Updated hosts file with %d entries", len(hostEntries))
				}
			}
		}
	}
}

// GetStatus returns the current engine status
func (e *Engine) GetStatus() map[string]interface{} {
	status := map[string]interface{}{
		"running": true,
		"version": e.daemonVersion,
		"paused":  e.paused,
	}

	if e.wgClient != nil {
		wgStatus, err := e.wgClient.GetStatus()
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
	ticker := time.NewTicker(e.config.Daemon.HealthCheckInterval)
	defer ticker.Stop()

	e.sendHeartbeat()

	for {
		select {
		case <-ticker.C:
			e.sendHeartbeat()
		case <-e.stopChan:
			e.logf.Info("Heartbeat loop stopped.")
			return
		}
	}
}

func (e *Engine) sendHeartbeat() {
	id := e.idMgr.Get()
	if id.DeviceID == "" { // Token is now in keyring, not in identity struct
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := api.HeartbeatRequest{
		DaemonVer: e.daemonVersion,
		OSVersion: system.GetOSVersion(),
	}

	if err := e.apiClient.SendHeartbeat(ctx, id.DeviceID, req); err != nil { // No authToken needed here
		e.logf.Error("Heartbeat failed: ", err.Error())
	}
}