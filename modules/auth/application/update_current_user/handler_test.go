package updatecurrentuser_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	updatecurrentuser "github.com/jcsoftdev/pulzifi-back/modules/auth/application/update_current_user"
	getcurrentuser "github.com/jcsoftdev/pulzifi-back/modules/auth/application/get_current_user"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
	repomocks "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories/mocks"
)

func TestUpdateCurrentUserHandler_Handle(t *testing.T) {
	userID := uuid.New()
	existingUser := &entities.User{ID: userID, FirstName: "Alice", LastName: "Smith", Email: "alice@example.com"}

	tests := []struct {
		name     string
		req      *updatecurrentuser.Request
		setup    func(*repomocks.MockUserRepository)
		wantErr  bool
		wantName string
	}{
		{
			name: "updates first name",
			req:  &updatecurrentuser.Request{UserID: userID, FirstName: "Bob", Roles: []string{"USER"}},
			setup: func(m *repomocks.MockUserRepository) {
				m.GetByIDUser = existingUser
				m.GetUserFirstOrganizationResult = nil
			},
			wantName: "Bob Smith",
		},
		{
			name: "user not found returns error",
			req:  &updatecurrentuser.Request{UserID: uuid.New()},
			setup: func(m *repomocks.MockUserRepository) {
				m.GetByIDUser = nil
				m.GetByIDErr = errors.New("not found")
			},
			wantErr: true,
		},
		{
			name: "update error propagates",
			req:  &updatecurrentuser.Request{UserID: userID, FirstName: "Bob"},
			setup: func(m *repomocks.MockUserRepository) {
				m.GetByIDUser = existingUser
				m.UpdateErr = errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &repomocks.MockUserRepository{}
			tt.setup(userRepo)

			// Build a real get_current_user handler with the same mock repo
			getCurrentUserHandler := getcurrentuser.NewHandler(userRepo, nil)

			h := updatecurrentuser.NewHandler(userRepo, getCurrentUserHandler)
			resp, err := h.Handle(context.Background(), tt.req)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantName != "" && resp.Name != tt.wantName {
				t.Errorf("name: want %q, got %q", tt.wantName, resp.Name)
			}
		})
	}
}
