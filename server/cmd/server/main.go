package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/audit"
	"github.com/orhaniscoding/goconnect/server/internal/handler"
	"github.com/orhaniscoding/goconnect/server/internal/metrics"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/orhaniscoding/goconnect/server/internal/service"
)

var (
	version = "dev"
	commit  = "none"
	date    = "2025-09-22"
	builtBy = "orhaniscoding"
)

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	asyncAudit := flag.Bool("audit-async", true, "enable async audit buffering")
	auditQueue := flag.Int("audit-queue", 1024, "audit async queue size")
	auditWorkers := flag.Int("audit-workers", 1, "audit async worker count")
	flag.Parse()

	if *showVersion {
		fmt.Printf("goconnect-server %s (commit %s, build %s) built by %s\n", version, commit, date, builtBy)
		return
	}

	// Initialize repositories
	networkRepo := repository.NewInMemoryNetworkRepository()
	idempotencyRepo := repository.NewInMemoryIdempotencyRepository()
	membershipRepo := repository.NewInMemoryMembershipRepository()
	joinRepo := repository.NewInMemoryJoinRequestRepository()
	ipamRepo := repository.NewInMemoryIPAM()

	// Initialize services
	networkService := service.NewNetworkService(networkRepo, idempotencyRepo)
	membershipService := service.NewMembershipService(networkRepo, membershipRepo, joinRepo, idempotencyRepo)
	ipamService := service.NewIPAMService(networkRepo, membershipRepo, ipamRepo)
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
		} else {
			// Fallback remains stdout auditor on error; do not log secrets/DSN
		}
	}
	if *asyncAudit {
		aud = audit.NewAsyncAuditor(aud, audit.WithQueueSize(*auditQueue), audit.WithWorkers(*auditWorkers))
	}
	membershipService.SetAuditor(aud)
	networkService.SetAuditor(aud)
	ipamService.SetAuditor(aud)

	// Initialize handlers
	networkHandler := handler.NewNetworkHandler(networkService, membershipService).WithIPAM(ipamService)

	// Setup router
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(metrics.GinMiddleware())

	// Register basic routes
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true, "service": "goconnect-server"})
	})

	r.POST("/v1/auth/login", func(c *gin.Context) {
		c.JSON(200, gin.H{"data": gin.H{"access_token": "dev", "refresh_token": "dev"}})
	})

	// Role middleware must run early to populate membership_role
	r.Use(handler.RoleMiddleware(membershipRepo))
	// Register network routes
	handler.RegisterNetworkRoutes(r, networkHandler)

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
