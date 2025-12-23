package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/audit"
	"github.com/orhaniscoding/goconnect/server/internal/config"
	"github.com/orhaniscoding/goconnect/server/internal/database"
	"github.com/orhaniscoding/goconnect/server/internal/handler"
	"github.com/orhaniscoding/goconnect/server/internal/logger"
	"github.com/orhaniscoding/goconnect/server/internal/metrics"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/orhaniscoding/goconnect/server/internal/service"
	"github.com/orhaniscoding/goconnect/server/internal/websocket"
	"github.com/orhaniscoding/goconnect/server/internal/wireguard"
	"github.com/redis/go-redis/v9"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	// Parse command-line flags
	showVersion := flag.Bool("version", false, "Print version information and exit")
	configPath := flag.String("config", "", "Path to configuration file (optional)")
	runMigrate := flag.Bool("migrate", false, "Run database migrations and exit")
	autoMigrate := flag.Bool("auto-migrate", false, "Run database migrations before starting server")
	flag.Parse()

	if *showVersion {
		fmt.Printf("goconnect-server %s\n", version)
		fmt.Printf("  commit:  %s\n", commit)
		fmt.Printf("  built:   %s\n", date)
		fmt.Printf("  builder: %s\n", builtBy)
		return
	}

	// Load configuration
	var cfg *config.Config
	var err error

	if *configPath != "" {
		cfg, err = config.LoadFromFile(*configPath)
	} else {
		cfg, err = config.LoadFromFileOrEnv(config.DefaultConfigPath())
	}
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Initialize structured logging
	logger.Setup(logger.Config{
		Environment: cfg.Server.Environment,
		Level:       "info", // Could actully pull from config if available, default to info
	})
	logger.Info("Configuration loaded", "environment", cfg.Server.Environment)

	// Set Gin mode based on environment
	if cfg.Server.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database
	db, err := initDatabase(cfg)
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Run migrations if requested
	shouldMigrate := *runMigrate || *autoMigrate || strings.ToLower(os.Getenv("AUTO_MIGRATE")) == "true"
	if shouldMigrate {
		migrationsPath := getMigrationsPath(cfg)
		if migrationsPath != "" {
			slog.Info("Running database migrations", "path", migrationsPath)
			if err := database.RunMigrations(db, migrationsPath); err != nil {
				slog.Error("Migration failed", "error", err)
				os.Exit(1)
			}
			slog.Info("Database migrations completed successfully")
		} else {
			slog.Warn("Migrations requested but no migrations path found for backend", "backend", cfg.Database.Backend)
		}

		// Exit if only running migrations
		if *runMigrate {
			slog.Info("Migrations completed, exiting")
			return
		}
	}

	// Register Prometheus metrics
	metrics.Register()

	// Initialize repositories, services, and handlers
	repos := initRepositories(db, cfg)
	svcs, auditor := initServices(repos, cfg)
	handlers := initHandlers(svcs, repos, cfg, auditor)

	// Initialize HTTP router with all handlers
	router := setupRouter(cfg, handlers, svcs, repos)

	// Create HTTP server with secure defaults
	srv := &http.Server{
		Addr:              cfg.Server.Address(),
		Handler:           router,
		ReadTimeout:       cfg.Server.ReadTimeout,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
		MaxHeaderBytes:    1 << 20, // 1 MB
	}

	// Start server in goroutine
	go func() {
		logger.Info("GoConnect Server starting",
			"version", version,
			"address", cfg.Server.Address(),
			"environment", cfg.Server.Environment,
			"database", cfg.Database.Backend,
		)

		auditType := "stdout"
		if cfg.Audit.SQLiteDSN != "" {
			auditType = "sqlite"
		}
		logger.Info("Audit initialized", "type", auditType)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	slog.Info("Server exited")
}

// buildBaseURL constructs the base URL from server configuration
func buildBaseURL(cfg *config.Config) string {
	protocol := "http"
	if cfg.Server.IsProduction() {
		protocol = "https"
	}
	host := cfg.Server.Host
	// Use localhost for 0.0.0.0 bind address
	if host == "0.0.0.0" {
		host = "localhost"
	}
	return fmt.Sprintf("%s://%s:%s", protocol, host, cfg.Server.Port)
}

// initDatabase initializes the database connection based on configuration
func initDatabase(cfg *config.Config) (*sql.DB, error) {
	switch cfg.Database.Backend {
	case "sqlite", "memory":
		path := cfg.Database.SQLitePath
		if cfg.Database.Backend == "memory" {
			path = ":memory:"
		}
		return database.ConnectSQLite(path)
	default: // postgres
		dbCfg := &database.Config{
			Host:     cfg.Database.Host,
			Port:     cfg.Database.Port,
			User:     cfg.Database.User,
			Password: cfg.Database.Password,
			DBName:   cfg.Database.DBName,
			SSLMode:  cfg.Database.SSLMode,
		}
		return database.Connect(dbCfg)
	}
}

// Repositories holds all repository instances
type Repositories struct {
	User            repository.UserRepository
	Tenant          repository.TenantRepository
	TenantMember    repository.TenantMemberRepository
	TenantInvite    repository.TenantInviteRepository
	TenantChat      repository.TenantChatRepository
	TenantAnnounce  repository.TenantAnnouncementRepository
	Network         repository.NetworkRepository
	Membership      repository.MembershipRepository
	JoinRequest     repository.JoinRequestRepository
	Device          repository.DeviceRepository
	Peer            repository.PeerRepository
	Chat            repository.ChatRepository
	IPAM            repository.IPAMRepository
	Idempotency     repository.IdempotencyRepository
	InviteToken     repository.InviteTokenRepository
	IPRule          repository.IPRuleRepository
	Post            repository.PostRepository
	DeletionRequest repository.DeletionRequestRepository
	Admin           *repository.AdminRepository
}

// Services holds all service instances
type Services struct {
	Auth             *service.AuthService
	TenantMembership *service.TenantMembershipService
	Network          *service.NetworkService
	Membership       *service.MembershipService
	Device           *service.DeviceService
	Peer             *service.PeerService
	Chat             *service.ChatService
	IPAM             *service.IPAMService
	Invite           *service.InviteService
	Admin            *service.AdminService
	GDPR             *service.GDPRService
	Post             *service.PostService
	IPRule           *service.IPRuleService
	PeerProvisioning *service.PeerProvisioningService
	Redis            *redis.Client
}

// Handlers holds all handler instances
type Handlers struct {
	Auth      *handler.AuthHandler
	Tenant    *handler.TenantHandler
	Network   *handler.NetworkHandler
	Device    *handler.DeviceHandler
	Peer      *handler.PeerHandler
	Chat      *handler.ChatHandler
	Invite    *handler.InviteHandler
	WireGuard *handler.WireGuardHandler
	WebSocket *handler.WebSocketHandler
	Admin     *handler.AdminHandler
	GDPR      *handler.GDPRHandler
	Post      *handler.PostHandler
	IPRule    *handler.IPRuleHandler
	Upload    *handler.UploadHandler
	Voice     *handler.VoiceHandler
}

// initRepositories initializes all repositories based on database backend
func initRepositories(db *sql.DB, cfg *config.Config) *Repositories {
	backend := cfg.Database.Backend
	if backend == "" {
		backend = "postgres"
	}
	backend = strings.ToLower(backend)

	if backend == "sqlite" || backend == "memory" {
		return &Repositories{
			User:            repository.NewSQLiteUserRepository(db),
			Tenant:          repository.NewSQLiteTenantRepository(db),
			TenantMember:    repository.NewSQLiteTenantMemberRepository(db),
			TenantInvite:    repository.NewSQLiteTenantInviteRepository(db),
			TenantChat:      repository.NewSQLiteTenantChatRepository(db),
			TenantAnnounce:  repository.NewSQLiteTenantAnnouncementRepository(db),
			Network:         repository.NewSQLiteNetworkRepository(db),
			Membership:      repository.NewSQLiteMembershipRepository(db),
			JoinRequest:     repository.NewSQLiteJoinRequestRepository(db),
			Device:          repository.NewSQLiteDeviceRepository(db),
			Peer:            repository.NewSQLitePeerRepository(db),
			Chat:            repository.NewSQLiteChatRepository(db),
			IPAM:            repository.NewSQLiteIPAMRepository(db),
			Idempotency:     repository.NewInMemoryIdempotencyRepository(),
			InviteToken:     repository.NewSQLiteInviteTokenRepository(db),
			IPRule:          repository.NewSQLiteIPRuleRepository(db),
			Post:            repository.NewPostRepository(db),
			DeletionRequest: repository.NewSQLiteDeletionRequestRepository(db),
			Admin:           repository.NewAdminRepository(db),
		}
	}

	// PostgreSQL
	return &Repositories{
		User:            repository.NewPostgresUserRepository(db),
		Tenant:          repository.NewPostgresTenantRepository(db),
		TenantMember:    repository.NewPostgresTenantMemberRepository(db),
		TenantInvite:    repository.NewPostgresTenantInviteRepository(db),
		TenantChat:      repository.NewPostgresTenantChatRepository(db),
		TenantAnnounce:  repository.NewPostgresTenantAnnouncementRepository(db),
		Network:         repository.NewPostgresNetworkRepository(db),
		Membership:      repository.NewPostgresMembershipRepository(db),
		JoinRequest:     repository.NewPostgresJoinRequestRepository(db),
		Device:          repository.NewPostgresDeviceRepository(db),
		Peer:            repository.NewPostgresPeerRepository(db),
		Chat:            repository.NewPostgresChatRepository(db),
		IPAM:            repository.NewPostgresIPAMRepository(db),
		Idempotency:     repository.NewPostgresIdempotencyRepository(db),
		InviteToken:     repository.NewPostgresInviteTokenRepository(db),
		IPRule:          repository.NewPostgresIPRuleRepository(db),
		Post:            repository.NewPostRepository(db),
		DeletionRequest: repository.NewPostgresDeletionRequestRepository(db),
		Admin:           repository.NewAdminRepository(db),
	}
}

// initServices initializes all services with repositories
// Returns services and auditor
func initServices(repos *Repositories, cfg *config.Config) (*Services, audit.Auditor) {
	// Initialize Redis client (optional)
	var redisClient *redis.Client
	if cfg.Redis.Host != "" {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})
	}

	// Initialize auditor
	var auditor audit.Auditor
	if cfg.Audit.SQLiteDSN != "" {
		var err error
		auditor, err = audit.NewSqliteAuditor(cfg.Audit.SQLiteDSN)
		if err != nil {
			slog.Warn("Failed to initialize SQLite auditor, falling back to stdout", "error", err)
			auditor = audit.NewStdoutAuditor()
		}
	} else {
		auditor = audit.NewStdoutAuditor()
	}

	// Initialize services
	authService := service.NewAuthServiceWithSecret(repos.User, repos.Tenant, redisClient, cfg.JWT.Secret)
	tenantMembershipService := service.NewTenantMembershipService(repos.TenantMember, repos.TenantInvite, repos.TenantAnnounce, repos.TenantChat, repos.Tenant, repos.User)
	networkService := service.NewNetworkService(repos.Network, repos.Idempotency)
	membershipService := service.NewMembershipService(repos.Network, repos.Membership, repos.JoinRequest, repos.Idempotency)
	membershipService.SetInviteTokenRepository(repos.InviteToken)
	deviceService := service.NewDeviceService(repos.Device, repos.User, repos.Peer, repos.Network, cfg.WireGuard)
	peerService := service.NewPeerService(repos.Peer, repos.Device, repos.Network)
	chatService := service.NewChatService(repos.Chat, repos.User)
	chatService.SetAuditor(auditor)
	ipamService := service.NewIPAMService(repos.Network, repos.Membership, repos.IPAM)
	// Build base URL from config
	baseURL := buildBaseURL(cfg)
	inviteService := service.NewInviteService(repos.InviteToken, repos.Network, repos.Membership, baseURL)
	inviteService.SetAuditor(auditor)
	adminService := service.NewAdminService(repos.User, repos.Admin, repos.Tenant, repos.Network, repos.Device, repos.Chat, auditor, redisClient, func() int { return 0 })
	gdprService := service.NewGDPRService(repos.User, repos.Device, repos.Network, repos.Membership, repos.DeletionRequest)
	postService := service.NewPostService(repos.Post, repos.User)
	ipRuleService := service.NewIPRuleService(repos.IPRule)
	peerProvisioningService := service.NewPeerProvisioningService(repos.Peer, repos.Device, repos.Network, repos.Membership, repos.IPAM)

	// Set auditors
	ipamService.SetAuditor(auditor)
	deviceService.SetAuditor(auditor)
	deviceService.SetPeerProvisioning(peerProvisioningService)

	return &Services{
		Auth:             authService,
		TenantMembership: tenantMembershipService,
		Network:          networkService,
		Membership:       membershipService,
		Device:           deviceService,
		Peer:             peerService,
		Chat:             chatService,
		IPAM:             ipamService,
		Invite:           inviteService,
		Admin:            adminService,
		GDPR:             gdprService,
		Post:             postService,
		IPRule:           ipRuleService,
		PeerProvisioning: peerProvisioningService,
		Redis:            redisClient,
	}, auditor
}

// initHandlers initializes all handlers with services
func initHandlers(svcs *Services, repos *Repositories, cfg *config.Config, auditor audit.Auditor) *Handlers {
	// Initialize WireGuard profile generator
	profileGenerator := wireguard.NewProfileGenerator(
		cfg.WireGuard.ServerEndpoint,
		cfg.WireGuard.ServerPubKey,
		cfg.WireGuard.DNS,
		cfg.WireGuard.MTU,
		cfg.WireGuard.Keepalive,
	)

	// Initialize WebSocket hub with default message handler
	wsMessageHandler := websocket.NewDefaultMessageHandler(nil, svcs.Chat, svcs.Membership, svcs.Device, svcs.Auth)
	wsHub := websocket.NewHub(wsMessageHandler)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go wsHub.Run(ctx)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(svcs.Auth, nil) // OIDC can be added later
	tenantHandler := handler.NewTenantHandler(svcs.TenantMembership)
	networkHandler := handler.NewNetworkHandler(svcs.Network, svcs.Membership, svcs.Device, repos.Peer, cfg.WireGuard).WithIPAM(svcs.IPAM)
	deviceHandler := handler.NewDeviceHandler(svcs.Device)
	peerHandler := handler.NewPeerHandler(svcs.Peer)
	chatHandler := handler.NewChatHandler(svcs.Chat)
	inviteHandler := handler.NewInviteHandler(svcs.Invite)
	wireGuardHandler := handler.NewWireGuardHandler(repos.Network, repos.Membership, svcs.Device, svcs.Peer, repos.User, profileGenerator, auditor)
	webSocketHandler := handler.NewWebSocketHandler(wsHub)
	adminHandler := handler.NewAdminHandler(svcs.Admin)
	gdprHandler := handler.NewGDPRHandler(svcs.GDPR, auditor)
	postHandler := handler.NewPostHandler(svcs.Post)
	ipRuleHandler := handler.NewIPRuleHandler(svcs.IPRule)
	// Build base URL from config
	baseURL := buildBaseURL(cfg)
	uploadHandler := handler.NewUploadHandler("/tmp/uploads", baseURL+"/uploads")
	voiceHandler := handler.NewVoiceHandler(svcs.Redis)

	return &Handlers{
		Auth:      authHandler,
		Tenant:    tenantHandler,
		Network:   networkHandler,
		Device:    deviceHandler,
		Peer:      peerHandler,
		Chat:      chatHandler,
		Invite:    inviteHandler,
		WireGuard: wireGuardHandler,
		WebSocket: webSocketHandler,
		Admin:     adminHandler,
		GDPR:      gdprHandler,
		Post:      postHandler,
		IPRule:    ipRuleHandler,
		Upload:    uploadHandler,
		Voice:     voiceHandler,
	}
}

// setupRouter configures the Gin router with all routes
func setupRouter(cfg *config.Config, handlers *Handlers, svcs *Services, repos *Repositories) *gin.Engine {
	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(handler.RequestIDMiddleware())
	router.Use(handler.NewCORSMiddleware(&cfg.CORS)) // Use config-based CORS
	router.Use(metrics.GinMiddleware())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 1*time.Second)
		defer cancel()

		dbStatus := "up"
		var pingErr error
		if repos.Admin == nil {
			pingErr = fmt.Errorf("admin repository not initialized")
		} else {
			pingErr = repos.Admin.PingContext(ctx)
		}
		if pingErr != nil {
			slog.Error("Health check failed: database unreachable", "error", pingErr)
			dbStatus = "down"
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":      "unhealthy",
				"database":    dbStatus,
				"version":     version,
				"environment": cfg.Server.Environment,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":      "healthy",
			"database":    dbStatus,
			"version":     version,
			"environment": cfg.Server.Environment,
		})
	})

	// Metrics endpoint (Prometheus)
	router.GET("/metrics", metrics.Handler())

	// API v1 routes (canonical prefix)
	v1 := router.Group("/v1")

	// Backward compatibility: redirect /api/v1/* to /v1/*
	router.Any("/api/v1/*path", func(c *gin.Context) {
		path := c.Param("path")
		newURL := "/v1" + path
		if c.Request.URL.RawQuery != "" {
			newURL += "?" + c.Request.URL.RawQuery
		}
		c.Redirect(http.StatusPermanentRedirect, newURL)
	})

	// Auth middleware (used for protected routes)
	authMiddleware := handler.AuthMiddleware(svcs.Auth)

	// Register all handler routes
	handler.RegisterAuthRoutes(router, handlers.Auth, authMiddleware)
	handlers.Tenant.RegisterRoutes(v1, authMiddleware)
	handler.RegisterNetworkRoutes(router, handlers.Network, svcs.Auth, repos.Membership)
	handler.RegisterDeviceRoutes(router, handlers.Device, authMiddleware)
	handler.RegisterPeerRoutes(router, handlers.Peer, authMiddleware)
	handler.RegisterChatRoutes(router, handlers.Chat, authMiddleware)
	handler.RegisterInviteRoutes(router, handlers.Invite, authMiddleware)
	handler.RegisterWireGuardRoutes(router, handlers.WireGuard, authMiddleware)
	handler.RegisterWebSocketRoutes(router, handlers.WebSocket, authMiddleware)
	handler.RegisterUploadRoutes(router, handlers.Upload, authMiddleware)

	// Admin routes (require admin role)
	adminGroup := v1.Group("/admin")
	adminGroup.Use(authMiddleware)
	adminGroup.Use(handler.RequireAdmin())
	{
		adminGroup.GET("/stats", handlers.Admin.GetSystemStats)
		adminGroup.GET("/users", handlers.Admin.ListUsers)
		adminGroup.GET("/tenants", handlers.Admin.ListTenants)
		adminGroup.GET("/networks", handlers.Admin.ListNetworks)
		adminGroup.GET("/devices", handlers.Admin.ListDevices)
	}

	// GDPR routes
	gdprGroup := v1.Group("/gdpr")
	gdprGroup.Use(authMiddleware)
	{
		gdprGroup.POST("/request-deletion", handlers.GDPR.RequestDeletion)
		gdprGroup.POST("/export-data", handlers.GDPR.ExportData)
		gdprGroup.GET("/export-data/download", handlers.GDPR.ExportDataDownload)
	}

	// Post routes
	postGroup := v1.Group("/posts")
	postGroup.Use(authMiddleware)
	{
		postGroup.POST("", handlers.Post.CreatePost)
		postGroup.GET("", handlers.Post.GetPosts)
		postGroup.GET("/:id", handlers.Post.GetPost)
		postGroup.PATCH("/:id", handlers.Post.UpdatePost)
		postGroup.DELETE("/:id", handlers.Post.DeletePost)
	}

	// Voice routes
	voiceGroup := v1.Group("/voice")
	voiceGroup.Use(authMiddleware)
	{
		voiceGroup.POST("/signal", handlers.Voice.Signal)
		voiceGroup.GET("/signals", handlers.Voice.GetSignals)
	}

	// IP Rule routes
	ipRuleGroup := v1.Group("/ip-rules")
	ipRuleGroup.Use(authMiddleware)
	ipRuleGroup.Use(handler.RequireAdmin())
	{
		ipRuleGroup.POST("", handlers.IPRule.CreateIPRule)
		ipRuleGroup.GET("", handlers.IPRule.ListIPRules)
		ipRuleGroup.POST("/check", handlers.IPRule.CheckIP)
		ipRuleGroup.GET("/:id", handlers.IPRule.GetIPRule)
		ipRuleGroup.DELETE("/:id", handlers.IPRule.DeleteIPRule)
	}

	// Health/Info
	v1.GET("/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"version":     version,
			"commit":      commit,
			"environment": cfg.Server.Environment,
		})
	})

	return router
}

// getMigrationsPath returns the path to database migrations based on the database backend.
// Returns empty string if migrations are not supported for the backend.
func getMigrationsPath(cfg *config.Config) string {
	switch cfg.Database.Backend {
	case "postgres":
		// Check for migrations in standard locations
		paths := []string{"migrations", "./migrations", "../migrations", "core/migrations"}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
		return "migrations" // Default fallback
	case "sqlite":
		// SQLite has separate simplified migrations
		paths := []string{"migrations_sqlite", "./migrations_sqlite", "../migrations_sqlite", "core/migrations_sqlite"}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
		return "migrations_sqlite" // Default fallback
	case "memory":
		// In-memory database doesn't need migrations (schema created on connect)
		return ""
	default:
		return ""
	}
}
