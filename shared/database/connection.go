package database

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/jcsoftdev/pulzifi-back/shared/config"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// Connect creates a connection pool to PostgreSQL with retry logic for transient
// DNS/network failures (common on Railway private networking).
func Connect(cfg *config.Config) (*sql.DB, error) {
	// Support DATABASE_URL (e.g. Railway managed Postgres) with fallback to individual vars
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			cfg.DBHost,
			cfg.DBPort,
			cfg.DBUser,
			cfg.DBPassword,
			cfg.DBName,
		)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.Error("Failed to open database connection", zap.Error(err))
		return nil, err
	}

	// Retry ping with exponential backoff to handle transient DNS/network errors
	maxRetries := 5
	backoff := 2 * time.Second
	for i := 0; i < maxRetries; i++ {
		if err = db.Ping(); err == nil {
			break
		}
		if i < maxRetries-1 {
			logger.Warn("Failed to ping database, retrying...",
				zap.Error(err),
				zap.Int("attempt", i+1),
				zap.Duration("backoff", backoff),
			)
			time.Sleep(backoff)
			backoff *= 2
		}
	}
	if err != nil {
		logger.Error("Failed to ping database after retries", zap.Error(err), zap.Int("attempts", maxRetries))
		db.Close()
		return nil, err
	}

	// Set connection pool parameters
	db.SetMaxOpenConns(cfg.DBMaxConnections)
	db.SetMaxIdleConns(cfg.DBMaxConnections / 2)

	logger.Info("Database connection established", zap.String("host", cfg.DBHost), zap.String("db", cfg.DBName))

	return db, nil
}
