// Package outbox provides a durable transactional outbox implementation of
// eventbus.MessageBus backed by PostgreSQL. Events are written to
// public.outbox_events and dispatched by the Relay polling loop.
package outbox

import "time"

// Config holds tuning knobs for the outbox bus and its relay.
type Config struct {
	// PollInterval is how often the relay wakes up to claim pending rows.
	PollInterval time.Duration

	// BatchSize is the number of rows claimed per relay tick.
	BatchSize int

	// MaxAttempts is the number of delivery attempts before a row is moved to
	// dead-letter status (status='dead').
	MaxAttempts int

	// StaleThreshold is how long a row may remain in status='publishing' before
	// the relay considers the previous publisher crashed and resets the row to
	// status='pending' without incrementing attempts.
	StaleThreshold time.Duration

	// BackoffBase is the minimum retry delay applied after the first failure.
	// Subsequent delays grow exponentially: min(base*2^attempts, BackoffCap).
	BackoffBase time.Duration

	// BackoffCap is the maximum retry delay.
	BackoffCap time.Duration
}

// DefaultConfig returns production-sensible defaults.
func DefaultConfig() Config {
	return Config{
		PollInterval:   1 * time.Second,
		BatchSize:      100,
		MaxAttempts:    8,
		StaleThreshold: 30 * time.Second,
		BackoffBase:    5 * time.Second,
		BackoffCap:     1 * time.Hour,
	}
}
