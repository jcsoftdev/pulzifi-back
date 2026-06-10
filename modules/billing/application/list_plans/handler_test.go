package listplans_test

import (
	"context"
	"errors"
	"testing"

	listplans "github.com/jcsoftdev/pulzifi-back/modules/billing/application/list_plans"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
)

type fakePlanCatalogRepo struct {
	plans []*entities.CatalogPlan
	err   error
}

func (r *fakePlanCatalogRepo) ListForCatalog(_ context.Context) ([]*entities.CatalogPlan, error) {
	return r.plans, r.err
}

func ptr(n int) *int { return &n }

func TestListPlansHandler_ReturnsAllActivePlans(t *testing.T) {
	repo := &fakePlanCatalogRepo{
		plans: []*entities.CatalogPlan{
			{
				ID: "uuid-starter", Code: "starter", Name: "Starter",
				Description: "desc", PriceMonthlyCents: 2700, PriceYearlyCents: 26400,
				Currency: "usd", ChecksAllowedMonthly: ptr(500), StoragePeriodDays: 7,
				AIInsightsAllowedMonthly: ptr(50), MaxPages: ptr(10), MaxWorkspaces: ptr(3),
				IsPopular: false,
			},
			{
				ID: "uuid-pro", Code: "pro", Name: "Pro",
				Description: "desc", PriceMonthlyCents: 5400, PriceYearlyCents: 55200,
				Currency: "usd", ChecksAllowedMonthly: ptr(10000), StoragePeriodDays: 90,
				AIInsightsAllowedMonthly: ptr(250), MaxPages: ptr(100), MaxWorkspaces: nil,
				IsPopular: true,
			},
		},
	}

	h := listplans.NewHandler(repo)
	resp, err := h.Handle(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Plans) != 2 {
		t.Fatalf("want 2 plans, got %d", len(resp.Plans))
	}
	pro := resp.Plans[1]
	if pro.Code != "pro" {
		t.Errorf("want code=pro, got %q", pro.Code)
	}
	if !pro.IsPopular {
		t.Error("want pro.IsPopular=true")
	}
	if pro.MaxWorkspaces != nil {
		t.Errorf("want MaxWorkspaces=nil (unlimited), got %v", pro.MaxWorkspaces)
	}
	if pro.PriceYearlyCents != 55200 {
		t.Errorf("want PriceYearlyCents=55200, got %d", pro.PriceYearlyCents)
	}
}

func TestListPlansHandler_PropagatesRepoError(t *testing.T) {
	repo := &fakePlanCatalogRepo{err: errors.New("db down")}
	h := listplans.NewHandler(repo)
	_, err := h.Handle(context.Background())
	if err == nil {
		t.Fatal("expected error from repo, got nil")
	}
}

func TestListPlansHandler_EmptyRepoReturnsEmptySlice(t *testing.T) {
	repo := &fakePlanCatalogRepo{plans: nil}
	h := listplans.NewHandler(repo)
	resp, err := h.Handle(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Plans == nil {
		t.Error("want empty slice, not nil (JSON must encode as [])")
	}
}
