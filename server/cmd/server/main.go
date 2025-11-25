package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/audit"
	"github.com/orhaniscoding/goconnect/server/internal/config"
	"github.com/orhaniscoding/goconnect/server/internal/database"
	"github.com/orhaniscoding/goconnect/server/internal/handler"
	"github.com/orhaniscoding/goconnect/server/internal/metrics"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/orhaniscoding/goconnect/server/internal/service"
	ws "github.com/orhaniscoding/goconnect/server/internal/websocket"
	"github.com/orhaniscoding/goconnect/server/internal/wireguard"
	"github.com/redis/go-redis/v9"
)

var (
	version = "dev"
	commit  = "none"
	date    = "2025-09-22"
	builtBy = "orhaniscoding"
)

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	runMigrations := flag.Bool("migrate", false, "run database migrations and exit")
	usePostgres := flag.Bool("postgres", true, "use PostgreSQL instead of in-memory storage")
	asyncAudit := flag.Bool("audit-async", true, "enable async audit buffering")
	auditQueue := flag.Int("audit-queue", 1024, "audit async queue size")
	auditWorkers := flag.Int("audit-workers", 1, "audit async worker count")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), shutdownSignals()...)
	defer stop()

	// Register metrics early
	metrics.Register()

	if *showVersion {
		fmt.Printf("goconnect-server %s (commit %s, build %s) built by %s\n", version, commit, date, builtBy)
		return
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Warning: Failed to load full config: %v. Using defaults/env for DB.", err)
		// We continue because DB config is loaded separately below, but WireGuard config might be missing.
		// Actually, we should probably fail if we rely on it.
		// But for now, let's initialize an empty config if it fails, to avoid panic.
		cfg = &config.Config{}
	}

	// Database setup
	var db *sql.DB
	if *usePostgres {
		dbConfig := database.LoadConfigFromEnv()
		var err error
		db, err = database.Connect(dbConfig)
		if err != nil {
			log.Fatalf("Failed to connect to PostgreSQL: %v", err)
		}
		defer db.Close()

		fmt.Printf("Connected to PostgreSQL: %s@%s:%s/%s\n", dbConfig.User, dbConfig.Host, dbConfig.Port, dbConfig.DBName)

		// Run migrations if requested
		if *runMigrations {
			migrationsPath := getEnvOrDefault("MIGRATIONS_PATH", "./migrations")
			if err := database.RunMigrations(db, migrationsPath); err != nil {
				log.Fatalf("Failed to run migrations: %v", err)
			}
			fmt.Println("Migrations completed successfully")
			return
		}
	}

	// Initialize repositories
	var networkRepo repository.NetworkRepository
	var idempotencyRepo repository.IdempotencyRepository
	var membershipRepo repository.MembershipRepository
	var joinRepo repository.JoinRequestRepository
	var ipamRepo repository.IPAMRepository
	var userRepo repository.UserRepository
	var tenantRepo repository.TenantRepository
	var deviceRepo repository.DeviceRepository
	var peerRepo repository.PeerRepository
	var chatRepo repository.ChatRepository
	var inviteRepo repository.InviteTokenRepository

	if *usePostgres && db != nil {
		// PostgreSQL repositories
		networkRepo = repository.NewPostgresNetworkRepository(db)
		idempotencyRepo = repository.NewPostgresIdempotencyRepository(db)
		membershipRepo = repository.NewPostgresMembershipRepository(db)
		joinRepo = repository.NewPostgresJoinRequestRepository(db)
		ipamRepo = repository.NewPostgresIPAMRepository(db)
		userRepo = repository.NewPostgresUserRepository(db)
		tenantRepo = repository.NewPostgresTenantRepository(db)
		deviceRepo = repository.NewPostgresDeviceRepository(db)
		peerRepo = repository.NewPostgresPeerRepository(db)
		chatRepo = repository.NewPostgresChatRepository(db)
		inviteRepo = repository.NewInMemoryInviteTokenRepository() // TODO: Postgres implementation
		fmt.Println("Using PostgreSQL repositories")
	} else {
		// In-memory repositories (fallback)
		networkRepo = repository.NewInMemoryNetworkRepository()
		idempotencyRepo = repository.NewInMemoryIdempotencyRepository()
		membershipRepo = repository.NewInMemoryMembershipRepository()
		joinRepo = repository.NewInMemoryJoinRequestRepository()
		ipamRepo = repository.NewInMemoryIPAM()
		userRepo = repository.NewInMemoryUserRepository()
		tenantRepo = repository.NewInMemoryTenantRepository()
		deviceRepo = repository.NewInMemoryDeviceRepository()
		peerRepo = repository.NewInMemoryPeerRepository()
		chatRepo = repository.NewInMemoryChatRepository()
		inviteRepo = repository.NewInMemoryInviteTokenRepository()
		fmt.Println("Using in-memory repositories (no data persistence)")
	}

	// Initialize Redis
	var redisClient *redis.Client
	if cfg.Redis.Host != "" {
		var err error
		redisClient, err = database.NewRedisClient(cfg.Redis)
		if err != nil {
			log.Printf("Warning: Failed to connect to Redis: %v. Token blacklist will be disabled.", err)
		} else {
			fmt.Println("Connected to Redis")
			defer redisClient.Close()
		}
	}

	// Initialize services
	networkService := service.NewNetworkService(networkRepo, idempotencyRepo)
	membershipService := service.NewMembershipService(networkRepo, membershipRepo, joinRepo, idempotencyRepo)
	ipamService := service.NewIPAMService(networkRepo, membershipRepo, ipamRepo)
	authService := service.NewAuthService(userRepo, tenantRepo, redisClient)

	peerProvisioningService := service.NewPeerProvisioningService(peerRepo, deviceRepo, networkRepo, membershipRepo, ipamRepo)
	deviceService := service.NewDeviceService(deviceRepo, userRepo, peerRepo, networkRepo, cfg.WireGuard)
	deviceService.SetPeerProvisioning(peerProvisioningService)
	// Start offline detection (check every 30s, mark offline if unseen for 2m)
	go deviceService.StartOfflineDetection(ctx, 30*time.Second, 2*time.Minute)

	chatService := service.NewChatService(chatRepo, userRepo)

	// Initialize GDPR service
	gdprService := service.NewGDPRService(userRepo, deviceRepo, networkRepo, membershipRepo)

	// Initialize Invite service
	baseURL := getEnvOrDefault("APP_BASE_URL", "https://app.goconnect.example")
	inviteService := service.NewInviteService(inviteRepo, networkRepo, membershipRepo, baseURL)

	// Initialize WireGuard Manager & Sync Service
	var wgManager *wireguard.Manager
	if cfg.WireGuard.PrivateKey != "" {
		var err error
		wgManager, err = wireguard.NewManager(cfg.WireGuard.InterfaceName, cfg.WireGuard.PrivateKey, cfg.WireGuard.Port)
		if err != nil {
			log.Printf("Warning: Failed to initialize WireGuard manager: %v. Peer sync will be disabled.", err)
		} else {
			defer wgManager.Close()
			log.Printf("WireGuard manager initialized for interface %s", cfg.WireGuard.InterfaceName)

			wgSyncService := service.NewWireGuardSyncService(peerRepo, wgManager)
			// Start sync loop in background
			go wgSyncService.StartSyncLoop(ctx, 10*time.Second)

			// Start metrics collection loop
			go func(ctx context.Context) {
				ticker := time.NewTicker(15 * time.Second)
				defer ticker.Stop()
				for {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						if err := wgManager.UpdateMetrics(); err != nil {
							log.Printf("Error updating WireGuard metrics: %v", err)
						}
					}
				}
			}(ctx)
		}
	} else {
		log.Println("Warning: WG_PRIVATE_KEY not set. WireGuard manager disabled.")
	}

	// Auditor selection: default stdout, optionally SQLite-backed if configured via env.
	baseAud := audit.NewStdoutAuditor()
	var aud audit.Auditor = audit.WrapWithMetrics(baseAud, metrics.IncAudit)

	// Environment-driven SQLite auditor
	var sqliteAudRef *audit.SqliteAuditor
	if dsn := strings.TrimSpace(os.Getenv("AUDIT_SQLITE_DSN")); dsn != "" {
		var opts []audit.SqliteOption
		// Optional: comma-separated base64url or standard base64 secrets for hashing (rotation supported)
		if secEnv := strings.TrimSpace(os.Getenv("AUDIT_HASH_SECRETS_B64")); secEnv != "" {
			parts := strings.Split(secEnv, ",")
			var secrets [][]byte
			for _, p := range parts {
				s := strings.TrimSpace(p)
				if s == "" {
					continue
				}
				// try RawURLEncoding first, then StdEncoding
				if b, err := base64.RawURLEncoding.DecodeString(s); err == nil && len(b) > 0 {
					secrets = append(secrets, b)
				} else if b2, err2 := base64.StdEncoding.DecodeString(s); err2 == nil && len(b2) > 0 {
					secrets = append(secrets, b2)
				}
			}
			if len(secrets) > 0 {
				opts = append(opts, audit.WithSqliteHashSecrets(secrets...))
			}
		}
		// Optional retention by rows
		if mr := strings.TrimSpace(os.Getenv("AUDIT_MAX_ROWS")); mr != "" {
			if n, err := strconv.Atoi(mr); err == nil && n > 0 {
				opts = append(opts, audit.WithMaxRows(n))
			}
		}
		// Optional retention by age (seconds)
		if ma := strings.TrimSpace(os.Getenv("AUDIT_MAX_AGE_SECONDS")); ma != "" {
			if n, err := strconv.Atoi(ma); err == nil && n > 0 {
				opts = append(opts, audit.WithMaxAge(time.Duration(n)*time.Second))
			}
		}
		// Optional anchor interval
		if ai := strings.TrimSpace(os.Getenv("AUDIT_ANCHOR_INTERVAL")); ai != "" {
			if n, err := strconv.Atoi(ai); err == nil && n > 0 {
				opts = append(opts, audit.WithAnchorInterval(n))
			}
		}
		// Optional Ed25519 signing key for integrity export
		if sk := strings.TrimSpace(os.Getenv("AUDIT_SIGNING_KEY_ED25519_B64")); sk != "" {
			// try url-safe, then standard
			if b, err := base64.RawURLEncoding.DecodeString(sk); err == nil && len(b) == 64 {
				if kid := strings.TrimSpace(os.Getenv("AUDIT_SIGNING_KID")); kid != "" {
					opts = append(opts, audit.WithIntegritySigningKeyID(kid, b))
				} else {
					opts = append(opts, audit.WithIntegritySigningKey(b))
				}
			} else if b2, err2 := base64.StdEncoding.DecodeString(sk); err2 == nil && len(b2) == 64 {
				if kid := strings.TrimSpace(os.Getenv("AUDIT_SIGNING_KID")); kid != "" {
					opts = append(opts, audit.WithIntegritySigningKeyID(kid, b2))
				} else {
					opts = append(opts, audit.WithIntegritySigningKey(b2))
				}
			}
		}

		if sqliteAud, err := audit.NewSqliteAuditor(dsn, opts...); err == nil {
			// Replace base auditor with sqlite-backed one (wrapped with metrics below already)
			aud = audit.WrapWithMetrics(sqliteAud, metrics.IncAudit)
			sqliteAudRef = sqliteAud
			// If async requested, it will wrap this below
		}
	}
	if sqliteAudRef != nil {
		defer sqliteAudRef.Close()
	}
	var asyncAuditor *audit.AsyncAuditor
	if *asyncAudit {
		asyncAuditor = audit.NewAsyncAuditor(aud, audit.WithQueueSize(*auditQueue), audit.WithWorkers(*auditWorkers))
		aud = asyncAuditor
	}
	defer func() {
		if asyncAuditor != nil {
			asyncAuditor.Close()
		}
	}()
	membershipService.SetAuditor(aud)
	networkService.SetAuditor(aud)
	ipamService.SetAuditor(aud)
	deviceService.SetAuditor(aud)
	chatService.SetAuditor(aud)
	inviteService.SetAuditor(aud)

	// Initialize OIDC Service
	oidcService, err := service.NewOIDCService(ctx)
	if err != nil {
		log.Printf("Warning: Failed to initialize OIDC service: %v", err)
	}

	// Initialize handlers
	networkHandler := handler.NewNetworkHandler(networkService, membershipService, deviceService, peerRepo, cfg.WireGuard).WithIPAM(ipamService)
	authHandler := handler.NewAuthHandler(authService, oidcService)
	deviceHandler := handler.NewDeviceHandler(deviceService)
	chatHandler := handler.NewChatHandler(chatService)
	uploadHandler := handler.NewUploadHandler("./uploads", "/uploads")
	gdprHandler := handler.NewGDPRHandler(gdprService, aud)
	inviteHandler := handler.NewInviteHandler(inviteService)

	// Initialize WebSocket components
	// Circular dependency resolution: Handler -> Hub -> Handler
	wsMsgHandler := ws.NewDefaultMessageHandler(nil, chatService, membershipService, deviceService, authService)
	hub := ws.NewHub(wsMsgHandler)
	wsMsgHandler.SetHub(hub)

	// Start Hub
	go hub.Run(ctx)

	// Initialize Admin Service
	adminService := service.NewAdminService(userRepo, tenantRepo, networkRepo, deviceRepo, chatRepo, hub.GetActiveConnectionCount)
	adminHandler := handler.NewAdminHandler(adminService)

	// Initialize WebSocket HTTP handler
	webSocketHandler := handler.NewWebSocketHandler(hub)

	// Initialize rate limit store
	rateLimits := handler.LoadEndpointRateLimitsFromEnv()
	rateLimitStore := handler.NewRateLimitStore(rateLimits)

	// Setup router
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(metrics.GinMiddleware())

	// Register basic routes
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true, "service": "goconnect-server"})
	})

	// Register auth routes with rate limiting
	// Auth routes are rate limited to prevent brute force attacks
	authGroup := r.Group("/v1/auth")
	authGroup.Use(rateLimitStore.AuthRateLimit())
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.Refresh)
		authGroup.GET("/oidc/login", authHandler.LoginOIDC)
		authGroup.GET("/oidc/callback", authHandler.CallbackOIDC)
	}
	// Protected auth routes (no auth rate limit, use default)
	authProtected := r.Group("/v1/auth")
	authProtected.Use(handler.AuthMiddleware(authService))
	{
		authProtected.POST("/password", authHandler.ChangePassword)
		authProtected.GET("/me", authHandler.Me)
		authProtected.POST("/2fa/generate", authHandler.Generate2FA)
		authProtected.POST("/2fa/enable", authHandler.Enable2FA)
		authProtected.POST("/2fa/disable", authHandler.Disable2FA)
	}

	// Register network routes (auth + role middleware applied within)
	handler.RegisterNetworkRoutes(r, networkHandler, authService, membershipRepo)

	// Register device routes
	handler.RegisterDeviceRoutes(r, deviceHandler, handler.AuthMiddleware(authService))

	// Register chat routes with rate limiting
	chatGroup := r.Group("/v1/chat")
	chatGroup.Use(handler.AuthMiddleware(authService))
	chatGroup.Use(rateLimitStore.ChatRateLimit()) // Rate limit chat operations
	{
		chatGroup.GET("", chatHandler.ListMessages)
		chatGroup.POST("", chatHandler.SendMessage)
		chatGroup.GET("/:id", chatHandler.GetMessage)
		chatGroup.PATCH("/:id", chatHandler.EditMessage)
		chatGroup.DELETE("/:id", chatHandler.DeleteMessage)
		chatGroup.GET("/:id/edits", chatHandler.GetEditHistory)
		chatGroup.POST("/:id/redact", handler.RequireModerator(), chatHandler.RedactMessage)
	}

	// Register upload routes
	handler.RegisterUploadRoutes(r, uploadHandler, handler.AuthMiddleware(authService))
	r.Static("/uploads", "./uploads")

	// Register GDPR routes (DSR - Data Subject Rights)
	meGroup := r.Group("/v1/me")
	meGroup.Use(handler.AuthMiddleware(authService))
	{
		meGroup.GET("/export", gdprHandler.ExportData)
		meGroup.GET("/export/download", gdprHandler.ExportDataDownload)
		meGroup.DELETE("/delete", gdprHandler.RequestDeletion)
	}

	// Register invite routes with rate limiting
	// Public validation endpoint (no rate limit needed)
	r.GET("/v1/invites/:token/validate", inviteHandler.ValidateInvite)

	// Protected invite management endpoints with rate limiting
	inviteGroup := r.Group("/v1/networks/:id/invites")
	inviteGroup.Use(handler.AuthMiddleware(authService))
	inviteGroup.Use(rateLimitStore.InviteRateLimit())
	{
		inviteGroup.POST("", inviteHandler.CreateInvite)
		inviteGroup.GET("", inviteHandler.ListInvites)
		inviteGroup.GET("/:invite_id", inviteHandler.GetInvite)
		inviteGroup.DELETE("/:invite_id", inviteHandler.RevokeInvite)
	}

	// Register admin routes
	adminGroup := r.Group("/v1/admin")
	adminGroup.Use(handler.AuthMiddleware(authService))
	adminGroup.Use(handler.RequireAdmin())
	{
		adminGroup.GET("/users", adminHandler.ListUsers)
		adminGroup.POST("/users/:id/toggle-admin", adminHandler.ToggleUserAdmin)
		adminGroup.DELETE("/users/:id", adminHandler.DeleteUser)
		adminGroup.GET("/tenants", adminHandler.ListTenants)
		adminGroup.DELETE("/tenants/:id", adminHandler.DeleteTenant)
		adminGroup.GET("/networks", adminHandler.ListNetworks)
		adminGroup.GET("/devices", adminHandler.ListDevices)
		adminGroup.GET("/stats", adminHandler.GetSystemStats)
	}

	// Register WebSocket route
	r.GET("/ws", handler.AuthMiddleware(authService), webSocketHandler.HandleUpgrade)

	// Metrics endpoint
	r.GET("/metrics", metrics.Handler())

	// Audit integrity export (best-effort) – future: restrict via RBAC
	// Pass through the underlying sqlite auditor if present so ExportIntegrity works.
	if sqliteAudRef != nil {
		r.GET("/v1/audit/integrity", handler.AuditIntegrityHandler(sqliteAudRef))
		r.GET("/v1/audit/logs", handler.AuthMiddleware(authService), handler.AuditListHandler(sqliteAudRef))
	} else {
		r.GET("/v1/audit/integrity", handler.AuditIntegrityHandler(aud))
		r.GET("/v1/audit/logs", handler.AuthMiddleware(authService), handler.AuditListHandler(aud))
	}

	// Start server with timeouts
	srv := &http.Server{
		Addr:              serverAddress(cfg),
		Handler:           r,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		<-ctx.Done()
		fmt.Println("Shutdown signal received. Draining HTTP server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Error during graceful shutdown: %v\n", err)
		}
	}()

	fmt.Printf("GoConnect Server starting on %s...\n", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}

func serverAddress(cfg *config.Config) string {
	const defaultPort = "8080"
	if cfg == nil {
		return ":" + defaultPort
	}
	host := cfg.Server.Host
	port := cfg.Server.Port
	if port == "" {
		port = defaultPort
	}
	if host == "" {
		return ":" + port
	}
	return host + ":" + port
}

func shutdownSignals() []os.Signal {
	return []os.Signal{os.Interrupt, syscall.Signal(15)}
}

// getEnvOrDefault gets an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
