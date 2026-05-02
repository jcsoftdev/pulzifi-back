package services

import (
	"encoding/json"
	"fmt"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
)

type PayloadBuilder struct{}

func NewPayloadBuilder() *PayloadBuilder { return &PayloadBuilder{} }

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
		p.Title = "Page change detected"
		p.PageURL, _ = raw["page_url"].(string)
		p.Body = fmt.Sprintf("A change was detected on %s", p.PageURL)
		p.Severity = "warning"
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
