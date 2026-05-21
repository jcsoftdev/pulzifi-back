package deletecurrentuser_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	deletecurrentuser "github.com/jcsoftdev/pulzifi-back/modules/auth/application/delete_current_user"
	repomocks "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories/mocks"
)

func TestDeleteCurrentUserHandler_Handle(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name    string
		setup   func(*repomocks.MockUserRepository)
		wantErr bool
	}{
		{
			name: "successfully deletes user",
			setup: func(m *repomocks.MockUserRepository) {
				// DeleteErr is nil by default
			},
		},
		{
			name: "propagates repo error",
			setup: func(m *repomocks.MockUserRepository) {
				m.DeleteErr = errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &repomocks.MockUserRepository{}
			tt.setup(userRepo)

			// nil eventBus: no event published, but handler still runs
			h := deletecurrentuser.NewHandler(userRepo, nil)
			err := h.Handle(context.Background(), &deletecurrentuser.Request{UserID: userID})

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
