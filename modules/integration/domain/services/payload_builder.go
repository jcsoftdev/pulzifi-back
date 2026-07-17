package services

import (
	"encoding/json"
	"fmt"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
)

// PayloadBuilder turns a raw DomainEvent into a provider-agnostic
// NotificationPayload. appDomain is the apex application host (e.g. "pulzifi.com")
// used to build tenant-subdomain dashboard deep links; empty disables them.
type PayloadBuilder struct {
	appDomain string
}

func NewPayloadBuilder(appDomain string) *PayloadBuilder {
	return &PayloadBuilder{appDomain: appDomain}
}

func (b *PayloadBuilder) Build(ev eventbus.DomainEvent) (*entities.NotificationPayload, error) {
	var raw map[string]any
	if len(ev.Data) > 0 {
		if err := json.Unmarshal(ev.Data, &raw); err != nil {
			return nil, err
		}
	}
	p := &entities.NotificationPayload{
		EventType: ev.Type,
		Severity:  "info",
		Raw:       raw,
	}
	switch ev.Type {
	case eventbus.TopicChangeDetected:
		b.buildChangeDetected(p, raw)
	case eventbus.TopicAlertCreated:
		p.Title, _ = raw["title"].(string)
		p.Body, _ = raw["message"].(string)
		p.Severity = "critical"
	default:
		p.Title = ev.Type
		p.Body = "Pulzifi event"
	}
	return p, nil
}

// buildChangeDetected enriches the payload for a page change event, using the
// summary, page name, change type, diff image and a dashboard deep link when
// present. Falls back to the raw page URL when richer data is missing.
func (b *PayloadBuilder) buildChangeDetected(p *entities.NotificationPayload, raw map[string]any) {
	p.Severity = "warning"
	p.PageURL, _ = raw["page_url"].(string)
	p.PageTitle, _ = raw["page_title"].(string)
	p.ChangeType, _ = raw["change_type"].(string)
	p.DiffSummary, _ = raw["diff_summary"].(string)
	p.DiffImageURL, _ = raw["diff_image_url"].(string)
	p.ChangedAt, _ = raw["changed_at"].(string)
	p.DashboardURL = b.changesLink(raw)

	name := p.PageTitle
	if name == "" {
		name = p.PageURL
	}
	p.Title = fmt.Sprintf("Change detected on %s", name)

	p.Body = p.DiffSummary
	if p.Body == "" {
		p.Body = fmt.Sprintf("A change was detected on %s", p.PageURL)
	}
}

// changesLink builds the tenant-subdomain deep link to a page's changes view:
//
//	https://{tenant}.{appDomain}/workspaces/{workspaceID}/pages/{pageID}/changes
//
// Returns "" when appDomain is unset or any identifier is missing, so callers
// can fall back to the raw page URL.
func (b *PayloadBuilder) changesLink(raw map[string]any) string {
	tenant, _ := raw["tenant"].(string)
	workspaceID, _ := raw["workspace_id"].(string)
	pageID, _ := raw["page_id"].(string)
	if b.appDomain == "" || tenant == "" || workspaceID == "" || pageID == "" {
		return ""
	}
	return fmt.Sprintf("https://%s.%s/workspaces/%s/pages/%s/changes",
		tenant, b.appDomain, workspaceID, pageID)
}
