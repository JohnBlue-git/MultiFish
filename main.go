package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"multifish/config"
	"multifish/handler"
	"multifish/middleware"
	"multifish/utility"
)

// ResponseError is defined in utility/errors.go
type ResponseError = utility.ResponseError

// ========== Main Function ==========

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "", "Path to configuration file (optional)")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		// Use stderr for pre-logger errors
		os.Stderr.WriteString("Failed to load configuration: " + err.Error() + "\n")
		os.Exit(1)
	}

	// Initialize logger with configured log level
	utility.InitLogger(cfg.LogLevel)
	log := utility.GetLogger()

	// Log configuration
	log.Info().
		Int("port", cfg.Port).
		Str("logLevel", cfg.LogLevel).
		Int("workerPoolSize", cfg.WorkerPoolSize).
		Str("logsDir", cfg.LogsDir).
		Bool("rateLimitEnabled", cfg.RateLimitEnabled).
		Float64("rateLimitRate", cfg.RateLimitRate).
		Int("rateLimitBurst", cfg.RateLimitBurst).
		Bool("authEnabled", cfg.Auth.Enabled).
		Str("authMode", cfg.Auth.Mode).
		Msg("Configuration loaded")

	// Set Gin mode based on log level
	if cfg.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.Default()

	// Add rate limiting middleware if enabled
	if cfg.RateLimitEnabled {
		rateLimiter := middleware.NewRateLimiter(rate.Limit(cfg.RateLimitRate), cfg.RateLimitBurst)
		router.Use(rateLimiter.RateLimitMiddleware())
		
		log.Info().
			Float64("rate", cfg.RateLimitRate).
			Int("burst", cfg.RateLimitBurst).
			Msg("Rate limiting enabled")
	} else {
		log.Warn().Msg("Rate limiting is DISABLED - API is vulnerable to abuse")
	}

	// Add authentication middleware if enabled
	if cfg.Auth.Enabled && cfg.Auth.Mode != "none" {
		router.Use(middleware.AuthMiddleware(cfg.Auth))
		
		log.Info().
			Str("mode", cfg.Auth.Mode).
			Msg("Authentication enabled")
	} else {
		log.Warn().Msg("Authentication is DISABLED - API is publicly accessible")
	}

	// Root service
	router.GET("/MultiFish/v1", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"@odata.type":    "#ServiceRoot.v1_0_0.ServiceRoot",
			"@odata.id":      "/MultiFish/v1",
			"Id":             "MultiFish",
			"Name":           "MultiFish Service",
			"RedfishVersion": "1.0.0",
			"Platform": gin.H{
				"@odata.id": "/MultiFish/v1/Platform",
			},
			"JobService": gin.H{
				"@odata.id": "/MultiFish/v1/JobService",
			},
		})
	})

	// Platform routes
	handler.PlatformRoutes(router)

	// Job service routes (now uses config for worker pool size)
	handler.JobServiceRoutes(router, cfg)

	// Manager routes
	handler.ManagerRoutes(router)

	// Create HTTP server
	srv := &http.Server{
		Addr:    cfg.GetServerAddr(),
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Info().Str("address", cfg.GetServerAddr()).Msg("MultiFish API server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down server...")

	// Cleanup all machines and close connections
	log.Info().Msg("Cleaning up machine connections...")
	handler.PlatformMgr.CleanupAll()

	// Stop job scheduler
	log.Info().Msg("Stopping job scheduler...")
	if handler.JobService != nil {
		handler.JobService.Stop()
	}

	// Graceful shutdown with configurable timeout
	shutdownTimeout := time.Duration(cfg.ShutdownTimeout) * time.Second
	log.Info().Msgf("Graceful shutdown initiated with timeout: %s", shutdownTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited")
}
