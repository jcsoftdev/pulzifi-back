package deletecurrentuser_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	deletecurrentuser "github.com/jcsoftdev/pulzifi-back/modules/auth/application/delete_current_user"
	authservices "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
	repomocks "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories/mocks"
)

// ── stubs ────────────────────────────────────────────────────────────────────

type stubOrgCascade struct {
	err error
}

func (s *stubOrgCascade) CascadeSolelyOwnedOrgs(_ context.Context, _ uuid.UUID) error {
	return s.err
}

type stubUserCleanup struct {
	called bool
	err    error
}

func (s *stubUserCleanup) PruneMemberships(_ context.Context, _ uuid.UUID) error {
	s.called = true
	return s.err
}

// ── existing tests (unchanged behavior) ─────────────────────────────────────

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

			// nil eventBus and nil orgCascade: backward-compat path
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

// ── PR3 extension tests ───────────────────────────────────────────────────────

func TestDeleteCurrentUserHandler_WithOrgCascade(t *testing.T) {
	userID := uuid.New()

	t.Run("owner path: cascade nil -> user deleted", func(t *testing.T) {
		deleteCalled := false
		repo := &repomocks.MockUserRepository{
			DeleteFn: func(_ context.Context, _ uuid.UUID) error {
				deleteCalled = true
				return nil
			},
		}
		cascade := &stubOrgCascade{err: nil}
		cleanup := &stubUserCleanup{}

		h := deletecurrentuser.NewHandler(repo, nil).
			WithOrgCascade(cascade).
			WithUserCleanup(cleanup)

		if err := h.Handle(context.Background(), &deletecurrentuser.Request{UserID: userID}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !deleteCalled {
			t.Fatal("expected user repo Delete to be called")
		}
		if !cleanup.called {
			t.Fatal("expected PruneMemberships to be called")
		}
	})

	t.Run("owner path: cascade returns ErrBillingActive -> user NOT deleted", func(t *testing.T) {
		deleteCalled := false
		repo := &repomocks.MockUserRepository{
			DeleteFn: func(_ context.Context, _ uuid.UUID) error {
				deleteCalled = true
				return nil
			},
		}
		cascade := &stubOrgCascade{err: authservices.ErrBillingActive}

		h := deletecurrentuser.NewHandler(repo, nil).
			WithOrgCascade(cascade)

		err := h.Handle(context.Background(), &deletecurrentuser.Request{UserID: userID})
		if !errors.Is(err, authservices.ErrBillingActive) {
			t.Fatalf("expected ErrBillingActive, got %v", err)
		}
		if deleteCalled {
			t.Fatal("user repo Delete must NOT be called on billing abort")
		}
	})

	t.Run("member path: cascade nil, prune called, user deleted", func(t *testing.T) {
		deleteCalled := false
		repo := &repomocks.MockUserRepository{
			DeleteFn: func(_ context.Context, _ uuid.UUID) error {
				deleteCalled = true
				return nil
			},
		}
		cascade := &stubOrgCascade{err: nil}
		cleanup := &stubUserCleanup{}

		h := deletecurrentuser.NewHandler(repo, nil).
			WithOrgCascade(cascade).
			WithUserCleanup(cleanup)

		if err := h.Handle(context.Background(), &deletecurrentuser.Request{UserID: userID}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !cleanup.called {
			t.Fatal("expected PruneMemberships to be called on member path")
		}
		if !deleteCalled {
			t.Fatal("expected user repo Delete to be called")
		}
	})

	t.Run("nil orgCascade -> backward compat, user deleted", func(t *testing.T) {
		deleteCalled := false
		repo := &repomocks.MockUserRepository{
			DeleteFn: func(_ context.Context, _ uuid.UUID) error {
				deleteCalled = true
				return nil
			},
		}

		h := deletecurrentuser.NewHandler(repo, nil) // no WithOrgCascade

		if err := h.Handle(context.Background(), &deletecurrentuser.Request{UserID: userID}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !deleteCalled {
			t.Fatal("expected user repo Delete to be called")
		}
	})
}
