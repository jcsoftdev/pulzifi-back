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
	"github.com/jcsoftdev/pulzifi-back/shared/noncestore"
	"github.com/jcsoftdev/pulzifi-back/shared/pubsub"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
	"github.com/jcsoftdev/pulzifi-back/shared/static"
	"github.com/jcsoftdev/pulzifi-back/shared/swagger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// @title Pulzifi API
// @version 1.0
// @description Pulzifi backend monolith API
// @contact.name API Support
// @contact.email support@pulzifi.com
// @license.name MIT
// @basePath /api/v1
// @schemes http https
// @host localhost:3000
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter "Bearer" followed by a space and your JWT access token
func main() { //nolint:gocyclo // startup orchestration: complexity is inherent
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
	defer func() { _ = db.Close() }()

	// Initialize Redis for caching.
	// In single mode this is optional (warn + continue).
	// In HA mode Redis is required — fatal if unreachable.
	if err := cache.InitRedis(cfg); err != nil {
		if cfg.InstanceMode == "ha" {
			logger.Logger.Fatal("HA mode requires Redis but connection failed — refusing to start",
				zap.Error(err),
				zap.String("instance_mode", cfg.InstanceMode),
			)
		}
		logger.Warn("Failed to initialize Redis - Caching disabled", zap.Error(err))
	} else {
		defer func() { _ = cache.CloseRedis() }()
		logger.Info("Redis initialized successfully")
	}

	// Initialize Event Bus — provider is "memory" by default, so this is
	// backward-compatible. Flip EVENT_BUS_PROVIDER=outbox to enable the
	// transactional outbox. Note: the relay is started by the worker process,
	// not the API server.
	eventBus, err := eventbus.NewFromConfig(cfg.EventBusProvider, cfg.OutboxBatchSize, cfg.OutboxMaxAttempts, db)
	if err != nil {
		logger.Error("Failed to initialize event bus", zap.Error(err))
		os.Exit(1) //nolint:gocritic // fatal startup error; OS reclaims db connection
	}
	defer eventBus.Close()
	logger.Info("Event Bus initialized", zap.String("provider", cfg.EventBusProvider))

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
					if idx := strings.Index(allowed, "*."); idx >= 0 {
						prefix := allowed[:idx]   // "http://" or "https://"
						suffix := allowed[idx+1:] // ".pulzifi.com" or ".pulzifi.com:3000"
						if strings.HasPrefix(origin, prefix) && strings.HasSuffix(origin, suffix) {
							return true
						}
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

	// Rate limiting middleware — API paths only; skip Next.js static assets.
	// Backend is selected via RATELIMIT_BACKEND (or derived from INSTANCE_MODE).
	var globalLimiter middlewarex.Limiter
	switch cfg.ResolveBackend(cfg.RateLimitBackend) {
	case "redis":
		if rdb := cache.GetRedisClient(); rdb != nil {
			globalLimiter = middlewarex.NewRedisRateLimiter(rdb, cfg.RateLimitRequests, cfg.RateLimitWindow, cfg.TrustedProxyCIDRs)
			logger.Info("Rate limiter: Redis backend")
		} else {
			logger.Warn("Rate limiter: Redis backend requested but client nil — falling back to memory")
			globalLimiter = middlewarex.NewRateLimiterWithTrustedProxies(cfg.RateLimitRequests, cfg.RateLimitWindow, cfg.TrustedProxyCIDRs)
		}
	default:
		globalLimiter = middlewarex.NewRateLimiterWithTrustedProxies(cfg.RateLimitRequests, cfg.RateLimitWindow, cfg.TrustedProxyCIDRs)
		logger.Info("Rate limiter: memory backend")
	}
	defer globalLimiter.Stop()
	httpRouter.Use(func(next http.Handler) http.Handler {
		limited := globalLimiter.Handler(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api") {
				limited.ServeHTTP(w, r)
			} else {
				next.ServeHTTP(w, r)
			}
		})
	})

	// Auth-specific rate limiter — applied per route to /login and /forgot-password.
	// Uses the same backend as the global rate limiter.
	var authLimiterBackend middlewarex.Limiter
	switch cfg.ResolveBackend(cfg.RateLimitBackend) {
	case "redis":
		if rdb := cache.GetRedisClient(); rdb != nil {
			authLimiterBackend = middlewarex.NewRedisRateLimiter(rdb, cfg.AuthRateLimitRequests, cfg.AuthRateLimitWindow, cfg.TrustedProxyCIDRs)
		} else {
			authLimiterBackend = middlewarex.NewRateLimiterWithTrustedProxies(cfg.AuthRateLimitRequests, cfg.AuthRateLimitWindow, cfg.TrustedProxyCIDRs)
		}
	default:
		authLimiterBackend = middlewarex.NewRateLimiterWithTrustedProxies(cfg.AuthRateLimitRequests, cfg.AuthRateLimitWindow, cfg.TrustedProxyCIDRs)
	}
	defer authLimiterBackend.Stop()
	authLimiter := middlewarex.NewAuthLimiter(authLimiterBackend, cfg.TrustedProxyCIDRs)

	// Health endpoint
	httpRouter.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy","service":"pulzifi-backend-monolith","time":` +
			strconv.FormatInt(time.Now().Unix(), 10) + `}`))
	})

	// Build backend-agnostic nonce store and pub/sub brokers.
	// Backend is resolved from NONCE_BACKEND / BROKER_BACKEND (or INSTANCE_MODE).
	// In HA mode Redis is already guaranteed by the fail-fast above.
	var (
		nonceStoreImpl    noncestore.NonceStore
		checkBrokerImpl   pubsub.CheckBroker
		insightBrokerImpl pubsub.InsightBroker
	)
	rdb := cache.GetRedisClient()

	nonceBackend := cfg.ResolveBackend(cfg.NonceBackend)
	if nonceBackend == "redis" {
		if rdb != nil {
			nonceStoreImpl = noncestore.NewRedisStore(rdb)
			logger.Info("Nonce store: Redis backend")
		} else {
			logger.Warn("Nonce store: Redis requested but client nil — falling back to memory")
			nonceStoreImpl = noncestore.New()
		}
	} else {
		nonceStoreImpl = noncestore.New()
		logger.Info("Nonce store: memory backend")
	}

	brokerBackend := cfg.ResolveBackend(cfg.BrokerBackend)
	if brokerBackend == "redis" {
		if rdb != nil {
			checkBrokerImpl = pubsub.NewRedisCheckBroker(ctx, rdb)
			insightBrokerImpl = pubsub.NewRedisInsightBroker(ctx, rdb)
			logger.Info("Pub/sub brokers: Redis backend")
		} else {
			logger.Warn("Pub/sub brokers: Redis requested but client nil — falling back to memory")
			checkBrokerImpl = pubsub.NewCheckBroker()
			insightBrokerImpl = pubsub.NewInsightBroker()
		}
	} else {
		checkBrokerImpl = pubsub.NewCheckBroker()
		insightBrokerImpl = pubsub.NewInsightBroker()
		logger.Info("Pub/sub brokers: memory backend")
	}

	// Create module registry and register all modules
	logger.Info("Registering module routes...")
	registry := router.NewRegistry(logger.Logger)
	bffHandler, integrationMod := registerAllModulesInternal(registry, db, eventBus, enableWorkers, nonceStoreImpl, checkBrokerImpl, insightBrokerImpl)

	// Mount BFF auth routes BEFORE /api/v1 (these handle cookies/nonces).
	// The /login route additionally gets the per-user auth rate limiter.
	logger.Info("Registering BFF auth handler at /api/auth")
	httpRouter.Route("/api/auth", func(r chi.Router) {
		// Apply auth-specific rate limiter to /login only (before other BFF routes).
		r.With(authLimiter.KeyedHandler("login")).Post("/login", bffHandler.HandleLogin)
		// Remaining BFF routes without the per-user limiter.
		bffHandler.RegisterRoutesExceptLogin(r)
	})
	logger.Info("BFF auth handler registered successfully")

	// OAuth callback for integrations: lands on root domain (no X-Tenant), so register
	// directly on the parent router BEFORE the tenant-middleware-wrapped v1Router mounts.
	httpRouter.Get("/api/v1/integrations/oauth/{provider}/callback", integrationMod.HandleOAuthCallback)

	// Register routes from all modules under /api/v1
	v1Router := chi.NewRouter()

	// Add tenant middleware FIRST to extract subdomain and resolve to schema
	// This must come before LoggingMiddleware so tenant is in context for logging
	v1Router.Use(middlewarex.TenantMiddleware(db))

	// Add response logging middleware to capture request/response details
	v1Router.Use(middlewarex.ResponseLoggerMiddleware)

	// Add logging middleware AFTER tenant middleware to capture tenant in logs
	v1Router.Use(middlewarex.LoggingMiddleware)

	// Trial guard: when the active org plan is the trial plan and trial_ends_at
	// has elapsed without a converted_at, all writes (except /billing/checkout)
	// receive HTTP 402 Payment Required. Reads always pass.
	trialGuard := middlewarex.NewTrialGuard(middlewarex.NewPostgresTrialPlanReader(db))
	v1Router.Use(trialGuard.Handler)

	// Apply auth-specific rate limiter to /auth/forgot-password.
	// Chi does not support per-route middleware post-registration, so we use a
	// path-matching inline middleware applied to the v1Router before module registration.
	forgotPasswordLimiter := authLimiter.KeyedHandler("forgot-password")
	v1Router.Use(func(next http.Handler) http.Handler {
		forgotLimited := forgotPasswordLimiter(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost && r.URL.Path == "/auth/forgot-password" {
				forgotLimited.ServeHTTP(w, r)
			} else {
				next.ServeHTTP(w, r)
			}
		})
	})

	// Setup Swagger UI on v1 router (bypassed by isPublicPath). MUST come after all
	// v1Router.Use(...) calls: chi panics if middleware is registered after any route,
	// and SetupSwaggerForChi registers the /swagger/doc.json route.
	swagger.SetupSwaggerForChi(v1Router)

	registry.RegisterAll(v1Router)
	httpRouter.Mount("/api/v1", v1Router)

	// Reverse proxy: unmatched routes → Next.js for page rendering
	static.Setup(httpRouter, cfg.NextJSURL, cfg.StaticDir, db, logger.Logger)

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
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 15 * time.Second,
		// ReadTimeout is intentionally not set: it applies a deadline on the
		// underlying connection that cancels r.Context() during handler
		// execution, killing long-lived SSE streams. ReadHeaderTimeout only
		// limits reading request headers, leaving handlers free to run as
		// long as they need. Individual handlers enforce their own deadlines.
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("HTTP server shutdown error", zap.Error(err))
		}
	}()

	// Optional TLS for local dev: when both cert/key files are provided the server
	// terminates HTTPS directly. Mirrors prod, where Traefik terminates TLS upstream
	// and these env vars stay unset. HTTPS gives the browser a secure context on custom
	// dev domains like lvh.me — which Chrome requires before it sends Sec-Fetch-* headers
	// that Payload's admin cookie auth depends on (otherwise the CMS login loops).
	certFile := os.Getenv("TLS_CERT_FILE")
	keyFile := os.Getenv("TLS_KEY_FILE")
	if certFile != "" && keyFile != "" {
		logger.Info("HTTPS server starting", zap.String("port", port))
		if err := server.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTPS server error", zap.Error(err))
		}
		return
	}

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
	createOrgHandler := createorgapp.NewCreateOrganizationHandler(orgRepo, orgSvc, nil)
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
