package entities_test

import (
	"testing"

	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
)

func TestCatalogPlan_IsUnlimitedChecks(t *testing.T) {
	p := &entities.CatalogPlan{ChecksAllowedMonthly: nil}
	if !p.IsUnlimitedChecks() {
		t.Fatal("want unlimited checks when nil")
	}
	n := 500
	p2 := &entities.CatalogPlan{ChecksAllowedMonthly: &n}
	if p2.IsUnlimitedChecks() {
		t.Fatal("want capped when non-nil")
	}
}
