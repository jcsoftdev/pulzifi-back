package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jcsoftdev/pulzifi-back/docs"
	"github.com/jcsoftdev/pulzifi-back/shared/config"
	"github.com/jcsoftdev/pulzifi-back/shared/database"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	middlewarex "github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
	"github.com/jcsoftdev/pulzifi-back/shared/static"
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
// @host localhost:8080
func main() {
	// Load configuration
	cfg := config.Load()
	logger.Info("Starting Pulzifi Backend - Unified Monolith", zap.String("environment", cfg.Environment))

	// Connect to database
	db, err := database.Connect(cfg)
	if err != nil {
		logger.Error("Failed to connect to database", zap.Error(err))
		os.Exit(1)
	}
	defer db.Close()

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

	// Health endpoint
	httpRouter.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"pulzifi-backend-monolith","time":` +
			string(rune(time.Now().Unix())) + `}`))
	})

	// Create module registry and register all modules
	logger.Info("Registering module routes...")
	registry := router.NewRegistry(logger.Logger)
	registerAllModulesInternal(registry, db)

	// Register routes from all modules under /api/v1
	v1Router := chi.NewRouter()

	// Add tenant middleware to extract tenant from request
	v1Router.Use(middlewarex.TenantMiddleware)

	// Setup Swagger UI on v1 router
	swagger.SetupSwaggerForChi(v1Router)

	registry.RegisterAll(v1Router)
	httpRouter.Mount("/api/v1", v1Router)

	// Setup frontend (proxy o static)
	static.Setup(httpRouter, cfg.FrontendURL, cfg.StaticDir, logger.Logger)

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
		startGRPCServer(ctx, cfg.GRPCPort)
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
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
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

// startGRPCServer starts the gRPC server
func startGRPCServer(ctx context.Context, port string) {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Error("Failed to listen on gRPC port", zap.Error(err))
		return
	}

	grpcServer := grpc.NewServer()

	// Note: Register gRPC services here when available
	// Example: pb.RegisterYourServiceServer(grpcServer, &yourServiceServer{})

	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
	}()

	logger.Info("Starting gRPC server", zap.String("port", port))
	if err := grpcServer.Serve(lis); err != nil && err.Error() != "transport: Stop called" {
		logger.Error("gRPC server error", zap.Error(err))
	}
}
