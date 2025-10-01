package main

import (
    "flag"
    "fmt"
    "net/http"

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
	baseAud := audit.NewStdoutAuditor()
	var aud audit.Auditor = audit.WrapWithMetrics(baseAud, metrics.IncAudit)
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

	// Start server
	fmt.Printf("GoConnect Server starting on :8080...\n")
	if err := http.ListenAndServe(":8080", r); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}
