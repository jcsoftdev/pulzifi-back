package main

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/jcsoftdev/pulzifi-back/docs"
	createorgapp "github.com/jcsoftdev/pulzifi-back/modules/organization/application/create_organization"
	getorgapp "github.com/jcsoftdev/pulzifi-back/modules/organization/application/get_organization"
	orgservices "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/services"
	orggrpc "github.com/jcsoftdev/pulzifi-back/modules/organization/infrastructure/grpc"
	"github.com/jcsoftdev/pulzifi-back/modules/organization/infrastructure/grpc/pb"
	orgpersistence "github.com/jcsoftdev/pulzifi-back/modules/organization/infrastructure/persistence"
	"github.com/jcsoftdev/pulzifi-back/shared/cache"
	"github.com/jcsoftdev/pulzifi-back/shared/config"
	"github.com/jcsoftdev/pulzifi-back/shared/database"
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	middlewarex "github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
	"github.com/jcsoftdev/pulzifi-back/shared/swagger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var _ = docs.SwaggerInfo // Ensure docs is imported

// @title Pulzifi API
// @version 1.0
// @description Pulzifi backend monolith API
// @contact.name API Support
// @contact.email support@pulzifi.com
// @license.name MIT
// @basePath /api/v1
// @schemes http https
// @host localhost:9090
func main() {
	// Load configuration
	cfg := config.Load()
	logger.Info("Starting Pulzifi Backend - Unified Monolith",
		zap.String("environment", cfg.Environment),
		zap.String("cookie_domain", cfg.CookieDomain),
		zap.Bool("cookie_secure", cfg.Environment == "production"),
	)

	// Connect to database
	db, err := database.Connect(cfg)
	if err != nil {
		logger.Error("Failed to connect to database", zap.Error(err))
		os.Exit(1)
	}
	defer db.Close()

	// Initialize Redis for caching (Optional for MVP)
	if err := cache.InitRedis(cfg); err != nil {
		logger.Warn("Failed to initialize Redis - Caching disabled", zap.Error(err))
	} else {
		defer cache.CloseRedis()
		logger.Info("Redis initialized successfully")
	}

	// Initialize Event Bus (for MVP)
	eventBus := eventbus.GetInstance()
	logger.Info("Event Bus initialized")

	// Check if we should run in "All-in-One" mode or just "API" mode
	// Default to true for backward compatibility unless explicitly disabled
	enableWorkers := os.Getenv("ENABLE_WORKERS") != "false"

	if enableWorkers {
		logger.Info("Running in Monolith All-in-One mode (API + Workers)")
	} else {
		logger.Info("Running in API-only mode (Workers disabled)")
	}

	// Create and Start Server
	// Pass enableWorkers flag to registerAllModulesInternal
	// The srv variable was unused and registerAllModulesInternal signature was mismatching in the commented out block above
	// which I replaced. But registerAllModulesInternal returns void in line 128.
	// The logic should be:
	// 1. Setup registry
	// 2. Call registerAllModulesInternal with enableWorkers
	// 3. Mount routes

	// I will remove the redundant call I added earlier at line 77-79 because it is called later at line 128
	// and registerAllModulesInternal expects *router.Registry as first arg, not *config.Config.

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Channel to handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// WaitGroup to manage goroutines
	var wg sync.WaitGroup

	// Setup shared HTTP router using Chi
	httpRouter := chi.NewRouter()

	// Configure CORS middleware
	corsHandler := cors.Handler(cors.Options{
		AllowOriginFunc: func(r *http.Request, origin string) bool {
			// Allow all localhost subdomains in development
			if strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1") {
				return true
			}
			// Check configured origins (supports wildcard subdomains like https://*.pulzifi.com)
			allowedOrigins := strings.Split(cfg.CORSAllowedOrigins, ",")
			for _, allowed := range allowedOrigins {
				allowed = strings.TrimSpace(allowed)
				if allowed == origin {
					return true
				}
				// Wildcard subdomain matching: http://*.example.com matches http://foo.example.com
				if strings.HasPrefix(allowed, "http://*.") || strings.HasPrefix(allowed, "https://*.") {
					// Split at *. to get the suffix (e.g., "http://*.pulzifi.com" → suffix ".pulzifi.com")
					idx := strings.Index(allowed, "*.")
					prefix := allowed[:idx]  // "http://" or "https://"
					suffix := allowed[idx+1:] // ".pulzifi.com" or ".pulzifi.com:3000"
					if strings.HasPrefix(origin, prefix) && strings.HasSuffix(origin, suffix) {
						return true
					}
				}
			}
			return false
		},
		AllowedMethods:   strings.Split(cfg.CORSAllowedMethods, ","),
		AllowedHeaders:   strings.Split(cfg.CORSAllowedHeaders, ","),
		ExposedHeaders:   []string{"X-Request-ID", "Content-Length"},
		AllowCredentials: true,
		MaxAge:           300, // 5 minutes
	})
	httpRouter.Use(corsHandler)

	// Rate limiting middleware
	rateLimiter := middlewarex.NewRateLimiter(cfg.RateLimitRequests, cfg.RateLimitWindow)
	defer rateLimiter.Stop()
	httpRouter.Use(rateLimiter.Handler)

	// Health endpoint
	httpRouter.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"pulzifi-backend-monolith","time":` +
			strconv.FormatInt(time.Now().Unix(), 10) + `}`))
	})

	// Create module registry and register all modules
	logger.Info("Registering module routes...")
	registry := router.NewRegistry(logger.Logger)
	registerAllModulesInternal(registry, db, eventBus, enableWorkers)

	// Register routes from all modules under /api/v1
	v1Router := chi.NewRouter()

	// Add tenant middleware FIRST to extract subdomain and resolve to schema
	// This must come before LoggingMiddleware so tenant is in context for logging
	v1Router.Use(middlewarex.TenantMiddleware(db))

	// Add response logging middleware to capture request/response details
	v1Router.Use(middlewarex.ResponseLoggerMiddleware)

	// Add logging middleware AFTER tenant middleware to capture tenant in logs
	v1Router.Use(middlewarex.LoggingMiddleware)

	// Setup Swagger UI on v1 router (bypassed by isPublicPath)
	swagger.SetupSwaggerForChi(v1Router)

	registry.RegisterAll(v1Router)
	httpRouter.Mount("/api/v1", v1Router)

	// Frontend proxy disabled - Frontend runs independently on localhost:3000
	// Nginx handles routing and CORS
	// static.Setup(httpRouter, cfg.FrontendURL, cfg.StaticDir, logger.Logger)

	// Start HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		startHTTPServer(ctx, httpRouter, cfg.HTTPPort)
	}()

	// Start gRPC server (for inter-module communication if needed)
	wg.Add(1)
	go func() {
		defer wg.Done()
		startGRPCServer(ctx, cfg.GRPCPort, db)
	}()

	logger.Info("Pulzifi Backend monolith is running")
	logger.Info("HTTP API available at", zap.String("url", "http://localhost:"+cfg.HTTPPort+"/api/v1"))
	logger.Info("Swagger UI available at", zap.String("url", "http://localhost:"+cfg.HTTPPort+"/api/v1/swagger/"))
	logger.Info("gRPC server available at", zap.String("port", cfg.GRPCPort))
	logger.Info("Modules loaded", zap.Int("count", registry.Count()))

	// Wait for shutdown signal
	<-sigChan
	logger.Info("Shutdown signal received, gracefully shutting down...")
	cancel()

	// Wait for all goroutines to finish
	wg.Wait()
	logger.Info("Pulzifi Backend monolith stopped successfully")
}

// startHTTPServer starts the HTTP server and listens for context cancellation
func startHTTPServer(ctx context.Context, router chi.Router, port string) {
	server := &http.Server{
		Addr:        ":" + port,
		Handler:     router,
		ReadTimeout: 15 * time.Second,
		// WriteTimeout is intentionally 0 (disabled) to support long-lived
		// SSE connections that can take up to 120s to stream their response.
		// Individual handlers enforce their own per-request deadlines.
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("HTTP server shutdown error", zap.Error(err))
		}
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("HTTP server error", zap.Error(err))
	}
}

// startGRPCServer starts the gRPC server and registers services
func startGRPCServer(ctx context.Context, port string, db *sql.DB) {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Error("Failed to listen on gRPC port", zap.Error(err))
		return
	}

	grpcServer := grpc.NewServer()

	// Register Organization gRPC service
	orgRepo := orgpersistence.NewOrganizationPostgresRepository(db)
	orgSvc := orgservices.NewOrganizationService()
	createOrgHandler := createorgapp.NewCreateOrganizationHandler(orgRepo, orgSvc, db, nil)
	getOrgHandler := getorgapp.NewGetOrganizationHandler(orgRepo)
	orgServiceServer := orggrpc.NewOrganizationServiceServer(createOrgHandler, getOrgHandler, orgRepo)
	pb.RegisterOrganizationServiceServer(grpcServer, orgServiceServer)

	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
	}()

	logger.Info("Starting gRPC server", zap.String("port", port))
	if err := grpcServer.Serve(lis); err != nil && err.Error() != "transport: Stop called" {
		logger.Error("gRPC server error", zap.Error(err))
	}
}
