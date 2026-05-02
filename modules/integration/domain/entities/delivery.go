package entities

import (
	"time"

	"github.com/google/uuid"
)

type DeliveryStatus string

const (
	DeliveryPending   DeliveryStatus = "pending"
	DeliveryInFlight  DeliveryStatus = "in_flight"
	DeliveryDelivered DeliveryStatus = "delivered"
	DeliveryFailed    DeliveryStatus = "failed"
	DeliveryDead      DeliveryStatus = "dead"
)

type Delivery struct {
	ID             uuid.UUID
	DestinationID  uuid.UUID
	EventType      string
	EventPayload   map[string]any
	Status         DeliveryStatus
	Attempts       int
	LastAttemptAt  *time.Time
	NextAttemptAt  *time.Time
	ResponseCode   *int
	ResponseBody   *string
	ErrorMessage   *string
	DeliveredAt    *time.Time
	CreatedAt      time.Time
	AttemptHistory []Attempt
}

// Attempt records a single delivery attempt (appended by the worker on each try).
type Attempt struct {
	At    time.Time `json:"at"`
	Code  *int      `json:"code,omitempty"`
	Error string    `json:"error,omitempty"`
}
