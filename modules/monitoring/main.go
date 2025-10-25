package main

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jcsoftdev/pulzifi-back/shared/config"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.Load()
	logger.Info("Starting Monitoring Service", zap.String("config", cfg.String()))

	router := gin.Default()
	router.GET("/health", middleware.HealthCheck())

	router.GET("/api/checks", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "list checks"})
	})

	router.POST("/api/checks", func(c *gin.Context) {
		c.JSON(201, gin.H{"message": "create check"})
	})

	go func() {
		logger.Info("Starting HTTP server", zap.String("port", cfg.HTTPPort))
		if err := router.Run(":" + cfg.HTTPPort); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", zap.Error(err))
		}
	}()

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		logger.Error("Failed to listen on gRPC port", zap.Error(err))
		panic(err)
	}

	grpcServer := grpc.NewServer()
	logger.Info("Starting gRPC server", zap.String("port", cfg.GRPCPort))

	if err := grpcServer.Serve(lis); err != nil {
		logger.Error("gRPC server error", zap.Error(err))
		panic(err)
	}
}
