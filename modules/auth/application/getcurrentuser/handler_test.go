package getcurrentuser

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
	repomocks "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories/mocks"
)

func TestGetCurrentUserHandler_Handle(t *testing.T) {
	userID := uuid.New()
	tenant := "acme"
	testUser := &entities.User{
		ID:        userID,
		FirstName: "Alice",
		LastName:  "Smith",
		Email:     "alice@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tests := []struct {
		name        string
		userID      uuid.UUID
		roles       []string
		setupMocks  func(u *repomocks.MockUserRepository)
		wantErr     bool
		checkResult func(t *testing.T, resp *Response)
	}{
		{
			name:   "authenticated user resolved from repo",
			userID: userID,
			roles:  []string{"ADMIN"},
			setupMocks: func(u *repomocks.MockUserRepository) {
				u.GetByIDUser = testUser
				u.GetUserFirstOrganizationResult = &tenant
			},
			wantErr: false,
			checkResult: func(t *testing.T, resp *Response) {
				t.Helper()
				if resp.ID != userID.String() {
					t.Errorf("user ID: want %q, got %q", userID.String(), resp.ID)
				}
				if resp.Email != "alice@example.com" {
					t.Errorf("email: want %q, got %q", "alice@example.com", resp.Email)
				}
				if resp.Role != "ADMIN" {
					t.Errorf("role: want ADMIN, got %q", resp.Role)
				}
				if resp.Tenant == nil || *resp.Tenant != "acme" {
					t.Errorf("tenant: want %q, got %v", "acme", resp.Tenant)
				}
			},
		},
		{
			name:   "user not found in repo returns nil response without error",
			userID: userID,
			roles:  []string{"USER"},
			setupMocks: func(u *repomocks.MockUserRepository) {
				u.GetByIDUser = nil // no user returned
				u.GetByIDErr = nil
			},
			wantErr: false,
			checkResult: func(t *testing.T, resp *Response) {
				t.Helper()
				if resp != nil {
					t.Errorf("expected nil response for missing user, got %+v", resp)
				}
			},
		},
		{
			name:   "repo error propagates",
			userID: userID,
			roles:  []string{"USER"},
			setupMocks: func(u *repomocks.MockUserRepository) {
				u.GetByIDErr = errors.New("db error")
			},
			wantErr: true,
		},
		{
			name:   "GetByID is called with correct user ID",
			userID: userID,
			roles:  []string{"USER"},
			setupMocks: func(u *repomocks.MockUserRepository) {
				u.GetByIDFn = func(ctx context.Context, id uuid.UUID) (*entities.User, error) {
					if id != userID {
						return nil, errors.New("wrong user ID passed")
					}
					return testUser, nil
				}
				u.GetUserFirstOrganizationResult = &tenant
			},
			wantErr: false,
			checkResult: func(t *testing.T, resp *Response) {
				t.Helper()
				if resp == nil {
					t.Fatal("expected non-nil response")
				}
				if resp.ID != userID.String() {
					t.Errorf("ID mismatch: want %q, got %q", userID.String(), resp.ID)
				}
			},
		},
		{
			name:   "highest role selected when multiple roles provided",
			userID: userID,
			roles:  []string{"USER", "ADMIN", "VIEWER"},
			setupMocks: func(u *repomocks.MockUserRepository) {
				u.GetByIDUser = testUser
			},
			wantErr: false,
			checkResult: func(t *testing.T, resp *Response) {
				t.Helper()
				if resp.Role != "ADMIN" {
					t.Errorf("expected highest role ADMIN, got %q", resp.Role)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &repomocks.MockUserRepository{}
			if tt.setupMocks != nil {
				tt.setupMocks(userRepo)
			}

			h := NewHandler(userRepo, nil) // nil orgLookup — Organization field omitted
			resp, err := h.Handle(context.Background(), tt.userID, tt.roles)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.checkResult != nil {
				tt.checkResult(t, resp)
			}
		})
	}
}
