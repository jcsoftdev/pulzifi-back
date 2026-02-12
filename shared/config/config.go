package config

import (
	"fmt"
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

	// Kafka
	KafkaBrokers string

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

	// Extractor
	ExtractorURL string
}

func Load() *Config {
	// Load .env file if it exists
	_ = godotenv.Load()

	return &Config{
		DBHost:               getEnv("DB_HOST", "localhost"),
		DBPort:               getEnv("DB_PORT", "5434"),
		DBName:               getEnv("DB_NAME", "pulzifi"),
		DBUser:               getEnv("DB_USER", "pulzifi_user"),
		DBPassword:           getEnv("DB_PASSWORD", "pulzifi_password"),
		DBMaxConnections:     25,
		RedisHost:            getEnv("REDIS_HOST", "localhost"),
		RedisPort:            getEnv("REDIS_PORT", "6379"),
		RedisPassword:        getEnv("REDIS_PASSWORD", ""),
		KafkaBrokers:         getEnv("KAFKA_BROKERS", "localhost:9092"),
		HTTPPort:             getEnv("HTTP_PORT", "9090"),
		GRPCPort:             getEnv("GRPC_PORT", "9000"),
		Environment:          getEnv("ENVIRONMENT", "development"),
		LogLevel:             getEnv("LOG_LEVEL", "info"),
		JWTSecret:            getEnv("JWT_SECRET", "secret"),
		JWTExpiration:        getEnvDurationSeconds("JWT_EXPIRATION", 900),            // Default 15 minutes
		JWTRefreshExpiration: getEnvDurationSeconds("JWT_REFRESH_EXPIRATION", 604800), // Default 7 days
		CookieDomain:         getEnv("COOKIE_DOMAIN", ""),
		FrontendURL:          getEnv("FRONTEND_URL", ""),
		StaticDir:            getEnv("STATIC_DIR", "./frontend/dist"),
		CORSAllowedOrigins:   getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:9090,http://*.localhost:9090"),
		CORSAllowedMethods:   getEnv("CORS_ALLOWED_METHODS", "GET,POST,PUT,DELETE,OPTIONS,PATCH"),
		CORSAllowedHeaders:   getEnv("CORS_ALLOWED_HEADERS", "Content-Type,Authorization,X-Tenant"),
		ModuleName:           getEnv("MODULE_NAME", "unknown"),
		MinIOEndpoint:        getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKey:       getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecretKey:       getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinIOBucket:          getEnv("MINIO_BUCKET", "pulzifi-snapshots"),
		MinIOUseSSL:          getEnvBool("MINIO_USE_SSL", false),
		MinIOPublicURL:       getEnv("MINIO_PUBLIC_URL", "http://localhost:9000"),
		ExtractorURL:         getEnv("EXTRACTOR_URL", "http://localhost:3000"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
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
