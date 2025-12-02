package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/config"
	"github.com/orhaniscoding/goconnect/server/internal/database"
	"github.com/orhaniscoding/goconnect/server/internal/handler"
	"github.com/orhaniscoding/goconnect/server/internal/metrics"
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
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set Gin mode based on environment
	if cfg.Server.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database
	db, err := initDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Register Prometheus metrics
	metrics.Register()

	// Initialize HTTP router
	router := setupRouter(cfg)

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
		log.Printf("GoConnect Server %s starting on %s", version, cfg.Server.Address())
		log.Printf("Environment: %s", cfg.Server.Environment)
		log.Printf("Database: %s", cfg.Database.Backend)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
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

// setupRouter configures the Gin router with all routes
func setupRouter(cfg *config.Config) *gin.Engine {
	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(handler.RequestIDMiddleware())
	router.Use(handler.CORSMiddleware())
	router.Use(metrics.GinMiddleware())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"version": version,
		})
	})

	// Metrics endpoint (Prometheus)
	router.GET("/metrics", metrics.Handler())

	// API v1 routes
	v1 := router.Group("/v1")
	{
		// Auth routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", placeholder("register"))
			auth.POST("/login", placeholder("login"))
			auth.POST("/refresh", placeholder("refresh"))
			auth.POST("/logout", placeholder("logout"))
		}

		// Health/Info
		v1.GET("/info", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"version":     version,
				"commit":      commit,
				"environment": cfg.Server.Environment,
			})
		})
	}

	return router
}

// placeholder returns a handler that indicates the endpoint is not fully wired
func placeholder(name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{
			"error":   "not_implemented",
			"message": fmt.Sprintf("Endpoint '%s' requires full service wiring", name),
		})
	}
}
