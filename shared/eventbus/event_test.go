package eventbus_test

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
)

func TestPublishSubscribeDomainEvent(t *testing.T) {
	bus := eventbus.GetInstance()
	var got eventbus.DomainEvent
	var wg sync.WaitGroup
	wg.Add(1)
	eventbus.SubscribeDomainEvent(bus, "test.topic", func(ev eventbus.DomainEvent) {
		got = ev
		wg.Done()
	})
	raw, _ := json.Marshal(map[string]string{"k": "v"})
	eventbus.PublishDomainEvent(bus, eventbus.DomainEvent{
		Type: "test.topic", OrgID: uuid.New(), Tenant: "t1", Data: raw,
	})
	wg.Wait()
	if got.Type != "test.topic" {
		t.Fatal("type mismatch")
	}
}
