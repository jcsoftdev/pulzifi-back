package main_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	pb "github.com/jcsoftdev/pulzifi-back/modules/organization/infrastructure/grpc/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TestCreateOrganizationE2E tests creating an organization end-to-end
func TestCreateOrganizationE2E(t *testing.T) {
	conn, err := grpc.Dial("localhost:9082", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewOrganizationServiceClient(conn)

	// Create organization
	subdomain := "e2e-test-" + uuid.New().String()[:8]
	req := &pb.CreateOrganizationRequest{
		Name:        "E2E Test Organization",
		Subdomain:   subdomain,
		OwnerUserId: "770e8400-e29b-41d4-a716-446655440000",
	}

	resp, err := client.CreateOrganization(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to create organization: %v", err)
	}

	if resp == nil || resp.Organization == nil {
		t.Fatal("CreateOrganization returned nil response")
	}

	org := resp.Organization

	// Verify response fields
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

	t.Logf("✓ Created organization: ID=%s, Name=%s, Subdomain=%s", org.Id, org.Name, org.Subdomain)
}

// TestGetOrganizationE2E tests retrieving an organization end-to-end
func TestGetOrganizationE2E(t *testing.T) {
	conn, err := grpc.Dial("localhost:9082", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewOrganizationServiceClient(conn)

	// Step 1: Create organization
	subdomain := "get-e2e-test-" + uuid.New().String()[:8]
	createResp, err := client.CreateOrganization(context.Background(), &pb.CreateOrganizationRequest{
		Name:        "Get E2E Test Org",
		Subdomain:   subdomain,
		OwnerUserId: "770e8400-e29b-41d4-a716-446655440000",
	})
	if err != nil {
		t.Fatalf("Failed to create organization: %v", err)
	}

	orgID := createResp.Organization.Id
	t.Logf("Created organization: %s", orgID)

	// Step 2: Retrieve organization
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

	// Verify fields
	if org.Id != orgID {
		t.Errorf("Expected ID %q, got %q", orgID, org.Id)
	}
	if org.Name != "Get E2E Test Org" {
		t.Errorf("Expected name 'Get E2E Test Org', got %q", org.Name)
	}
	if org.Subdomain != subdomain {
		t.Errorf("Expected subdomain %q, got %q", subdomain, org.Subdomain)
	}
	if org.OwnerUserId == "" {
		t.Error("Owner user ID is empty")
	}
	if org.UpdatedAt == "" {
		t.Error("Updated at timestamp is empty")
	}

	t.Logf("✓ Retrieved organization: ID=%s, Name=%s, Owner=%s", org.Id, org.Name, org.OwnerUserId)
}

// TestCreateAndListOrganizationsE2E tests creating multiple organizations
func TestCreateMultipleOrganizationsE2E(t *testing.T) {
	conn, err := grpc.Dial("localhost:9082", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewOrganizationServiceClient(conn)

	// Create multiple organizations
	orgCount := 3
	orgIDs := make([]string, orgCount)

	for i := 0; i < orgCount; i++ {
		subdomain := "multi-e2e-" + uuid.New().String()[:12]
		resp, err := client.CreateOrganization(context.Background(), &pb.CreateOrganizationRequest{
			Name:        "Multi Org " + uuid.New().String()[:8],
			Subdomain:   subdomain,
			OwnerUserId: "770e8400-e29b-41d4-a716-446655440000",
		})
		if err != nil {
			t.Fatalf("Failed to create organization %d: %v", i+1, err)
		}
		orgIDs[i] = resp.Organization.Id
		t.Logf("Created organization %d: %s", i+1, orgIDs[i])
	}

	// Verify all organizations can be retrieved
	for i, orgID := range orgIDs {
		resp, err := client.GetOrganization(context.Background(), &pb.GetOrganizationRequest{
			Id: orgID,
		})
		if err != nil {
			t.Fatalf("Failed to get organization %d: %v", i+1, err)
		}
		if resp.Organization == nil {
			t.Fatalf("Organization %d response is nil", i+1)
		}
		t.Logf("✓ Retrieved organization %d: %s", i+1, resp.Organization.Name)
	}

	t.Logf("✓ Successfully created and retrieved %d organizations", orgCount)
}

// TestValidationErrorsE2E tests validation error cases
func TestValidationErrorsE2E(t *testing.T) {
	conn, err := grpc.Dial("localhost:9082", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
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
				OwnerUserId: "770e8400-e29b-41d4-a716-446655440000",
			},
			shouldError: false,
		},
		{
			name: "short subdomain (< 3 chars)",
			req: &pb.CreateOrganizationRequest{
				Name:        "Invalid Org",
				Subdomain:   "ab",
				OwnerUserId: "770e8400-e29b-41d4-a716-446655440000",
			},
			shouldError: true,
		},
		{
			name: "subdomain with underscore",
			req: &pb.CreateOrganizationRequest{
				Name:        "Invalid Org",
				Subdomain:   "invalid_org",
				OwnerUserId: "770e8400-e29b-41d4-a716-446655440000",
			},
			shouldError: true,
		},
		{
			name: "subdomain too long",
			req: &pb.CreateOrganizationRequest{
				Name:        "Invalid Org",
				Subdomain:   "this-is-a-very-long-subdomain-that-exceeds-sixty-three-characters-limit",
				OwnerUserId: "770e8400-e29b-41d4-a716-446655440000",
			},
			shouldError: true,
		},
		{
			name: "empty name",
			req: &pb.CreateOrganizationRequest{
				Name:        "",
				Subdomain:   "test-org",
				OwnerUserId: "770e8400-e29b-41d4-a716-446655440000",
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
				} else {
					t.Logf("✓ Got expected error: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if resp == nil || resp.Organization == nil {
					t.Error("Response is nil")
				} else {
					t.Logf("✓ Successfully created organization: %s", resp.Organization.Id)
				}
			}
		})
	}
}

// TestDuplicateSubdomainE2E tests that duplicate subdomains are rejected
func TestDuplicateSubdomainE2E(t *testing.T) {
	conn, err := grpc.Dial("localhost:9082", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewOrganizationServiceClient(conn)

	subdomain := "duplicate-test-" + uuid.New().String()[:8]

	// Create first organization
	resp1, err := client.CreateOrganization(context.Background(), &pb.CreateOrganizationRequest{
		Name:        "First Org",
		Subdomain:   subdomain,
		OwnerUserId: "770e8400-e29b-41d4-a716-446655440000",
	})
	if err != nil {
		t.Fatalf("Failed to create first organization: %v", err)
	}
	t.Logf("✓ Created first organization: %s", resp1.Organization.Id)

	// Try to create second organization with same subdomain
	resp2, err := client.CreateOrganization(context.Background(), &pb.CreateOrganizationRequest{
		Name:        "Second Org",
		Subdomain:   subdomain, // Same subdomain
		OwnerUserId: "770e8400-e29b-41d4-a716-446655440000",
	})

	if err == nil {
		t.Error("Expected error when creating organization with duplicate subdomain, but got none")
		if resp2 != nil {
			t.Logf("Unexpectedly got response: %v", resp2)
		}
	} else {
		t.Logf("✓ Got expected error for duplicate subdomain: %v", err)
	}
}

// TestInvalidOrganizationIDE2E tests getting non-existent organization
func TestInvalidOrganizationIDE2E(t *testing.T) {
	conn, err := grpc.Dial("localhost:9082", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewOrganizationServiceClient(conn)

	tests := []struct {
		name    string
		id      string
		isError bool
	}{
		{
			name:    "valid UUID but non-existent organization",
			id:      "550e8400-e29b-41d4-a716-446655440099",
			isError: true,
		},
		{
			name:    "invalid UUID format",
			id:      "invalid-uuid",
			isError: true,
		},
		{
			name:    "empty ID",
			id:      "",
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.GetOrganization(context.Background(), &pb.GetOrganizationRequest{
				Id: tt.id,
			})

			if tt.isError {
				if err == nil {
					t.Error("Expected error but got nil")
				} else {
					t.Logf("✓ Got expected error: %v", err)
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
