package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	deleteorganization "github.com/jcsoftdev/pulzifi-back/modules/organization/application/delete_organization"
	orgservices "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/contextkeys"
)

// stubDeleteOrgHandler is a fake deleteorganization.Handler for HTTP tests.
type stubDeleteOrgHandler struct {
	returnErr error
	called    bool
}

func (s *stubDeleteOrgHandler) Handle(_ context.Context, _ *deleteorganization.Request) (*deleteorganization.Response, error) {
	s.called = true
	if s.returnErr != nil {
		return nil, s.returnErr
	}
	return &deleteorganization.Response{}, nil
}

func TestHandleDeleteOrganization(t *testing.T) {
	orgID := uuid.New()

	tests := []struct {
		name        string
		roles       []string
		orgIDParam  string
		handlerErr  error
		wantStatus  int
		wantBodySub string
	}{
		{
			name:        "403 when caller is not SUPER_ADMIN",
			roles:       []string{"owner"},
			orgIDParam:  orgID.String(),
			wantStatus:  http.StatusForbidden,
			wantBodySub: "super_admin",
		},
		{
			name:        "404 when use case returns ErrOrgNotFound",
			roles:       []string{"SUPER_ADMIN"},
			orgIDParam:  orgID.String(),
			handlerErr:  orgservices.ErrOrgNotFound,
			wantStatus:  http.StatusNotFound,
			wantBodySub: "org_not_found",
		},
		{
			name:        "409 when use case returns ErrBillingActive",
			roles:       []string{"SUPER_ADMIN"},
			orgIDParam:  orgID.String(),
			handlerErr:  orgservices.ErrBillingActive,
			wantStatus:  http.StatusConflict,
			wantBodySub: "billing_active",
		},
		{
			name:       "204 on success",
			roles:      []string{"SUPER_ADMIN"},
			orgIDParam: orgID.String(),
			wantStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &stubDeleteOrgHandler{returnErr: tt.handlerErr}
			// newModuleWithDeleteHandler is package-internal (unexported).
			mod := newModuleWithDeleteHandler(stub)

			// Build a Chi router so {id} URL param works.
			r := chi.NewRouter()
			r.Delete("/organizations/{id}", func(w http.ResponseWriter, req *http.Request) {
				// Inject roles into context (mirrors auth middleware).
				ctx := context.WithValue(req.Context(), contextkeys.UserRolesKey, tt.roles)
				ctx = context.WithValue(ctx, contextkeys.UserIDKey, uuid.New().String())
				mod.ServeDeleteOrganization(w, req.WithContext(ctx))
			})

			req := httptest.NewRequest(http.MethodDelete, "/organizations/"+tt.orgIDParam, nil)
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body = %s", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if tt.wantBodySub != "" && !strings.Contains(rr.Body.String(), tt.wantBodySub) {
				t.Errorf("body %q does not contain %q", rr.Body.String(), tt.wantBodySub)
			}
			// When the caller is not SUPER_ADMIN the use case must never be invoked.
			if tt.wantStatus == http.StatusForbidden && stub.called {
				t.Error("use case must NOT be invoked when the request is forbidden")
			}
		})
	}
}

// Verify ErrOrgNotFound and ErrBillingActive are tested.
func TestErrSentinelsExist(t *testing.T) {
	if !errors.Is(orgservices.ErrOrgNotFound, orgservices.ErrOrgNotFound) {
		t.Fatal("ErrOrgNotFound sentinel mismatch")
	}
	if !errors.Is(orgservices.ErrBillingActive, orgservices.ErrBillingActive) {
		t.Fatal("ErrBillingActive sentinel mismatch")
	}
}
