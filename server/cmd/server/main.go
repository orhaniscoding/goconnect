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
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/audit"
	"github.com/orhaniscoding/goconnect/server/internal/database"
	"github.com/orhaniscoding/goconnect/server/internal/handler"
	"github.com/orhaniscoding/goconnect/server/internal/metrics"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/orhaniscoding/goconnect/server/internal/service"
	ws "github.com/orhaniscoding/goconnect/server/internal/websocket"
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

	if *showVersion {
		fmt.Printf("goconnect-server %s (commit %s, build %s) built by %s\n", version, commit, date, builtBy)
		return
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
		fmt.Println("Using in-memory repositories (no data persistence)")
	}

	// Initialize services
	networkService := service.NewNetworkService(networkRepo, idempotencyRepo)
	membershipService := service.NewMembershipService(networkRepo, membershipRepo, joinRepo, idempotencyRepo)
	ipamService := service.NewIPAMService(networkRepo, membershipRepo, ipamRepo)
	authService := service.NewAuthService(userRepo, tenantRepo)

	peerProvisioningService := service.NewPeerProvisioningService(peerRepo, deviceRepo, networkRepo, membershipRepo, ipamRepo)
	deviceService := service.NewDeviceService(deviceRepo, userRepo, peerRepo)
	deviceService.SetPeerProvisioning(peerProvisioningService)

	chatService := service.NewChatService(chatRepo, userRepo)

	metrics.Register()
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
	if *asyncAudit {
		aud = audit.NewAsyncAuditor(aud, audit.WithQueueSize(*auditQueue), audit.WithWorkers(*auditWorkers))
	}
	membershipService.SetAuditor(aud)
	networkService.SetAuditor(aud)
	ipamService.SetAuditor(aud)
	deviceService.SetAuditor(aud)
	chatService.SetAuditor(aud)

	// Initialize handlers
	networkHandler := handler.NewNetworkHandler(networkService, membershipService).WithIPAM(ipamService)
	authHandler := handler.NewAuthHandler(authService)
	deviceHandler := handler.NewDeviceHandler(deviceService)
	chatHandler := handler.NewChatHandler(chatService)
	uploadHandler := handler.NewUploadHandler("./uploads", "/uploads")

	// Initialize WebSocket components
	// Circular dependency resolution: Handler -> Hub -> Handler
	wsMsgHandler := ws.NewDefaultMessageHandler(nil, chatService, membershipService)
	hub := ws.NewHub(wsMsgHandler)
	wsMsgHandler.SetHub(hub)

	// Start Hub
	go hub.Run(context.Background())

	// Initialize WebSocket HTTP handler
	webSocketHandler := handler.NewWebSocketHandler(hub)

	// Setup router
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(metrics.GinMiddleware())

	// Register basic routes
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true, "service": "goconnect-server"})
	})

	// Register auth routes (some require auth)
	handler.RegisterAuthRoutes(r, authHandler, handler.AuthMiddleware(authService))

	// Register network routes (auth + role middleware applied within)
	handler.RegisterNetworkRoutes(r, networkHandler, authService, membershipRepo)

	// Register device routes
	handler.RegisterDeviceRoutes(r, deviceHandler, handler.AuthMiddleware(authService))

	// Register chat routes
	handler.RegisterChatRoutes(r, chatHandler, handler.AuthMiddleware(authService))

	// Register upload routes
	handler.RegisterUploadRoutes(r, uploadHandler, handler.AuthMiddleware(authService))
	r.Static("/uploads", "./uploads")

	// Register WebSocket route
	r.GET("/ws", handler.AuthMiddleware(authService), webSocketHandler.HandleUpgrade)

	// Metrics endpoint
	r.GET("/metrics", metrics.Handler())

	// Audit integrity export (best-effort) – future: restrict via RBAC
	// Pass through the underlying sqlite auditor if present so ExportIntegrity works.
	if sqliteAudRef != nil {
		r.GET("/v1/audit/integrity", handler.AuditIntegrityHandler(sqliteAudRef))
	} else {
		r.GET("/v1/audit/integrity", handler.AuditIntegrityHandler(aud))
	}

	// Start server with timeouts
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           r,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
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
