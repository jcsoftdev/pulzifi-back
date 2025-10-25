package config

import (
	"fmt"
	"os"

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
	HTTPPort    string
	GRPCPort    string
	Environment string
	LogLevel    string
	JWTSecret   string

	// Module
	ModuleName string
}

func Load() *Config {
	// Load .env file if it exists
	_ = godotenv.Load()

	return &Config{
		DBHost:           getEnv("DB_HOST", "localhost"),
		DBPort:           getEnv("DB_PORT", "5434"),
		DBName:           getEnv("DB_NAME", "pulzifi"),
		DBUser:           getEnv("DB_USER", "pulzifi_user"),
		DBPassword:       getEnv("DB_PASSWORD", "pulzifi_password"),
		DBMaxConnections: 25,
		RedisHost:        getEnv("REDIS_HOST", "localhost"),
		RedisPort:        getEnv("REDIS_PORT", "6379"),
		RedisPassword:    getEnv("REDIS_PASSWORD", ""),
		KafkaBrokers:     getEnv("KAFKA_BROKERS", "localhost:9092"),
		HTTPPort:         getEnv("HTTP_PORT", "8080"),
		GRPCPort:         getEnv("GRPC_PORT", "9000"),
		Environment:      getEnv("ENVIRONMENT", "development"),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
		JWTSecret:        getEnv("JWT_SECRET", "secret"),
		ModuleName:       getEnv("MODULE_NAME", "unknown"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func (c *Config) String() string {
	return fmt.Sprintf("Config{Module: %s, DBHost: %s, HTTPPort: %s, GRPCPort: %s}",
		c.ModuleName, c.DBHost, c.HTTPPort, c.GRPCPort)
}
