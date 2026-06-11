package services

import "github.com/google/uuid"

// SocialSSEPublisher pushes a serialized change/alert payload to real-time SSE
// subscribers keyed by profile ID. Implemented by the shared pubsub broker.
type SocialSSEPublisher interface {
	Publish(profileID uuid.UUID, payload []byte)
}
