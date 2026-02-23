package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

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
	CookieDomain         string

	// Frontend
	FrontendURL string
	StaticDir   string

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
	ExtractorURL string

	// AI / OpenRouter
	OpenRouterAPIKey string
	OpenRouterModel  string

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
}

func Load() *Config {
	// Load .env file if it exists
	_ = godotenv.Load()

	env := getEnv("ENVIRONMENT", "development")
	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		if env == "production" {
			log.Fatal("FATAL: JWT_SECRET must be set in production — refusing to start with insecure default")
		}
		log.Println("WARNING: JWT_SECRET is not set — using insecure default 'secret'. Set JWT_SECRET before deploying to production.")
		jwtSecret = "secret"
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
		HTTPPort:              getEnvFallback("HTTP_PORT", "PORT", "9090"),
		GRPCPort:              getEnv("GRPC_PORT", "9000"),
		Environment:           env,
		LogLevel:              getEnv("LOG_LEVEL", "info"),
		JWTSecret:             jwtSecret,
		JWTExpiration:         getEnvDurationSeconds("JWT_EXPIRATION", 900),            // Default 15 minutes
		JWTRefreshExpiration:  getEnvDurationSeconds("JWT_REFRESH_EXPIRATION", 604800), // Default 7 days
		CookieDomain:          getEnv("COOKIE_DOMAIN", ""),
		FrontendURL:           getEnv("FRONTEND_URL", ""),
		StaticDir:             getEnv("STATIC_DIR", "./frontend/dist"),
		CORSAllowedOrigins:    mustGetEnv("CORS_ALLOWED_ORIGINS"),
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
		OpenRouterAPIKey:      getEnv("OPENROUTER_API_KEY", ""),
		OpenRouterModel:       getEnv("OPENROUTER_MODEL", "mistralai/mistral-7b-instruct:free"),
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
	}
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
		log.Fatalf("FATAL: required environment variable %q is not set — refusing to start", key)
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
