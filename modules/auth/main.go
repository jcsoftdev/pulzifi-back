package main

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jcsoftdev/pulzifi-back/shared/cache"
	"github.com/jcsoftdev/pulzifi-back/shared/config"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.Load()
	logger.Info("Starting Auth Service", zap.String("config", cfg.String()))

	// Initialize Redis
	if err := cache.InitRedis(cfg); err != nil {
		logger.Error("Failed to initialize Redis", zap.Error(err))
		panic(err)
	}
	defer cache.CloseRedis()
	logger.Info("Redis initialized successfully")

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy", "service": "auth"})
	})

	router.POST("/register", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "register endpoint"})
	})

	router.POST("/login", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "login endpoint"})
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
