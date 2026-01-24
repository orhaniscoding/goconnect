package daemon

import (
	"context"
	"net"
	"runtime"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/backend"
	"github.com/orhaniscoding/goconnect/server/internal/auth"
	"github.com/orhaniscoding/goconnect/server/internal/logger"
	"github.com/orhaniscoding/goconnect/server/internal/rpc"
	"github.com/orhaniscoding/goconnect/server/internal/proto"
)

type Daemon struct {
	ctx      context.Context
	cancel   context.CancelFunc
	Shutdown *ShutdownManager
	Identity auth.KeyManager
	RPC      *rpc.Server
	Listener net.Listener
	TokenManager auth.TokenManager
	Backend      *backend.Client
	Version      string
}

// New creates a new Daemon instance with a context and shutdown manager.
func New(keyMgr auth.KeyManager, tokenMgr auth.TokenManager, version string, backendURL string) *Daemon {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &Daemon{
		ctx:          ctx,
		cancel:       cancel,
		Shutdown:     NewShutdownManager(),
		Identity:     keyMgr,
		TokenManager: tokenMgr,
		Backend:      backend.NewClient(backendURL),
		Version:      version,
	}
}

// Run starts the daemon and blocks until the context is canceled.
func (d *Daemon) Run() error {
	if err := d.Identity.LoadOrGenerate(); err != nil {
		logger.Error("failed to load identity", "error", err)
		return err
	}

	logger.Info("Daemon Started",
		"version", d.Version,
		"os", runtime.GOOS,
		"arch", runtime.GOARCH,
		"public_key", d.Identity.GetPublicKey().String(),
	)

	// Start Session Maintenance
	d.StartSessionMaintenance()

	// Start RPC Server if listener is provided
	if d.Listener != nil {
		d.RPC = rpc.NewServer()

		// Register handlers
		daemonHandler := rpc.NewDaemonHandler(d, d.Version)
		proto.RegisterDaemonServiceServer(d.RPC.GetGRPCServer(), daemonHandler)

		// Register NetworkService handler
		networkHandler := rpc.NewNetworkServiceHandler(d.Backend, d.TokenManager)
		proto.RegisterNetworkServiceServer(d.RPC.GetGRPCServer(), networkHandler)
	}

	if d.RPC != nil && d.Listener != nil {
		go func() {
			if err := d.RPC.Start(d.Listener); err != nil {
				logger.Error("RPC server failed", "error", err)
			}
		}()
	}

	// Block until context is canceled (by Stop() or parent)
	<-d.ctx.Done()
	
	logger.Info("Shutting down")

	// Create a timeout context for cleanup
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if d.RPC != nil {
		d.RPC.GracefulStop()
	}

	d.Shutdown.Execute(shutdownCtx)
	
	logger.Info("Daemon exit complete")
	return nil
}


// Stop signals the daemon to shut down.
func (d *Daemon) Stop() {
	d.cancel()
}

// GetBackend returns the backend client interface for RPC.
func (d *Daemon) GetBackend() interface {
	RequestDeviceCode(context.Context) (*backend.DeviceCodeResponse, error)
	PollDeviceToken(context.Context, string) (*backend.AuthResponse, error)
} {
	return d.Backend
}

// GetTokenManager returns the token manager.
func (d *Daemon) GetTokenManager() auth.TokenManager {
	return d.TokenManager
}

// StartSessionMaintenance starts a background routine to refresh tokens.
func (d *Daemon) StartSessionMaintenance() {
	go func() {
		// Initial check
		d.checkSession()

		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-d.ctx.Done():
				return
			case <-ticker.C:
				d.checkSession()
			}
		}
	}()
}

func (d *Daemon) checkSession() {
	session, err := d.TokenManager.LoadSession()
	if err != nil {
		// No session or error (e.g. not logged in)
		return
	}

	// Refresh if expiring in < 10 mins
	// If expiry is zero (legacy/unknown), we might want to refresh anyway or assume it's valid?
	// For now, only refresh if we have an expiry and it's close.
	if session.Expiry.IsZero() || time.Until(session.Expiry) < 10*time.Minute {
		logger.Info("Refreshing session", "expiry", session.Expiry)
		
		// Create a separate context with timeout for the refresh call
		ctx, cancel := context.WithTimeout(d.ctx, 30*time.Second)
		defer cancel()

		resp, err := d.Backend.RefreshToken(ctx, session.RefreshToken)
		if err != nil {
			logger.Error("Failed to refresh token", "error", err)
			return
		}

		newSession := &auth.TokenSession{
			AccessToken:  resp.AccessToken,
			RefreshToken: resp.RefreshToken,
			Expiry:       time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second),
		}
		
		if err := d.TokenManager.SaveSession(newSession); err != nil {
			logger.Error("Failed to save refreshed session", "error", err)
		} else {
			logger.Info("Session refreshed successfully")
		}
	}
}

