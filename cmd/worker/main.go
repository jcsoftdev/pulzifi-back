package main

import (
	"os"
	"os/signal"
	"syscall"

	monitoring "github.com/jcsoftdev/pulzifi-back/modules/monitoring/infrastructure/http"
	"github.com/jcsoftdev/pulzifi-back/shared/config"
	"github.com/jcsoftdev/pulzifi-back/shared/database"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()
	logger.Info("Starting Pulzifi Worker Service", zap.String("environment", cfg.Environment))

	// Connect to database
	db, err := database.Connect(cfg)
	if err != nil {
		logger.Error("Failed to connect to database", zap.Error(err))
		os.Exit(1)
	}
	defer db.Close()

	// Initialize Monitoring Module (which contains the Scheduler logic)
	// Note: We don't need EventBus here for the scheduler/orchestrator loop as implemented currently
	// but NewModuleWithDB requires it. We can pass nil if we don't need to listen to API events here.
	// But if we want to support cross-module events later, we might need it.
	// For now, pass nil.
	mod := monitoring.NewModuleWithDB(db, nil, nil, "")

	// Cast to concrete type to access StartBackgroundProcesses
	if monitoringModule, ok := mod.(*monitoring.Module); ok {
		monitoringModule.StartBackgroundProcesses()
	} else {
		logger.Logger.Fatal("Failed to cast monitoring module")
	}

	logger.Info("Worker Service is running...")

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutdown signal received, shutting down worker...")
}
