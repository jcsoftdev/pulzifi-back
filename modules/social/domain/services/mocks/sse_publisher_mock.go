package mocks

import (
	"sync"

	"github.com/google/uuid"
)

// MockSocialSSEPublisher is a hand-rolled mock for services.SocialSSEPublisher.
// Publish is safe for concurrent use because run_check may call it from a goroutine.
type MockSocialSSEPublisher struct {
	mu             sync.Mutex
	PublishCalls   int
	PublishedPairs []PublishedPair
}

type PublishedPair struct {
	ProfileID uuid.UUID
	Payload   []byte
}

func (m *MockSocialSSEPublisher) Publish(profileID uuid.UUID, payload []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.PublishCalls++
	m.PublishedPairs = append(m.PublishedPairs, PublishedPair{ProfileID: profileID, Payload: payload})
}

// Calls returns the publish count under lock (use in tests after goroutines settle).
func (m *MockSocialSSEPublisher) Calls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.PublishCalls
}
