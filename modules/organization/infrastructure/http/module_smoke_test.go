package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	orghttp "github.com/jcsoftdev/pulzifi-back/modules/organization/infrastructure/http"
	orgentities "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/entities"
	orgrepos "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/repositories"
)

// stubOrgRepo is a minimal in-memory implementation of OrganizationRepository.
type stubOrgRepo struct{}

func (stubOrgRepo) Create(_ context.Context, _ *orgentities.Organization) error { return nil }
func (stubOrgRepo) GetByID(_ context.Context, _ uuid.UUID) (*orgentities.Organization, error) {
	return nil, nil
}
func (stubOrgRepo) GetBySubdomain(_ context.Context, _ string) (*orgentities.Organization, error) {
	return nil, nil
}
func (stubOrgRepo) List(_ context.Context, _ uuid.UUID, _, _ int) ([]*orgentities.Organization, error) {
	return nil, nil
}
func (stubOrgRepo) Update(_ context.Context, _ *orgentities.Organization) error { return nil }
func (stubOrgRepo) Delete(_ context.Context, _ uuid.UUID) error                 { return nil }
func (stubOrgRepo) CountBySubdomain(_ context.Context, _ string) (int, error)   { return 0, nil }

var _ orgrepos.OrganizationRepository = stubOrgRepo{}

// TestOrganizationModuleSmoke verifies that GET /organizations is registered (not 404).
func TestOrganizationModuleSmoke(t *testing.T) {
	mod := orghttp.NewModule(stubOrgRepo{})

	r := chi.NewRouter()
	mod.RegisterHTTPRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/organizations", http.NoBody)
	rr := httptest.NewRecorder()

	panicked := false
	func() {
		defer func() {
			if rec := recover(); rec != nil {
				panicked = true
			}
		}()
		r.ServeHTTP(rr, req)
	}()

	if !panicked && rr.Code == http.StatusNotFound {
		t.Errorf("GET /organizations returned 404 — route is not registered")
	}
}
