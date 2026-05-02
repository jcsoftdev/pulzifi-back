package eventbus

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	TopicChangeDetected = "change.detected"
	TopicAlertCreated   = "alert.created"
	TopicInsightReady   = "insight.ready"
	TopicCheckFailed    = "check.failed"
	TopicReportWeekly   = "report.weekly"
)

type DomainEvent struct {
	ID          uuid.UUID       `json:"id"`
	Type        string          `json:"type"`
	OccurredAt  time.Time       `json:"occurred_at"`
	OrgID       uuid.UUID       `json:"org_id"`
	WorkspaceID *uuid.UUID      `json:"workspace_id,omitempty"`
	PageID      *uuid.UUID      `json:"page_id,omitempty"`
	Tenant      string          `json:"tenant"`
	Data        json.RawMessage `json:"data"`
}

func PublishDomainEvent(bus MessageBus, ev DomainEvent) error {
	if ev.ID == uuid.Nil {
		ev.ID = uuid.New()
	}
	if ev.OccurredAt.IsZero() {
		ev.OccurredAt = time.Now().UTC()
	}
	payload, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	return bus.Publish(ev.Type, ev.ID.String(), payload)
}

type DomainEventHandler func(ev DomainEvent)

func SubscribeDomainEvent(bus MessageBus, topic string, h DomainEventHandler) error {
	return bus.Subscribe(topic, func(_ string, payload []byte) {
		var ev DomainEvent
		if err := json.Unmarshal(payload, &ev); err != nil {
			return
		}
		h(ev)
	})
}
