package main

import (
	"context"
	"net"

	"github.com/gin-gonic/gin"
	"github.com/jcsoftdev/pulzifi-back/modules/organization/application/create_organization"
	"github.com/jcsoftdev/pulzifi-back/modules/organization/application/get_organization"
	"github.com/jcsoftdev/pulzifi-back/modules/organization/domain/services"
	grpcadapter "github.com/jcsoftdev/pulzifi-back/modules/organization/infrastructure/grpc"
	pb "github.com/jcsoftdev/pulzifi-back/modules/organization/infrastructure/grpc/pb"
	httpadapter "github.com/jcsoftdev/pulzifi-back/modules/organization/infrastructure/http"
	"github.com/jcsoftdev/pulzifi-back/modules/organization/infrastructure/messaging"
	"github.com/jcsoftdev/pulzifi-back/modules/organization/infrastructure/persistence"
	"github.com/jcsoftdev/pulzifi-back/shared/config"
	"github.com/jcsoftdev/pulzifi-back/shared/database"
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	_ "github.com/jcsoftdev/pulzifi-back/modules/organization/docs"
)

// @title Pulzifi Organization API
// @version 1.0
// @description Organization service API for the Pulzifi monitoring platform
// @host localhost:8081
// @basePath /
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @securitySchemes.BearerAuth apiKey
// @name Authorization
// @in header
// @description Type "Bearer" followed by a space and JWT token
func main() {
	cfg := config.Load()
	logger.Info("Starting Organization Service", zap.String("config", cfg.String()))

	// Connect to database
	db, err := database.Connect(cfg)
	if err != nil {
		logger.Error("Failed to connect to database", zap.Error(err))
		panic(err)
	}
	defer db.Close()
	logger.Info("Connected to database")

	// Initialize repositories
	orgRepo := persistence.NewOrganizationPostgresRepository(db)

	// Initialize domain services
	orgService := services.NewOrganizationService()

	// Initialize messaging (EventBus for MVP, Kafka for later)
	// We use the singleton EventBus for in-memory communication
	messageBus := eventbus.GetInstance()

	// Initialize messaging adapters
	messagePublisher := messaging.NewPublisher(messageBus)
	messageSubscriber := messaging.NewSubscriber(messageBus, db)

	// Initialize application handlers
	createOrgHandler := create_organization.NewCreateOrganizationHandler(orgRepo, orgService, db, messagePublisher)
	getOrgHandler := get_organization.NewGetOrganizationHandler(orgRepo)

	// Start Kafka subscriber in background
	go func() {
		messageSubscriber.ListenToEvents(context.Background())
	}()

	// Setup HTTP Router
	router := gin.Default()
	router.GET("/health", middleware.HealthCheck())

	// Setup organization routes
	httpRouter := httpadapter.NewRouter(createOrgHandler, getOrgHandler)
	httpRouter.Setup(router)

	// Start HTTP server in a goroutine
	go func() {
		logger.Info("Starting HTTP server", zap.String("port", cfg.HTTPPort))
		if err := router.Run(":" + cfg.HTTPPort); err != nil {
			logger.Error("HTTP server error", zap.Error(err))
		}
	}()

	// Setup gRPC server
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		logger.Error("Failed to listen on gRPC port", zap.Error(err))
		panic(err)
	}

	grpcServer := grpc.NewServer()

	// Register gRPC services
	grpcService := grpcadapter.NewOrganizationServiceServer(createOrgHandler, getOrgHandler, orgRepo)
	pb.RegisterOrganizationServiceServer(grpcServer, grpcService)

	logger.Info("Starting gRPC server", zap.String("port", cfg.GRPCPort))

	if err := grpcServer.Serve(lis); err != nil {
		logger.Error("gRPC server error", zap.Error(err))
		panic(err)
	}
}
