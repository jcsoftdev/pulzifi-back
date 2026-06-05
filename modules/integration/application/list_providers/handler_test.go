package listproviders_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	listproviders "github.com/jcsoftdev/pulzifi-back/modules/integration/application/list_providers"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/services"
)

type fakeRegistry struct{ present map[string]bool }

func (f fakeRegistry) Get(t string) (services.ProviderClient, bool) {
	return nil, f.present[t]
}

type fakeFlags struct{ on map[string]bool }

func (f fakeFlags) IsOn(_ context.Context, _ uuid.UUID, key string) (bool, error) {
	return f.on[key], nil
}

type fakeRepo struct{ integrations []*entities.Integration }

func (f fakeRepo) ListByOrg(_ context.Context, _ uuid.UUID) ([]*entities.Integration, error) {
	return f.integrations, nil
}

func find(states []listproviders.ProviderState, key string) (listproviders.ProviderState, bool) {
	for _, s := range states {
		if s.Key == key {
			return s, true
		}
	}
	return listproviders.ProviderState{}, false
}

func TestHandle_ComputesEnabledConnectedAuthType(t *testing.T) {
	reg := fakeRegistry{present: map[string]bool{
		"slack": true, "email": true, "twilio": true, "sheets": true,
	}}
	flags := fakeFlags{on: map[string]bool{"integrations.sheets": true}}
	repo := fakeRepo{integrations: []*entities.Integration{
		{ServiceType: "slack", Status: entities.IntegrationActive},
		{ServiceType: "sheets", Status: entities.IntegrationDisconnected},
	}}

	h := listproviders.NewHandler(reg, flags, repo)
	resp, err := h.Handle(context.Background(), listproviders.Request{OrgID: uuid.New()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	slack, _ := find(resp.Providers, "slack")
	if !slack.Enabled || !slack.Connected || slack.AuthType != "oauth" {
		t.Errorf("slack = %+v, want enabled+connected+oauth", slack)
	}
	email, _ := find(resp.Providers, "email")
	if !email.Enabled || !email.Connected || email.AuthType != "none" {
		t.Errorf("email = %+v, want enabled+connected(none always-on)+none", email)
	}
	gmail, _ := find(resp.Providers, "gmail")
	if gmail.Enabled || gmail.Connected {
		t.Errorf("gmail = %+v, want disabled+disconnected (not registered)", gmail)
	}
	twilio, _ := find(resp.Providers, "twilio")
	if twilio.Enabled || twilio.AuthType != "byo" {
		t.Errorf("twilio = %+v, want disabled (flag off) + byo", twilio)
	}
	sheets, _ := find(resp.Providers, "sheets")
	if !sheets.Enabled || sheets.Connected {
		t.Errorf("sheets = %+v, want enabled (registered+flag) but NOT connected (integration disconnected)", sheets)
	}
}
