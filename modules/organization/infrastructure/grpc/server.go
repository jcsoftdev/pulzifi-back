package grpc

import (
	"context"

	"github.com/google/uuid"
	createorgapp "github.com/jcsoftdev/pulzifi-back/modules/organization/application/create_organization"
	getorgapp "github.com/jcsoftdev/pulzifi-back/modules/organization/application/get_organization"
	"github.com/jcsoftdev/pulzifi-back/modules/organization/infrastructure/grpc/pb"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// OrganizationServiceServer implements the pb.OrganizationServiceServer interface
type OrganizationServiceServer struct {
	pb.UnimplementedOrganizationServiceServer
	createOrgHandler *createorgapp.CreateOrganizationHandler
	getOrgHandler    *getorgapp.GetOrganizationHandler
}

// NewOrganizationServiceServer creates a new gRPC server
func NewOrganizationServiceServer(
	createOrgHandler *createorgapp.CreateOrganizationHandler,
	getOrgHandler *getorgapp.GetOrganizationHandler,
) *OrganizationServiceServer {
	return &OrganizationServiceServer{
		createOrgHandler: createOrgHandler,
		getOrgHandler:    getOrgHandler,
	}
}

// CreateOrganization implements the gRPC CreateOrganization method
func (s *OrganizationServiceServer) CreateOrganization(
	ctx context.Context,
	req *pb.CreateOrganizationRequest,
) (*pb.CreateOrganizationReply, error) {

	// Parse owner user ID from request
	ownerUserID, err := uuid.Parse(req.OwnerUserId)
	if err != nil {
		logger.Error("Invalid owner user ID", zap.Error(err))
		return nil, err
	}

	appReq := &createorgapp.Request{
		Name:      req.Name,
		Subdomain: req.Subdomain,
	}

	resp, err := s.createOrgHandler.Handle(ctx, appReq, ownerUserID)
	if err != nil {
		logger.Error("Failed to create organization via gRPC", zap.Error(err))
		return nil, err
	}

	// Convert application response to gRPC response
	pbReply := &pb.CreateOrganizationReply{
		Organization: &pb.Organization{
			Id:         resp.ID.String(),
			Name:       resp.Name,
			Subdomain:  resp.Subdomain,
			SchemaName: resp.SchemaName,
			CreatedAt:  resp.CreatedAt,
		},
	}

	return pbReply, nil
}

// GetOrganization implements the gRPC GetOrganization method
func (s *OrganizationServiceServer) GetOrganization(
	ctx context.Context,
	req *pb.GetOrganizationRequest,
) (*pb.GetOrganizationReply, error) {

	// Parse organization ID from request
	orgID, err := uuid.Parse(req.Id)
	if err != nil {
		logger.Error("Invalid organization ID", zap.Error(err))
		return nil, err
	}

	resp, err := s.getOrgHandler.Handle(ctx, orgID)
	if err != nil {
		logger.Error("Failed to get organization via gRPC", zap.Error(err))
		return nil, err
	}

	// Convert application response to gRPC response
	pbReply := &pb.GetOrganizationReply{
		Organization: &pb.Organization{
			Id:          resp.ID.String(),
			Name:        resp.Name,
			Subdomain:   resp.Subdomain,
			SchemaName:  resp.SchemaName,
			OwnerUserId: resp.OwnerUserID.String(),
			CreatedAt:   resp.CreatedAt,
			UpdatedAt:   resp.UpdatedAt,
		},
	}

	return pbReply, nil
}

// GetOrganizationBySubdomain implements the gRPC GetOrganizationBySubdomain method
func (s *OrganizationServiceServer) GetOrganizationBySubdomain(
	ctx context.Context,
	req *pb.GetOrganizationBySubdomainRequest,
) (*pb.GetOrganizationBySubdomainReply, error) {

	// TODO: Implement GetOrganizationBySubdomain handler in application layer
	logger.Warn("GetOrganizationBySubdomain not yet implemented")
	return nil, nil
}
