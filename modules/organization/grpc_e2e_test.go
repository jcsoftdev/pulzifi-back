package main_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	pb "github.com/jcsoftdev/pulzifi-back/modules/organization/infrastructure/grpc/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func dialGRPC(t *testing.T) *grpc.ClientConn {
	t.Helper()
	addr := getGRPCAddress()
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server at %s: %v", addr, err)
	}
	return conn
}

// TestCreateOrganizationE2E tests creating an organization end-to-end
func TestCreateOrganizationE2E(t *testing.T) {
	skipUnlessE2E(t)

	conn := dialGRPC(t)
	defer conn.Close()

	client := pb.NewOrganizationServiceClient(conn)

	subdomain := "e2e-test-" + uuid.New().String()[:8]
	req := &pb.CreateOrganizationRequest{
		Name:        "E2E Test Organization",
		Subdomain:   subdomain,
		OwnerUserId: testUserID,
	}

	resp, err := client.CreateOrganization(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to create organization: %v", err)
	}

	if resp == nil || resp.Organization == nil {
		t.Fatal("CreateOrganization returned nil response")
	}

	org := resp.Organization
	if org.Id == "" {
		t.Error("Organization ID is empty")
	}
	if org.Name != req.Name {
		t.Errorf("Expected name %q, got %q", req.Name, org.Name)
	}
	if org.Subdomain != req.Subdomain {
		t.Errorf("Expected subdomain %q, got %q", req.Subdomain, org.Subdomain)
	}
	if org.SchemaName == "" {
		t.Error("Schema name is empty")
	}
	if org.CreatedAt == "" {
		t.Error("Created at timestamp is empty")
	}
}

// TestGetOrganizationE2E tests retrieving an organization end-to-end
func TestGetOrganizationE2E(t *testing.T) {
	skipUnlessE2E(t)

	conn := dialGRPC(t)
	defer conn.Close()

	client := pb.NewOrganizationServiceClient(conn)

	subdomain := "get-e2e-test-" + uuid.New().String()[:8]
	createResp, err := client.CreateOrganization(context.Background(), &pb.CreateOrganizationRequest{
		Name:        "Get E2E Test Org",
		Subdomain:   subdomain,
		OwnerUserId: testUserID,
	})
	if err != nil {
		t.Fatalf("Failed to create organization: %v", err)
	}

	orgID := createResp.Organization.Id

	getResp, err := client.GetOrganization(context.Background(), &pb.GetOrganizationRequest{
		Id: orgID,
	})
	if err != nil {
		t.Fatalf("Failed to get organization: %v", err)
	}

	if getResp == nil || getResp.Organization == nil {
		t.Fatal("GetOrganization returned nil response")
	}

	org := getResp.Organization
	if org.Id != orgID {
		t.Errorf("Expected ID %q, got %q", orgID, org.Id)
	}
	if org.Name != "Get E2E Test Org" {
		t.Errorf("Expected name 'Get E2E Test Org', got %q", org.Name)
	}
	if org.Subdomain != subdomain {
		t.Errorf("Expected subdomain %q, got %q", subdomain, org.Subdomain)
	}
}

// TestCreateMultipleOrganizationsE2E tests creating multiple organizations
func TestCreateMultipleOrganizationsE2E(t *testing.T) {
	skipUnlessE2E(t)

	conn := dialGRPC(t)
	defer conn.Close()

	client := pb.NewOrganizationServiceClient(conn)

	orgCount := 3
	orgIDs := make([]string, orgCount)

	for i := 0; i < orgCount; i++ {
		subdomain := "multi-e2e-" + uuid.New().String()[:12]
		resp, err := client.CreateOrganization(context.Background(), &pb.CreateOrganizationRequest{
			Name:        "Multi Org " + uuid.New().String()[:8],
			Subdomain:   subdomain,
			OwnerUserId: testUserID,
		})
		if err != nil {
			t.Fatalf("Failed to create organization %d: %v", i+1, err)
		}
		orgIDs[i] = resp.Organization.Id
	}

	for i, orgID := range orgIDs {
		resp, err := client.GetOrganization(context.Background(), &pb.GetOrganizationRequest{Id: orgID})
		if err != nil {
			t.Fatalf("Failed to get organization %d: %v", i+1, err)
		}
		if resp.Organization == nil {
			t.Fatalf("Organization %d response is nil", i+1)
		}
	}
}

// TestValidationErrorsE2E tests validation error cases
func TestValidationErrorsE2E(t *testing.T) {
	skipUnlessE2E(t)

	conn := dialGRPC(t)
	defer conn.Close()

	client := pb.NewOrganizationServiceClient(conn)

	tests := []struct {
		name        string
		req         *pb.CreateOrganizationRequest
		shouldError bool
	}{
		{
			name: "valid organization",
			req: &pb.CreateOrganizationRequest{
				Name:        "Valid Org",
				Subdomain:   "valid-org-" + uuid.New().String()[:8],
				OwnerUserId: testUserID,
			},
			shouldError: false,
		},
		{
			name: "short subdomain",
			req: &pb.CreateOrganizationRequest{
				Name:        "Invalid Org",
				Subdomain:   "ab",
				OwnerUserId: testUserID,
			},
			shouldError: true,
		},
		{
			name: "subdomain with underscore",
			req: &pb.CreateOrganizationRequest{
				Name:        "Invalid Org",
				Subdomain:   "invalid_org",
				OwnerUserId: testUserID,
			},
			shouldError: true,
		},
		{
			name: "subdomain too long",
			req: &pb.CreateOrganizationRequest{
				Name:        "Invalid Org",
				Subdomain:   "this-is-a-very-long-subdomain-that-exceeds-sixty-three-characters-limit",
				OwnerUserId: testUserID,
			},
			shouldError: true,
		},
		{
			name: "empty name",
			req: &pb.CreateOrganizationRequest{
				Name:        "",
				Subdomain:   "test-org",
				OwnerUserId: testUserID,
			},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.CreateOrganization(context.Background(), tt.req)

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if resp == nil || resp.Organization == nil {
					t.Error("Response is nil")
				}
			}
		})
	}
}

// TestDuplicateSubdomainE2E tests that duplicate subdomains are rejected
func TestDuplicateSubdomainE2E(t *testing.T) {
	skipUnlessE2E(t)

	conn := dialGRPC(t)
	defer conn.Close()

	client := pb.NewOrganizationServiceClient(conn)

	subdomain := "duplicate-test-" + uuid.New().String()[:8]

	_, err := client.CreateOrganization(context.Background(), &pb.CreateOrganizationRequest{
		Name:        "First Org",
		Subdomain:   subdomain,
		OwnerUserId: testUserID,
	})
	if err != nil {
		t.Fatalf("Failed to create first organization: %v", err)
	}

	_, err = client.CreateOrganization(context.Background(), &pb.CreateOrganizationRequest{
		Name:        "Second Org",
		Subdomain:   subdomain,
		OwnerUserId: testUserID,
	})

	if err == nil {
		t.Error("Expected error when creating organization with duplicate subdomain, but got none")
	}
}

// TestInvalidOrganizationIDE2E tests getting non-existent organization
func TestInvalidOrganizationIDE2E(t *testing.T) {
	skipUnlessE2E(t)

	conn := dialGRPC(t)
	defer conn.Close()

	client := pb.NewOrganizationServiceClient(conn)

	tests := []struct {
		name string
		id   string
	}{
		{"valid UUID but non-existent", "550e8400-e29b-41d4-a716-446655440099"},
		{"invalid UUID format", "invalid-uuid"},
		{"empty ID", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.GetOrganization(context.Background(), &pb.GetOrganizationRequest{Id: tt.id})
			if err == nil {
				t.Error("Expected error but got nil")
			}
		})
	}
}
