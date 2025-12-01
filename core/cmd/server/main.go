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
	"gopkg.in/yaml.v3"
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
	usePostgres := flag.Bool("postgres", true, "use PostgreSQL instead of in-memory storage (deprecated, use --db-backend)")
	dbBackendFlag := flag.String("db-backend", "", "database backend: postgres|sqlite|memory")
	sqlitePathFlag := flag.String("db-sqlite-path", "", "path to SQLite database file (when --db-backend=sqlite)")
	setupModeFlag := flag.Bool("setup-mode", false, "start server in setup/wizard mode (experimental)")
	asyncAudit := flag.Bool("audit-async", true, "enable async audit buffering")
	auditQueue := flag.Int("audit-queue", 1024, "audit async queue size")
	auditWorkers := flag.Int("audit-workers", 1, "audit async worker count")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), shutdownSignals()...)
	defer stop()

	configPath := "/data/goconnect.yaml"

	// Register metrics early
	metrics.Register()

	if *showVersion {
		fmt.Printf("goconnect-server %s (commit %s, build %s) built by %s\n", version, commit, date, builtBy)
		return
	}

	// Load configuration
	cfg, err := config.LoadFromFileOrEnv(configPath)
	setupMode := *setupModeFlag
	if err != nil {
		log.Printf("Warning: Failed to load full config: %v. Entering setup mode with defaults.", err)
		setupMode = true
		// Minimal defaults to keep HTTP server up; real values will be collected in setup wizard.
		cfg = &config.Config{
			Server: config.ServerConfig{Host: "0.0.0.0", Port: "8080"},
			Database: config.DatabaseConfig{
				Backend: "memory",
			},
		}
	}

	// Resolve backend selection priority: flag > config > env default
	dbBackend := "postgres"
	if cfg != nil && cfg.Database.Backend != "" {
		dbBackend = cfg.Database.Backend
	}
	if *dbBackendFlag != "" {
		dbBackend = *dbBackendFlag
	}
	// legacy compatibility: --postgres=false implies memory unless db-backend explicitly set
	if !*usePostgres && *dbBackendFlag == "" {
		dbBackend = "memory"
	}
	dbBackend = strings.ToLower(dbBackend)

	// Resolve SQLite path (only used when backend=sqlite)
	sqlitePath := "data/goconnect.db"
	if cfg != nil && cfg.Database.SQLitePath != "" {
		sqlitePath = cfg.Database.SQLitePath
	}
	if *sqlitePathFlag != "" {
		sqlitePath = *sqlitePathFlag
	}

	// Setup mode toggle (env or flag)
	if !setupMode {
		if smEnv := strings.TrimSpace(os.Getenv("SETUP_MODE")); smEnv != "" {
			if parsed, err := strconv.ParseBool(smEnv); err == nil {
				setupMode = parsed
			}
		}
	}
	// If config fell back to memory backend and no explicit override, prefer setup mode to avoid crashy startup.
	if cfg != nil && cfg.Database.Backend == "memory" && *dbBackendFlag == "" && !*usePostgres {
		setupMode = true
	}

	if setupMode {
		r := gin.New()
		r.Use(gin.Recovery())
		r.Use(metrics.GinMiddleware())
		registerSetupRoutes(r, dbBackend, sqlitePath, configPath, stop)
		startHTTPServer(ctx, cfg, r)
		return
	}

	// Database setup
	var db *sql.DB
	migrationsPath := getEnvOrDefault("MIGRATIONS_PATH", "./migrations")
	if dbBackend == "sqlite" {
		migrationsPath = getEnvOrDefault("MIGRATIONS_SQLITE_PATH", "./migrations_sqlite")
	}
	switch dbBackend {
	case "postgres":
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
			if err := database.RunMigrations(db, migrationsPath); err != nil {
				log.Fatalf("Failed to run migrations: %v", err)
			}
			fmt.Println("Migrations completed successfully")
			return
		}
	case "sqlite":
		var err error
		db, err = database.ConnectSQLite(sqlitePath)
		if err != nil {
			log.Fatalf("Failed to connect to SQLite: %v", err)
		}
		defer db.Close()
		fmt.Printf("Connected to SQLite (experimental) at %s\n", sqlitePath)
		if *runMigrations {
			if err := database.RunSQLiteMigrations(db, migrationsPath); err != nil {
				log.Fatalf("SQLite migrations failed: %v", err)
			}
			fmt.Println("SQLite migrations completed successfully")
			return
		}
	case "memory":
		// No DB connection required
	default:
		log.Fatalf("Unsupported DB backend: %s (use postgres|sqlite|memory)", dbBackend)
	}

	// Initialize repositories
	var networkRepo repository.NetworkRepository
	var idempotencyRepo repository.IdempotencyRepository
	var membershipRepo repository.MembershipRepository
	var joinRepo repository.JoinRequestRepository
	var ipamRepo repository.IPAMRepository
	var ipRuleRepo repository.IPRuleRepository
	var userRepo repository.UserRepository
	var tenantRepo repository.TenantRepository
	var deviceRepo repository.DeviceRepository
	// Tenant multi-membership repositories
	var tenantMemberRepo repository.TenantMemberRepository
	var tenantInviteRepo repository.TenantInviteRepository
	var tenantAnnouncementRepo repository.TenantAnnouncementRepository
	var tenantChatRepo repository.TenantChatRepository
	var peerRepo repository.PeerRepository
	var chatRepo repository.ChatRepository
	var inviteRepo repository.InviteTokenRepository
	var adminRepo *repository.AdminRepository
	var deletionRepo repository.DeletionRequestRepository

	if dbBackend == "postgres" && db != nil {
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
		inviteRepo = repository.NewPostgresInviteTokenRepository(db)
		adminRepo = repository.NewAdminRepository(db)
		// Tenant multi-membership repositories (PostgreSQL)
		tenantMemberRepo = repository.NewPostgresTenantMemberRepository(db)
		tenantInviteRepo = repository.NewPostgresTenantInviteRepository(db)
		tenantAnnouncementRepo = repository.NewPostgresTenantAnnouncementRepository(db)
		tenantChatRepo = repository.NewPostgresTenantChatRepository(db)
		// TODO: Implement PostgresDeletionRequestRepository
		deletionRepo = repository.NewInMemoryDeletionRequestRepository()
		fmt.Println("Using PostgreSQL repositories")
	} else if dbBackend == "sqlite" {
		userRepo = repository.NewSQLiteUserRepository(db)
		tenantRepo = repository.NewSQLiteTenantRepository(db)
		networkRepo = repository.NewSQLiteNetworkRepository(db)
		membershipRepo = repository.NewSQLiteMembershipRepository(db)
		joinRepo = repository.NewSQLiteJoinRequestRepository(db)
		tenantMemberRepo = repository.NewSQLiteTenantMemberRepository(db)
		tenantInviteRepo = repository.NewSQLiteTenantInviteRepository(db)
		tenantAnnouncementRepo = repository.NewSQLiteTenantAnnouncementRepository(db)
		tenantChatRepo = repository.NewSQLiteTenantChatRepository(db)
		peerRepo = repository.NewSQLitePeerRepository(db)
		inviteRepo = repository.NewSQLiteInviteTokenRepository(db)
		ipRuleRepo = repository.NewSQLiteIPRuleRepository(db)
		chatRepo = repository.NewSQLiteChatRepository(db)
		// Other repos still in-memory until SQLite variants are added.
		log.Println("SQLite backend: sqlite repos for users/tenants/networks/memberships/join_requests/tenant_members/tenant_invites/tenant_announcements/tenant_chat/peers/invite_tokens/ip_rules/chat/ipam/idempotency.")
		idempotencyRepo = repository.NewInMemoryIdempotencyRepository()
		ipamRepo = repository.NewSQLiteIPAMRepository(db)
		deviceRepo = repository.NewSQLiteDeviceRepository(db)
		deletionRepo = repository.NewSQLiteDeletionRequestRepository(db)
	} else {
		// In-memory repositories (fallback)
		networkRepo = repository.NewInMemoryNetworkRepository()
		idempotencyRepo = repository.NewInMemoryIdempotencyRepository()
		membershipRepo = repository.NewInMemoryMembershipRepository()
		joinRepo = repository.NewInMemoryJoinRequestRepository()
		ipamRepo = repository.NewInMemoryIPAM()
		ipRuleRepo = repository.NewInMemoryIPRuleRepository()
		userRepo = repository.NewInMemoryUserRepository()
		tenantRepo = repository.NewInMemoryTenantRepository()
		deviceRepo = repository.NewInMemoryDeviceRepository()
		peerRepo = repository.NewInMemoryPeerRepository()
		chatRepo = repository.NewInMemoryChatRepository()
		inviteRepo = repository.NewInMemoryInviteTokenRepository()
		// Tenant multi-membership repositories (In-Memory)
		tenantMemberRepo = repository.NewInMemoryTenantMemberRepository()
		tenantInviteRepo = repository.NewInMemoryTenantInviteRepository()
		tenantAnnouncementRepo = repository.NewInMemoryTenantAnnouncementRepository()
		tenantChatRepo = repository.NewInMemoryTenantChatRepository()
		deletionRepo = repository.NewInMemoryDeletionRequestRepository()
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

	// Initialize Tenant Membership Service (multi-tenant system)
	tenantMembershipService := service.NewTenantMembershipService(
		tenantMemberRepo,
		tenantInviteRepo,
		tenantAnnouncementRepo,
		tenantChatRepo,
		tenantRepo,
		userRepo,
	)

	peerProvisioningService := service.NewPeerProvisioningService(peerRepo, deviceRepo, networkRepo, membershipRepo, ipamRepo)
	peerService := service.NewPeerService(peerRepo, deviceRepo, networkRepo)
	deviceService := service.NewDeviceService(deviceRepo, userRepo, peerRepo, networkRepo, cfg.WireGuard)
	deviceService.SetPeerProvisioning(peerProvisioningService)
	// Start offline detection (check every 30s, mark offline if unseen for 2m)
	go deviceService.StartOfflineDetection(ctx, 30*time.Second, 2*time.Minute)

	chatService := service.NewChatService(chatRepo, userRepo)

	// Initialize GDPR service
	gdprService := service.NewGDPRService(userRepo, deviceRepo, networkRepo, membershipRepo, deletionRepo)
	// Start GDPR deletion worker
	gdprService.StartWorker(ctx, 1*time.Minute)

	// Initialize Invite service
	baseURL := getEnvOrDefault("APP_BASE_URL", "https://app.goconnect.example")
	inviteService := service.NewInviteService(inviteRepo, networkRepo, membershipRepo, baseURL)

	// Initialize IP Rule service
	ipRuleRepo = repository.NewInMemoryIPRuleRepository()
	ipRuleService := service.NewIPRuleService(ipRuleRepo)

	// Initialize Post service
	postRepo := repository.NewPostRepository(db)
	postService := service.NewPostService(postRepo, userRepo)

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
	peerHandler := handler.NewPeerHandler(peerService)
	authHandler := handler.NewAuthHandler(authService, oidcService)
	deviceHandler := handler.NewDeviceHandler(deviceService)
	chatHandler := handler.NewChatHandler(chatService)
	uploadHandler := handler.NewUploadHandler("./uploads", "/uploads")
	gdprHandler := handler.NewGDPRHandler(gdprService, aud)
	inviteHandler := handler.NewInviteHandler(inviteService)
	ipRuleHandler := handler.NewIPRuleHandler(ipRuleService)
	tenantHandler := handler.NewTenantHandler(tenantMembershipService)
	postHandler := handler.NewPostHandler(postService)

	// Initialize WebSocket components
	// Circular dependency resolution: Handler -> Hub -> Handler
	wsMsgHandler := ws.NewDefaultMessageHandler(nil, chatService, membershipService, deviceService, authService)
	hub := ws.NewHub(wsMsgHandler)
	wsMsgHandler.SetHub(hub)

	// Start Hub
	go hub.Run(ctx)

	// Initialize Admin Service
	adminService := service.NewAdminService(userRepo, adminRepo, tenantRepo, networkRepo, deviceRepo, chatRepo, aud, redisClient, hub.GetActiveConnectionCount)
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

	if setupMode {
		r.GET("/setup", func(c *gin.Context) {
			c.JSON(http.StatusNotImplemented, gin.H{
				"status":        "setup-mode",
				"message":       "Setup wizard placeholder; full flow coming soon.",
				"db_backend":    dbBackend,
				"sqlite_path":   sqlitePath,
				"migrations_ok": dbBackend == "postgres",
			})
		})
		r.NoRoute(func(c *gin.Context) {
			c.Redirect(http.StatusFound, "/setup")
		})
	}

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

	// User profile routes
	usersGroup := r.Group("/v1/users")
	usersGroup.Use(handler.AuthMiddleware(authService))
	{
		usersGroup.GET("/me", authHandler.Me)
		usersGroup.PUT("/me", authHandler.UpdateProfile)
	}

	// Register network routes (auth + role middleware applied within)
	handler.RegisterNetworkRoutes(r, networkHandler, authService, membershipRepo)

	// Register tenant routes (public discovery + authenticated operations)
	apiV1 := r.Group("/v1")
	tenantHandler.RegisterRoutes(apiV1, handler.AuthMiddleware(authService))

	// Register device routes
	handler.RegisterDeviceRoutes(r, deviceHandler, handler.AuthMiddleware(authService))

	// Register peer routes
	handler.RegisterPeerRoutes(r, peerHandler, handler.AuthMiddleware(authService))

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

	// Register post routes
	postGroup := r.Group("/v1/posts")
	postGroup.Use(handler.AuthMiddleware(authService))
	{
		postGroup.POST("", postHandler.CreatePost)
		postGroup.GET("", postHandler.GetPosts)
		postGroup.GET("/:id", postHandler.GetPost)
		postGroup.PUT("/:id", postHandler.UpdatePost)
		postGroup.DELETE("/:id", postHandler.DeletePost)
		postGroup.POST("/:id/like", postHandler.LikePost)
		postGroup.DELETE("/:id/like", postHandler.UnlikePost)
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
		// Legacy routes (keep for backward compatibility)
		adminGroup.GET("/users", adminHandler.ListUsers)
		adminGroup.POST("/users/:id/toggle-admin", adminHandler.ToggleUserAdmin)
		adminGroup.DELETE("/users/:id", adminHandler.DeleteUser)

		// New admin user management routes
		adminGroup.GET("/users/all", adminHandler.ListAllUsers)             // List all users with filters
		adminGroup.GET("/users/:id/details", adminHandler.GetUserDetails)   // Get user details
		adminGroup.PUT("/users/:id/role", adminHandler.UpdateUserRole)      // Update user role
		adminGroup.POST("/users/:id/suspend", adminHandler.SuspendUser)     // Suspend user
		adminGroup.DELETE("/users/:id/suspend", adminHandler.UnsuspendUser) // Unsuspend user
		adminGroup.GET("/users/stats", adminHandler.GetUserStats)           // Get user stats

		// Other admin routes
		adminGroup.GET("/tenants", adminHandler.ListTenants)
		adminGroup.DELETE("/tenants/:id", adminHandler.DeleteTenant)
		adminGroup.GET("/networks", adminHandler.ListNetworks)
		adminGroup.GET("/devices", adminHandler.ListDevices)
		adminGroup.GET("/stats", adminHandler.GetSystemStats)

		// IP Rule management endpoints
		adminGroup.POST("/ip-rules", handler.GinWrap(ipRuleHandler.CreateIPRule))
		adminGroup.GET("/ip-rules", handler.GinWrap(ipRuleHandler.ListIPRules))
		adminGroup.GET("/ip-rules/:id", handler.GinWrap(ipRuleHandler.GetIPRule))
		adminGroup.DELETE("/ip-rules/:id", handler.GinWrap(ipRuleHandler.DeleteIPRule))
		adminGroup.POST("/ip-rules/check", handler.GinWrap(ipRuleHandler.CheckIP))
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

type setupRequest struct {
	Config  config.Config `json:"config" binding:"required"`
	Restart bool          `json:"restart,omitempty"`
}

type setupStep struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Fields      []string `json:"fields"`
}

type setupStatus struct {
	Status          string      `json:"status"`
	Message         string      `json:"message,omitempty"`
	ConfigPath      string      `json:"config_path"`
	ConfigPresent   bool        `json:"config_present"`
	ConfigValid     bool        `json:"config_valid"`
	ValidationError string      `json:"validation_error,omitempty"`
	Mode            string      `json:"mode,omitempty"`
	CompletedSteps  []string    `json:"completed_steps,omitempty"`
	NextStep        string      `json:"next_step"`
	Steps           []setupStep `json:"steps,omitempty"`
}

const setupCompletedStep = "completed"

var setupSteps = []setupStep{
	{
		ID:          "mode",
		Title:       "Mode Selection",
		Description: "Choose Personal (SQLite) or Enterprise (Postgres).",
		Fields:      []string{"database.backend", "database.sqlite_path", "database.host", "database.port", "database.user", "database.password", "database.dbname"},
	},
	{
		ID:          "admin",
		Title:       "Admin Creation",
		Description: "Set JWT secret and WireGuard parameters.",
		Fields:      []string{"jwt.secret", "wireguard.server_endpoint", "wireguard.server_pubkey"},
	},
	{
		ID:          "finalize",
		Title:       "Finalize",
		Description: "Write config and restart server.",
		Fields:      []string{},
	},
}

func registerSetupRoutes(r *gin.Engine, dbBackend string, sqlitePath string, configPath string, stop context.CancelFunc) {
	r.GET("/setup", func(c *gin.Context) {
		// Check if browser wants HTML (Accept header)
		acceptHeader := c.GetHeader("Accept")
		if strings.Contains(acceptHeader, "text/html") {
			// Return HTML setup wizard
			html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GoConnect Setup Wizard</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 0; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); min-height: 100vh; }
        .container { max-width: 800px; margin: 0 auto; padding: 40px 20px; }
        .card { background: white; border-radius: 12px; padding: 40px; box-shadow: 0 20px 40px rgba(0,0,0,0.1); }
        .logo { font-size: 32px; font-weight: bold; color: #667eea; margin-bottom: 10px; }
        .subtitle { color: #666; margin-bottom: 40px; }
        .step { display: none; }
        .step.active { display: block; }
        .step-number { background: #667eea; color: white; width: 30px; height: 30px; border-radius: 50%; display: inline-flex; align-items: center; justify-content: center; margin-right: 10px; font-weight: bold; }
        .step-title { font-size: 24px; font-weight: bold; margin-bottom: 10px; }
        .step-desc { color: #666; margin-bottom: 30px; }
        .form-group { margin-bottom: 20px; }
        label { display: block; margin-bottom: 5px; font-weight: 500; }
        input, select, textarea { width: 100%; padding: 12px; border: 1px solid #ddd; border-radius: 6px; font-size: 16px; }
        button { background: #667eea; color: white; border: none; padding: 12px 24px; border-radius: 6px; font-size: 16px; cursor: pointer; margin-right: 10px; }
        button:hover { background: #5a6fd8; }
        button:disabled { background: #ccc; cursor: not-allowed; }
        .progress { background: #f0f0f0; height: 4px; border-radius: 2px; margin-bottom: 30px; }
        .progress-bar { background: #667eea; height: 4px; border-radius: 2px; transition: width 0.3s; }
        .success { color: #28a745; text-align: center; padding: 40px; }
        .error { color: #dc3545; background: #f8d7da; padding: 10px; border-radius: 6px; margin-bottom: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="card">
            <div class="logo">🚀 GoConnect</div>
            <div class="subtitle">Network Management Platform - Setup Wizard</div>
            
            <div class="progress">
                <div class="progress-bar" id="progress" style="width: 33%"></div>
            </div>
            
            <!-- Step 1: Mode Selection -->
            <div class="step active" id="step1">
                <div class="step-title"><span class="step-number">1</span>Choose Setup Mode</div>
                <div class="step-desc">Select how you want to store your data</div>
                
                <div class="form-group">
                    <label for="mode">Setup Mode</label>
                    <select id="mode" onchange="updateMode()">
                        <option value="personal">🏠 Personal (SQLite) - Simple, single computer</option>
                        <option value="enterprise" selected>🏢 Enterprise (PostgreSQL) - Multi-user, production ready</option>
                    </select>
                </div>
                
                <div id="personal-fields" style="display: none;">
                    <div class="form-group">
                        <label for="sqlite-path">Database File Path</label>
                        <input type="text" id="sqlite-path" value="./goconnect.db" placeholder="./goconnect.db">
                        <small><strong>Nedir?</strong> SQLite database dosyasının yolu. <strong>Örnek:</strong> "./goconnect.db" <strong>Neden gerekli?</strong> Tüm verilerin saklanacağı dosya. Personal mod için basit ve tek dosyalı çözüm.</small>
                    </div>
                </div>
                
                <div id="enterprise-fields">
                    <div class="form-group">
                        <label for="db-host">Database Host</label>
                        <input type="text" id="db-host" value="postgres" placeholder="localhost">
                        <small><strong>Nedir?</strong> PostgreSQL sunucusunun adresi. <strong>Örnekler:</strong> "localhost", "postgres", "db.mycompany.com" <strong>Neden gerekli?</strong> GoConnect'in veritabanına bağlanması için sunucu adresi.</small>
                    </div>
                    <div class="form-group">
                        <label for="db-port">Database Port</label>
                        <input type="text" id="db-port" value="5432" placeholder="5432">
                        <small><strong>Nedir?</strong> PostgreSQL sunucusunun portu. <strong>Örnek:</strong> "5432" (PostgreSQL default) <strong>Neden gerekli?</strong> Doğru porta bağlanmak için.</small>
                    </div>
                    <div class="form-group">
                        <label for="db-name">Database Name</label>
                        <input type="text" id="db-name" value="goconnect" placeholder="goconnect">
                        <small><strong>Nedir?</strong> PostgreSQL database adı. <strong>Örnek:</strong> "goconnect", "my_vpn_db" <strong>Neden gerekli?</strong> GoConnect verilerinin saklanacağı database.</small>
                    </div>
                    <div class="form-group">
                        <label for="db-user">Database User</label>
                        <input type="text" id="db-user" value="goconnect" placeholder="postgres">
                        <small><strong>Nedir?</strong> PostgreSQL kullanıcı adı. <strong>Örnekler:</strong> "goconnect", "postgres", "vpn_user" <strong>Neden gerekli?</strong> Database'e erişim için kullanıcı kimliği.</small>
                    </div>
                    <div class="form-group">
                        <label for="db-password">Database Password</label>
                        <input type="password" id="db-password" value="goconnect_secret" placeholder="your password">
                        <small><strong>Nedir?</strong> PostgreSQL şifresi. <strong>Örnek:</strong> "my_secure_password_123" <strong>Neden gerekli?</strong> Database güvenliği için. Production'da güçlü şifre kullanın!</small>
                    </div>
                </div>
                
                <button onclick="nextStep()">Next →</button>
            </div>
            
            <!-- Step 2: Admin Setup -->
            <div class="step" id="step2">
                <div class="step-title"><span class="step-number">2</span>Admin Account & Security</div>
                <div class="step-desc">Configure your admin account and security settings</div>
                
                <div class="form-group">
                    <label for="jwt-secret">JWT Secret</label>
                    <div style="display: flex; gap: 10px;">
                        <input type="text" id="jwt-secret" value="goconnect-super-secret-jwt-key-change-in-production" style="flex: 1;">
                        <button type="button" onclick="generateJWTSecret()" style="padding: 12px; background: #28a745;">🎲 Random</button>
                    </div>
                    <small><strong>Nedir?</strong> JWT (JSON Web Token) authentication için güvenlik anahtarı. <strong>Örnek:</strong> "my-super-secret-jwt-key-1234567890" <strong>Neden gerekli?</strong> Kullanıcı girişi ve API güvenliği için şifreleme anahtarı.</small>
                </div>
                
                <div class="form-group">
                    <label for="wg-endpoint">WireGuard VPN Endpoint</label>
                    <input type="text" id="wg-endpoint" value="vpn.example.com:51820" placeholder="vpn.example.com:51820">
                    <small><strong>Nedir?</strong> WireGuard VPN sunucusunun public IP adresi ve portu. <strong>Örnekler:</strong> "123.45.67.89:51820" veya "vpn.mycompany.com:51820" <strong>Neden gerekli?</strong> VPN client'larının sunucuya bağlanması için adres. Development için "localhost:51820" kullanabilirsiniz.</small>
                </div>
                
                <div class="form-group">
                    <label for="wg-pubkey">WireGuard Public Key</label>
                    <div style="display: flex; gap: 10px;">
                        <input type="text" id="wg-pubkey" value="auto-generated" readonly style="flex: 1;">
                        <button type="button" onclick="generateWGPublicKey()" style="padding: 12px; background: #28a745;">🎲 Generate</button>
                    </div>
                    <small><strong>Nedir?</strong> WireGuard sunucusunun public anahtarı. <strong>Örnek:</strong> "ABC123xyz789..." <strong>Neden gerekli?</strong> VPN client'larının sunucuyu doğrulaması için. Otomatik olarak oluşturulur.</small>
                </div>
                
                <button onclick="prevStep()">← Previous</button>
                <button onclick="nextStep()">Next →</button>
            </div>
            
            <!-- Step 3: Complete Setup -->
            <div class="step" id="step3">
                <div class="step-title"><span class="step-number">3</span>Complete Setup</div>
                <div class="step-desc">Review your configuration and start GoConnect</div>
                
                <div id="config-review">
                    <h4>Configuration Summary:</h4>
                    <div id="summary-content"></div>
                </div>
                
                <div id="setup-status"></div>
                
                <button onclick="prevStep()">← Previous</button>
                <button onclick="completeSetup()" id="complete-btn">🚀 Complete Setup</button>
            </div>
            
            <!-- Success -->
            <div class="step" id="success">
                <div class="success">
                    <div class="logo">🎉</div>
                    <h2>Setup Complete!</h2>
                    <p>GoConnect is now configured and ready to use.</p>
                    <p><strong>Next Steps:</strong></p>
                    <ol style="text-align: left; max-width: 400px; margin: 0 auto;">
                        <li>Create your admin account</li>
                        <li>Create your first network (tenant)</li>
                        <li>Invite team members</li>
                        <li>Start connecting!</li>
                    </ol>
                    <br>
                    <button onclick="window.location.href='/'">🚀 Go to GoConnect</button>
                </div>
            </div>
        </div>
    </div>
    
    <script>
        let currentStep = 1;
        const totalSteps = 3;
        
        function updateProgress() {
            const progress = (currentStep / totalSteps) * 100;
            document.getElementById('progress').style.width = progress + '%';
        }
        
        function showStep(step) {
            document.querySelectorAll('.step').forEach(el => el.classList.remove('active'));
            document.getElementById('step' + step).classList.add('active');
            currentStep = step;
            updateProgress();
            
            if (step === 3) {
                updateSummary();
            }
        }
        
        function generateJWTSecret() {
            const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*';
            let secret = '';
            for (let i = 0; i < 64; i++) {
                secret += chars.charAt(Math.floor(Math.random() * chars.length));
            }
            document.getElementById('jwt-secret').value = secret;
        }
        
        function generateWGPublicKey() {
            // Generate a valid WireGuard public key (44 base64 characters ending with =)
            const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/';
            let key = '';
            for (let i = 0; i < 43; i++) {
                key += chars.charAt(Math.floor(Math.random() * chars.length));
            }
            key += '=';
            document.getElementById('wg-pubkey').value = key;
        }
        
        function updateMode() {
            const mode = document.getElementById('mode').value;
            document.getElementById('personal-fields').style.display = mode === 'personal' ? 'block' : 'none';
            document.getElementById('enterprise-fields').style.display = mode === 'enterprise' ? 'block' : 'none';
        }
        
        function updateSummary() {
            const mode = document.getElementById('mode').value;
            let summary = '<ul>';
            summary += '<li><strong>Mode:</strong> ' + (mode === 'personal' ? 'Personal (SQLite)' : 'Enterprise (PostgreSQL)') + '</li>';
            
            if (mode === 'personal') {
                summary += '<li><strong>Database:</strong> ' + document.getElementById('sqlite-path').value + '</li>';
            } else {
                summary += '<li><strong>Database:</strong> ' + document.getElementById('db-host').value + ':' + document.getElementById('db-port').value + '/' + document.getElementById('db-name').value + '</li>';
            }
            
            summary += '<li><strong>Security:</strong> JWT configured</li>';
            summary += '<li><strong>VPN Endpoint:</strong> ' + document.getElementById('wg-endpoint').value + '</li>';
            summary += '</ul>';
            
            document.getElementById('summary-content').innerHTML = summary;
        }
        
        function nextStep() {
            if (currentStep < totalSteps) {
                showStep(currentStep + 1);
            }
        }
        
        function prevStep() {
            if (currentStep > 1) {
                showStep(currentStep - 1);
            }
        }
        
        async function completeSetup() {
            const btn = document.getElementById('complete-btn');
            btn.disabled = true;
            btn.textContent = '⏳ Setting up...';
            
            const status = document.getElementById('setup-status');
            
            // Validate JWT secret length
            const jwtSecret = document.getElementById('jwt-secret').value;
            if (jwtSecret.length < 32) {
                status.innerHTML = '<div class="error">❌ JWT Secret must be at least 32 characters. Click "🎲 Random" to generate a secure key.</div>';
                btn.disabled = false;
                btn.textContent = '🚀 Complete Setup';
                return;
            }
            
            // Validate SQLite path if personal mode
            if (mode === 'personal') {
                const sqlitePath = document.getElementById('sqlite-path').value;
                if (!sqlitePath || sqlitePath.trim() === '') {
                    status.innerHTML = '<div class="error">❌ SQLite database path is required for Personal mode. Please specify a file path like "./goconnect.db"</div>';
                    btn.disabled = false;
                    btn.textContent = '🚀 Complete Setup';
                    return;
                }
            }
            
            // Validate WireGuard public key
            const wgPubKey = document.getElementById('wg-pubkey').value;
            if (!wgPubKey || wgPubKey === 'auto-generated' || wgPubKey.trim() === '') {
                status.innerHTML = '<div class="error">❌ WireGuard Public Key is required. Click "🎲 Generate" to create a key.</div>';
                btn.disabled = false;
                btn.textContent = '🚀 Complete Setup';
                return;
            }
            
            try {
                const mode = document.getElementById('mode').value;
                const config = {
                    server: { host: "0.0.0.0", port: "8080" },
                    database: {},
                    jwt: { secret: document.getElementById('jwt-secret').value },
                    wireguard: { 
                        server_endpoint: document.getElementById('wg-endpoint').value,
                        server_pubkey: document.getElementById('wg-pubkey').value
                    }
                };
                
                if (mode === 'personal') {
                    config.database = {
                        backend: "sqlite",
                        sqlite_path: document.getElementById('sqlite-path').value
                    };
                } else {
                    config.database = {
                        backend: "postgres",
                        host: document.getElementById('db-host').value,
                        port: document.getElementById('db-port').value,
                        user: document.getElementById('db-user').value,
                        password: document.getElementById('db-password').value,
                        dbname: document.getElementById('db-name').value
                    };
                }
                
                const response = await fetch('/setup', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ 
                        config, 
                        restart: true
                    })
                });
                
                const result = await response.json();
                
                if (response.ok) {
                    status.innerHTML = '<div class="success">✅ Setup completed! Server is restarting...</div>';
                    setTimeout(() => showStep(4), 3000);
                } else {
                    status.innerHTML = '<div class="error">❌ Setup failed: ' + result.error + '</div>';
                    btn.disabled = false;
                    btn.textContent = '🚀 Complete Setup';
                }
            } catch (error) {
                status.innerHTML = '<div class="error">❌ Setup failed: ' + error.message + '</div>';
                btn.disabled = false;
                btn.textContent = '🚀 Complete Setup';
            }
        }
        
        // Initialize
        updateMode();
    </script>
</body>
</html>`
			c.Header("Content-Type", "text/html")
			c.String(http.StatusOK, html)
			return
		}

		// Original JSON response for API calls
		state := buildSetupStatus(configPath)
		state.Message = "Setup wizard ready. Complete the steps to persist config and restart."
		state.Mode = dbBackend
		state.Steps = setupSteps
		if state.NextStep == "" {
			state.NextStep = "mode"
		}
		if state.Mode == "" {
			state.Mode = dbBackend
		}
		state.ConfigPresent = state.ConfigPresent || fileExists(configPath)
		state.ConfigPath = configPath
		c.JSON(http.StatusOK, state)
	})
	r.GET("/setup/status", func(c *gin.Context) {
		state := buildSetupStatus(configPath)
		state.Steps = setupSteps
		c.JSON(http.StatusOK, state)
	})
	r.POST("/setup/validate", func(c *gin.Context) {
		var req setupRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid setup payload", "details": err.Error()})
			return
		}
		if err := req.Config.Validate(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "details": err.Error()})
			return
		}
		completed, next := evaluateSetupProgress(&req.Config)
		resp := setupStatus{
			Status:         "ok",
			Message:        "config validated",
			ConfigPath:     configPath,
			ConfigPresent:  fileExists(configPath),
			ConfigValid:    true,
			CompletedSteps: completed,
			NextStep:       next,
		}
		c.JSON(http.StatusOK, resp)
	})
	r.POST("/setup", func(c *gin.Context) {
		var req setupRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid setup payload", "details": err.Error()})
			return
		}

		if err := config.SaveToFile(&req.Config, configPath); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to persist config", "details": err.Error()})
			return
		}
		completed, next := evaluateSetupProgress(&req.Config)
		resp := setupStatus{
			Status:         "ok",
			Message:        "Configuration saved. Restart the server to exit setup mode.",
			ConfigPath:     configPath,
			ConfigPresent:  true,
			ConfigValid:    true,
			CompletedSteps: completed,
			NextStep:       next,
		}
		body := gin.H{
			"status":           resp.Status,
			"message":          resp.Message,
			"config_path":      resp.ConfigPath,
			"config_present":   resp.ConfigPresent,
			"config_valid":     resp.ConfigValid,
			"completed_steps":  resp.CompletedSteps,
			"next_step":        resp.NextStep,
			"restart_required": true,
		}
		c.JSON(http.StatusOK, body)
		if req.Restart && stop != nil {
			go func() {
				// Give response time to flush before shutdown.
				time.Sleep(500 * time.Millisecond)
				stop()
			}()
		}
	})
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true, "service": "goconnect-server", "mode": "setup"})
	})
	r.NoRoute(func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/setup")
	})
}

func buildSetupStatus(configPath string) setupStatus {
	state := setupStatus{
		Status:     "setup-mode",
		ConfigPath: configPath,
		NextStep:   "mode",
	}
	cfg, present, parseErr, validationErr := loadConfigForStatus(configPath)
	state.ConfigPresent = present
	if parseErr != nil {
		state.ValidationError = parseErr.Error()
		return state
	}
	if cfg != nil {
		state.Mode = setupBackendOrDefault(cfg.Database)
		state.CompletedSteps, state.NextStep = evaluateSetupProgress(cfg)
	}
	if validationErr == nil && cfg != nil {
		state.ConfigValid = true
	} else if validationErr != nil {
		state.ValidationError = validationErr.Error()
	}
	if state.NextStep == "" {
		state.NextStep = "mode"
	}
	return state
}

func evaluateSetupProgress(cfg *config.Config) ([]string, string) {
	completed := []string{}
	next := "mode"
	if cfg == nil {
		return completed, next
	}
	if isModeStepComplete(cfg) {
		completed = append(completed, "mode")
		next = "admin"
	} else {
		return completed, next
	}
	if isAdminStepComplete(cfg) {
		completed = append(completed, "admin")
		next = "finalize"
	} else {
		return completed, next
	}
	if err := cfg.Validate(); err == nil {
		completed = append(completed, "finalize")
		next = setupCompletedStep
	}
	return completed, next
}

func isModeStepComplete(cfg *config.Config) bool {
	backend := setupBackendOrDefault(cfg.Database)
	switch backend {
	case "sqlite":
		return strings.TrimSpace(cfg.Database.SQLitePath) != ""
	case "postgres":
		return strings.TrimSpace(cfg.Database.Host) != "" &&
			strings.TrimSpace(cfg.Database.Port) != "" &&
			strings.TrimSpace(cfg.Database.User) != "" &&
			strings.TrimSpace(cfg.Database.DBName) != ""
	case "memory":
		return true
	default:
		return false
	}
}

func isAdminStepComplete(cfg *config.Config) bool {
	if len(cfg.JWT.Secret) < 32 {
		return false
	}
	if strings.TrimSpace(cfg.WireGuard.ServerEndpoint) == "" {
		return false
	}
	if len(cfg.WireGuard.ServerPubKey) != 44 {
		return false
	}
	return true
}

func loadConfigForStatus(path string) (*config.Config, bool, error, error) {
	stat, err := os.Stat(path)
	if err != nil || stat.IsDir() {
		return nil, false, nil, fmt.Errorf("config file not found")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, true, err, err
	}
	cfg := &config.Config{}
	if err := yaml.Unmarshal(content, cfg); err != nil {
		return nil, true, err, err
	}
	if err := cfg.Validate(); err != nil {
		return cfg, true, nil, err
	}
	return cfg, true, nil, nil
}

func setupBackendOrDefault(db config.DatabaseConfig) string {
	if db.Backend == "" {
		return "postgres"
	}
	return strings.ToLower(db.Backend)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func startHTTPServer(ctx context.Context, cfg *config.Config, handler http.Handler) {
	srv := &http.Server{
		Addr:              serverAddress(cfg),
		Handler:           handler,
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

// getEnvOrDefault gets an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
