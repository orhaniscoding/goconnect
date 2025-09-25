package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/audit"
	"github.com/orhaniscoding/goconnect/server/internal/handler"
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

	// Initialize services
	networkService := service.NewNetworkService(networkRepo, idempotencyRepo)
	membershipService := service.NewMembershipService(networkRepo, membershipRepo, joinRepo, idempotencyRepo)
	membershipService.SetAuditor(audit.NewStdoutAuditor())

	// Initialize handlers
	networkHandler := handler.NewNetworkHandler(networkService, membershipService)

	// Setup router
	r := gin.Default()

	// Register basic routes
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true, "service": "goconnect-server"})
	})

	r.POST("/v1/auth/login", func(c *gin.Context) {
		c.JSON(200, gin.H{"data": gin.H{"access_token": "dev", "refresh_token": "dev"}})
	})

	// Register network routes
	handler.RegisterNetworkRoutes(r, networkHandler)

	// Start server
	fmt.Printf("GoConnect Server starting on :8080...\n")
	if err := http.ListenAndServe(":8080", r); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}
