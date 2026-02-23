package database

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/jcsoftdev/pulzifi-back/shared/config"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// Connect creates a connection pool to PostgreSQL
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

	// Test the connection
	if err := db.Ping(); err != nil {
		logger.Error("Failed to ping database", zap.Error(err))
		return nil, err
	}

	// Set connection pool parameters
	db.SetMaxOpenConns(cfg.DBMaxConnections)
	db.SetMaxIdleConns(cfg.DBMaxConnections / 2)

	logger.Info("Database connection established", zap.String("host", cfg.DBHost), zap.String("db", cfg.DBName))

	return db, nil
}
