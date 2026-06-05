package config

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// fatalFunc is the function used to abort startup on misconfiguration.
// Overridable in tests via testFatal.
var fatalFunc = func(msg string) { log.Fatal(msg) }

type Config struct {
	// Database
	DBHost           string
	DBPort           string
	DBName           string
	DBUser           string
	DBPassword       string
	DBMaxConnections int

	// Redis
	RedisHost     string
	RedisPort     string
	RedisPassword string

	// Server
	HTTPPort             string
	GRPCPort             string
	Environment          string
	LogLevel             string
	JWTSecret            string
	JWTExpiration        time.Duration
	JWTRefreshExpiration time.Duration
	ResetTokenTTL        time.Duration // RESET_TOKEN_TTL — password-reset token lifetime (default 1h)
	CookieDomain         string

	// Frontend
	FrontendURL string
	NextJSURL   string // Internal Next.js server URL for reverse proxy
	StaticDir   string
	AppDomain   string // Apex application domain (e.g. "pulzifi.com"); requests on the apex carry no tenant

	// CORS
	CORSAllowedOrigins string
	CORSAllowedMethods string
	CORSAllowedHeaders string

	// Module
	ModuleName string

	// MinIO / S3
	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOBucket    string
	MinIOUseSSL    bool
	MinIOPublicURL string

	// Snapshot object storage provider
	ObjectStorageProvider string
	CloudinaryCloudName   string
	CloudinaryAPIKey      string
	CloudinaryAPISecret   string
	CloudinaryFolder      string

	// Extractor
	ExtractorURL    string
	ExtractorAPIKey string // EXTRACTOR_API_KEY — must match scraper's SCRAPER_API_KEY

	// AI / OpenRouter
	OpenRouterAPIKey      string
	OpenRouterModel       string
	OpenRouterVisionModel string
	// AIBaseURL overrides the OpenAI-compatible endpoint for insight/vision
	// generation. Empty = OpenRouter default. Set it to target a pay-per-token
	// provider (e.g. https://api.deepinfra.com/v1/openai) or self-hosted Ollama
	// (http://ollama:11434/v1) without code changes.
	AIBaseURL          string
	PixelDiffThreshold float64

	// Email (Resend)
	ResendAPIKey     string
	EmailFromAddress string
	EmailFromName    string

	// OAuth
	GoogleClientID       string
	GoogleClientSecret   string
	GitHubClientID       string
	GitHubClientSecret   string
	OAuthRedirectBaseURL string

	// Rate Limiting
	RateLimitRequests int
	RateLimitWindow   time.Duration

	// Integrations
	IntegrationTokenKey          string // 32-byte hex; validated at wiring if any provider configured
	IntegrationOAuthRedirectBase string // e.g. https://app.com (root domain)
	SlackClientID                string
	SlackClientSecret            string
	DeliveryMaxAttempts          int
	DeliveryPollInterval         time.Duration
	DeliveryWorkerPoolSize       int

	// Phase 2: additional integration providers
	DiscordClientID     string
	DiscordClientSecret string
	TwilioAccountSID    string
	TwilioAuthToken     string
	TwilioFromNumber    string
	TwilioPaidPlans     []string // comma-separated env (TWILIO_PAID_PLANS)

	// IntegrationPaidPlans is the set of plan codes that unlock all notification
	// channels (Slack, Discord, Teams, Gmail, Outlook, Sheets, Twilio).
	// Orgs with NO active plan or a plan code outside this list are "free" and
	// may only receive email notifications.
	// Default: trial,starter,pro,enterprise (all seeded plans).
	// Env var: INTEGRATION_PAID_PLANS (comma-separated).
	IntegrationPaidPlans []string

	// Phase 3
	SheetsClientID          string
	SheetsClientSecret      string
	MicrosoftClientID       string
	MicrosoftClientSecret   string
	TwilioQuotaPaidPerMonth int

	// Gmail integration — reuses GoogleClientID/Secret; guarded by explicit flag
	GmailIntegrationEnabled bool // GMAIL_INTEGRATION_ENABLED

	// Stripe / Billing
	// Note: checkout success/cancel URLs and portal return URL are built
	// per-request from the tenant subdomain in the HTTP layer — they are not
	// configurable via env. Only the Stripe API credentials live here.
	StripeSecretKey      string
	StripeWebhookSecret  string
	StripePublishableKey string
	BillingEnabled       bool // BILLING_ENABLED — gates /billing/* routes

	// Trial (self-serve onboarding)
	TrialDays           int    // TRIAL_DAYS — number of days the no-card trial lasts
	TrialChecksPerMonth int    // TRIAL_CHECKS_PER_MONTH — monthly checks allowed during trial
	TrialExpiryCron     string // TRIAL_EXPIRY_CRON — cron expression for the daily trial-expirer job

	// Snapshot retention
	RetentionJobInterval time.Duration // RETENTION_JOB_INTERVAL — how often the retention sweep runs (default 24h)
	DefaultRetentionDays int           // DEFAULT_RETENTION_DAYS — fallback when tenant has no storage_period_days set (default 7)

	// Snapshot storage isolation
	SnapshotBucketPrivate bool          // SNAPSHOT_BUCKET_PRIVATE — when true, EnsureBucket skips public-read policy and reads use presigned URLs (default false)
	SnapshotPresignTTL    time.Duration // SNAPSHOT_PRESIGN_TTL — lifetime of presigned GET URLs (default 15m)

	// Durable EventBus (transactional outbox)
	EventBusProvider     string        // EVENT_BUS_PROVIDER — memory|outbox (default: memory)
	OutboxPollInterval   time.Duration // OUTBOX_POLL_INTERVAL — relay poll cadence (default 1s)
	OutboxBatchSize      int           // OUTBOX_BATCH_SIZE — rows claimed per relay tick (default 100)
	OutboxMaxAttempts    int           // OUTBOX_MAX_ATTEMPTS — attempts before DLQ (default 8)
	OutboxStaleThreshold time.Duration // OUTBOX_STALE_THRESHOLD — stuck-in-publishing recovery window (default 30s)

	// HA / multi-instance
	// InstanceMode controls whether the server runs in single-instance or HA mode.
	// Accepted values: "single" (default) | "ha".
	// In "ha" mode Redis is required; if not reachable the server will fatal at startup.
	InstanceMode string // INSTANCE_MODE — single|ha (default: single)

	// Per-subsystem backend overrides. Empty string means "derive from InstanceMode":
	//   single → "memory", ha → "redis".
	// Explicit "redis" or "memory" overrides InstanceMode derivation.
	NonceBackend     string // NONCE_BACKEND — redis|memory|"" (default: "")
	BrokerBackend    string // BROKER_BACKEND — redis|memory|"" (default: "")
	RateLimitBackend string // RATELIMIT_BACKEND — redis|memory|"" (default: "")

	// SchedulerMode selects the monitoring scheduler claim strategy.
	// Accepted values: "poll" (default, claims on pages.last_checked_at) |
	// "queue" (claims on pages.next_run_at with FOR UPDATE SKIP LOCKED for
	// multi-instance safety). Any other value falls back to "poll".
	SchedulerMode string // SCHEDULER_MODE — poll|queue (default: poll)

	// TrustedProxyCIDRs is the list of CIDR ranges whose X-Forwarded-For header
	// entries are trusted for real-IP extraction. Empty slice = no trusted proxies
	// (RemoteAddr is used directly). Invalid CIDRs are logged and skipped.
	TrustedProxyCIDRs []net.IPNet // TRUSTED_PROXY_CIDRS — CSV of CIDRs (default: empty)

	// Auth-specific rate limiting (applied per route to /login and /forgot-password).
	AuthRateLimitRequests int           // AUTH_RATE_LIMIT_REQUESTS — default 5
	AuthRateLimitWindow   time.Duration // AUTH_RATE_LIMIT_WINDOW — default 15m

	// Ollama (section naming for change diffs). Optional; disabled by default.
	OllamaURL     string
	OllamaModel   string
	OllamaEnabled bool
}

func Load() *Config {
	// Load .env file if it exists
	_ = godotenv.Load()

	env := getEnv("ENVIRONMENT", "development")
	isProd := env == "production"

	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		if isProd {
			fatalFunc("FATAL: JWT_SECRET must be set in production — refusing to start with insecure default")
		}
		log.Println("WARNING: JWT_SECRET is not set — using insecure default 'secret'. Set JWT_SECRET before deploying to production.")
		jwtSecret = "secret"
	} else if isProd && len(jwtSecret) < 32 {
		fatalFunc("FATAL: JWT_SECRET must be at least 32 bytes in production — refusing to start with a weak secret")
	}

	extractorAPIKey := getEnv("EXTRACTOR_API_KEY", "")
	if extractorAPIKey == "" {
		if isProd {
			fatalFunc("FATAL: EXTRACTOR_API_KEY must be set in production — refusing to start without scraper authentication")
		}
		log.Println("WARNING: EXTRACTOR_API_KEY is not set — scraper endpoints are unauthenticated. Set this variable before deploying to production.")
	}

	return &Config{
		DBHost:                mustGetEnv("DB_HOST"),
		DBPort:                mustGetEnv("DB_PORT"),
		DBName:                mustGetEnv("DB_NAME"),
		DBUser:                mustGetEnv("DB_USER"),
		DBPassword:            mustGetEnv("DB_PASSWORD"),
		DBMaxConnections:      25,
		RedisHost:             getEnv("REDIS_HOST", ""),
		RedisPort:             getEnv("REDIS_PORT", "6379"),
		RedisPassword:         getEnv("REDIS_PASSWORD", ""),
		HTTPPort:              getEnvFallback("HTTP_PORT", "PORT", "3000"),
		GRPCPort:              getEnv("GRPC_PORT", "9000"),
		Environment:           env,
		LogLevel:              getEnv("LOG_LEVEL", "info"),
		JWTSecret:             jwtSecret,
		JWTExpiration:         getEnvDurationSeconds("JWT_EXPIRATION", 900),            // Default 15 minutes
		JWTRefreshExpiration:  getEnvDurationSeconds("JWT_REFRESH_EXPIRATION", 604800), // Default 7 days
		ResetTokenTTL:         getEnvDuration("RESET_TOKEN_TTL", 1*time.Hour),          // Default 1 hour
		CookieDomain:          getEnv("COOKIE_DOMAIN", ""),
		FrontendURL:           getEnv("FRONTEND_URL", ""),
		NextJSURL:             getEnv("NEXTJS_URL", "http://localhost:3001"),
		StaticDir:             getEnv("STATIC_DIR", "./frontend/dist"),
		AppDomain:             getEnv("APP_DOMAIN", deriveAppDomain(getEnv("FRONTEND_URL", ""))),
		CORSAllowedOrigins:    getEnv("CORS_ALLOWED_ORIGINS", ""),
		CORSAllowedMethods:    getEnv("CORS_ALLOWED_METHODS", "GET,POST,PUT,DELETE,OPTIONS,PATCH"),
		CORSAllowedHeaders:    getEnv("CORS_ALLOWED_HEADERS", "Content-Type,Authorization,X-Tenant"),
		ModuleName:            getEnv("MODULE_NAME", "unknown"),
		MinIOEndpoint:         getEnv("MINIO_ENDPOINT", ""),
		MinIOAccessKey:        getEnv("MINIO_ACCESS_KEY", ""),
		MinIOSecretKey:        getEnv("MINIO_SECRET_KEY", ""),
		MinIOBucket:           getEnv("MINIO_BUCKET", ""),
		MinIOUseSSL:           getEnvBool("MINIO_USE_SSL", false),
		MinIOPublicURL:        getEnv("MINIO_PUBLIC_URL", ""),
		ObjectStorageProvider: getEnv("OBJECT_STORAGE_PROVIDER", "minio"),
		CloudinaryCloudName:   getEnv("CLOUDINARY_CLOUD_NAME", ""),
		CloudinaryAPIKey:      getEnv("CLOUDINARY_API_KEY", ""),
		CloudinaryAPISecret:   getEnv("CLOUDINARY_API_SECRET", ""),
		CloudinaryFolder:      getEnv("CLOUDINARY_FOLDER", ""),
		ExtractorURL:          mustGetEnv("EXTRACTOR_URL"),
		ExtractorAPIKey:       extractorAPIKey,
		OpenRouterAPIKey:      getEnv("OPENROUTER_API_KEY", ""),
		OpenRouterModel:       getEnv("OPENROUTER_MODEL", "mistralai/mistral-7b-instruct:free"),
		OpenRouterVisionModel: getEnv("OPENROUTER_VISION_MODEL", ""),
		AIBaseURL:             getEnv("AI_BASE_URL", ""),
		PixelDiffThreshold:    getEnvFloat("PIXEL_DIFF_THRESHOLD", 0.001),
		ResendAPIKey:          getEnv("RESEND_API_KEY", ""),
		EmailFromAddress:      getEnv("EMAIL_FROM_ADDRESS", ""),
		EmailFromName:         getEnv("EMAIL_FROM_NAME", ""),
		GoogleClientID:        getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:    getEnv("GOOGLE_CLIENT_SECRET", ""),
		GitHubClientID:        getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret:    getEnv("GITHUB_CLIENT_SECRET", ""),
		OAuthRedirectBaseURL:  getEnv("OAUTH_REDIRECT_BASE_URL", ""),
		RateLimitRequests:     getEnvInt("RATE_LIMIT_REQUESTS", 500),
		RateLimitWindow:       getEnvDuration("RATE_LIMIT_WINDOW", 60*time.Second),

		IntegrationTokenKey:          getEnv("INTEGRATION_TOKEN_KEY", ""),
		IntegrationOAuthRedirectBase: getEnv("INTEGRATION_OAUTH_REDIRECT_BASE", ""),
		SlackClientID:                getEnv("SLACK_CLIENT_ID", ""),
		SlackClientSecret:            getEnv("SLACK_CLIENT_SECRET", ""),
		DeliveryMaxAttempts:          getEnvInt("DELIVERY_MAX_ATTEMPTS", 5),
		DeliveryPollInterval:         getEnvDuration("DELIVERY_POLL_INTERVAL", 5*time.Second),
		DeliveryWorkerPoolSize:       getEnvInt("DELIVERY_WORKER_POOL_SIZE", 10),

		DiscordClientID:      getEnv("DISCORD_CLIENT_ID", ""),
		DiscordClientSecret:  getEnv("DISCORD_CLIENT_SECRET", ""),
		TwilioAccountSID:     getEnv("TWILIO_ACCOUNT_SID", ""),
		TwilioAuthToken:      getEnv("TWILIO_AUTH_TOKEN", ""),
		TwilioFromNumber:     getEnv("TWILIO_FROM_NUMBER", ""),
		TwilioPaidPlans:      splitCSV(getEnv("TWILIO_PAID_PLANS", "pro,business")),
		IntegrationPaidPlans: splitCSV(getEnv("INTEGRATION_PAID_PLANS", "trial,starter,pro,enterprise")),

		SheetsClientID:          getEnv("SHEETS_CLIENT_ID", ""),
		SheetsClientSecret:      getEnv("SHEETS_CLIENT_SECRET", ""),
		MicrosoftClientID:       getEnv("MICROSOFT_CLIENT_ID", ""),
		MicrosoftClientSecret:   getEnv("MICROSOFT_CLIENT_SECRET", ""),
		TwilioQuotaPaidPerMonth: getEnvInt("TWILIO_QUOTA_PAID_PER_MONTH", 500),

		GmailIntegrationEnabled: getEnvBool("GMAIL_INTEGRATION_ENABLED", false),

		StripeSecretKey:      getEnv("STRIPE_SECRET_KEY", ""),
		StripeWebhookSecret:  getEnv("STRIPE_WEBHOOK_SECRET", ""),
		StripePublishableKey: getEnv("STRIPE_PUBLISHABLE_KEY", ""),
		BillingEnabled:       getEnvBool("BILLING_ENABLED", false),

		TrialDays:           getEnvInt("TRIAL_DAYS", 14),
		TrialChecksPerMonth: getEnvInt("TRIAL_CHECKS_PER_MONTH", 500),
		TrialExpiryCron:     getEnv("TRIAL_EXPIRY_CRON", "0 0 * * *"),

		RetentionJobInterval: getEnvDuration("RETENTION_JOB_INTERVAL", 24*time.Hour),
		DefaultRetentionDays: getEnvInt("DEFAULT_RETENTION_DAYS", 7),

		SnapshotBucketPrivate: getEnvBool("SNAPSHOT_BUCKET_PRIVATE", false),
		SnapshotPresignTTL:    getEnvDuration("SNAPSHOT_PRESIGN_TTL", 15*time.Minute),

		EventBusProvider:     getEnv("EVENT_BUS_PROVIDER", "memory"),
		OutboxPollInterval:   getEnvDuration("OUTBOX_POLL_INTERVAL", 1*time.Second),
		OutboxBatchSize:      getEnvInt("OUTBOX_BATCH_SIZE", 100),
		OutboxMaxAttempts:    getEnvInt("OUTBOX_MAX_ATTEMPTS", 8),
		OutboxStaleThreshold: getEnvDuration("OUTBOX_STALE_THRESHOLD", 30*time.Second),

		InstanceMode:          getEnv("INSTANCE_MODE", "single"),
		NonceBackend:          getEnv("NONCE_BACKEND", ""),
		BrokerBackend:         getEnv("BROKER_BACKEND", ""),
		RateLimitBackend:      getEnv("RATELIMIT_BACKEND", ""),
		SchedulerMode:         getEnv("SCHEDULER_MODE", "poll"),
		TrustedProxyCIDRs:     parseTrustedProxyCIDRs(getEnv("TRUSTED_PROXY_CIDRS", "")),
		AuthRateLimitRequests: getEnvInt("AUTH_RATE_LIMIT_REQUESTS", 5),
		AuthRateLimitWindow:   getEnvDuration("AUTH_RATE_LIMIT_WINDOW", 15*time.Minute),

		OllamaURL:     getEnv("OLLAMA_URL", ""),
		OllamaModel:   getEnv("OLLAMA_MODEL", "qwen3:4b"),
		OllamaEnabled: getEnvBool("OLLAMA_ENABLED", false),
	}
}

// ResolveBackend resolves the effective backend for a subsystem.
// If explicit is "redis" or "memory", it is returned as-is.
// Otherwise the backend is derived from InstanceMode:
//
//	"ha"     → "redis"
//	"single" → "memory"
func (c *Config) ResolveBackend(explicit string) string {
	if explicit == "redis" || explicit == "memory" {
		return explicit
	}
	if c.InstanceMode == "ha" {
		return "redis"
	}
	return "memory"
}

// parseTrustedProxyCIDRs parses a comma-separated list of CIDR strings.
// Invalid entries are logged and skipped; the remainder is returned.
// An empty input returns nil.
func parseTrustedProxyCIDRs(csv string) []net.IPNet {
	if csv == "" {
		return nil
	}
	parts := strings.Split(csv, ",")
	out := make([]net.IPNet, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		_, ipNet, err := net.ParseCIDR(p)
		if err != nil {
			log.Printf("WARNING: TRUSTED_PROXY_CIDRS: invalid CIDR %q — skipping: %v", p, err)
			continue
		}
		out = append(out, *ipNet)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// deriveAppDomain extracts the apex host from a frontend URL (e.g.
// "https://pulzifi.com" -> "pulzifi.com"). Returns "" when the URL is empty or
// unparseable, which keeps the legacy dev subdomain heuristic in place.
func deriveAppDomain(frontendURL string) string {
	if frontendURL == "" {
		return ""
	}
	u, err := url.Parse(frontendURL)
	if err != nil {
		return ""
	}
	return u.Hostname()
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func mustGetEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		fatalFunc(fmt.Sprintf("FATAL: required environment variable %q is not set — refusing to start", key))
	}
	return value
}

// getEnvFallback returns the first env var that is set, falling back to defaultValue.
// Useful for Railway which injects PORT instead of HTTP_PORT.
func getEnvFallback(keys ...string) string {
	// Last element is the default value, all others are env var keys
	defaultValue := keys[len(keys)-1]
	for _, key := range keys[:len(keys)-1] {
		if value, exists := os.LookupEnv(key); exists && value != "" {
			return value
		}
	}
	return defaultValue
}

func getEnvDurationSeconds(key string, defaultSeconds int) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if seconds, err := strconv.Atoi(value); err == nil {
			return time.Duration(seconds) * time.Second
		}
	}
	return time.Duration(defaultSeconds) * time.Second
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value, exists := os.LookupEnv(key); exists {
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}

func (c *Config) String() string {
	return fmt.Sprintf("Config{Module: %s, DBHost: %s, HTTPPort: %s, GRPCPort: %s}",
		c.ModuleName, c.DBHost, c.HTTPPort, c.GRPCPort)
}
